// NVM Express driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package nvme

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// Submission Queue Entry layout
// (NVM Express Base Specification 2.0 - 4.2 Submission Queue Entry).
const (
	sqeSize = 64

	SQE_CDW0  = 0x00
	CDW0_CID  = 16
	CDW0_PSDT = 14
	CDW0_FUSE = 8
	CDW0_OPC  = 0

	SQE_NSID  = 0x04
	SQE_CDW2  = 0x08
	SQE_MPTR  = 0x10
	SQE_PRP1  = 0x18
	SQE_PRP2  = 0x20
	SQE_CDW10 = 0x28
)

// Completion Queue Entry layout
// (NVM Express Base Specification 2.0 - 4.2.1 Completion Queue Entry).
const (
	cqeSize = 16

	CQE_DW0 = 0x00

	CQE_DW3 = 0x0c
	DW3_DNR = 31
	DW3_SCT = 25
	DW3_SC  = 17
	DW3_SF  = 17
	DW3_P   = 16
	DW3_CID = 0
)

// StatusError represents an NVMe command completion error.
type StatusError struct {
	// Opcode is the failed command opcode.
	Opcode uint8
	// Type is the Status Code Type (SCT).
	Type uint8
	// Code is the Status Code (SC).
	Code uint8
	// DoNotRetry is the DNR bit.
	DoNotRetry bool
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("command %#02x failed (sct:%#x sc:%#02x dnr:%v)", e.Opcode, e.Type, e.Code, e.DoNotRetry)
}

// command represents a submission queue entry.
type command struct {
	opcode uint8
	nsid   uint32
	cdw    [6]uint32 // CDW10-CDW15

	addr   uint
	length int
}

// queue represents a submission/completion queue pair.
type queue struct {
	sync.Mutex

	hw    *NVMe
	id    uint16
	depth int

	sqAddr uint
	sq     []byte
	cqAddr uint
	cq     []byte

	sqDoorbell uint64
	cqDoorbell uint64

	sqTail uint16
	cqHead uint16
	phase  uint32

	// command identifiers
	free  []uint16
	owner []int
	op    []uint8

	// per command PRP list pages (I/O queues only)
	prpAddr []uint
	prp     [][]byte

	// completion interrupt wake up channel
	irq chan struct{}
	// lost interrupt safety net timer
	timer *time.Timer

	// set on unrecoverable errors (e.g. timeouts)
	failed error
}

func (hw *NVMe) newQueue(id uint16, depth int) (q *queue, err error) {
	if depth < 2 {
		return nil, errors.New("invalid queue depth")
	}

	irqs := hw.irqs.Load()

	if irqs == nil || int(id) >= len(*irqs) {
		return nil, errors.New("no interrupt vector available for queue")
	}

	q = &queue{
		irq:        (*irqs)[id],
		timer:      time.NewTimer(recheck),
		hw:         hw,
		id:         id,
		depth:      depth,
		phase:      1,
		sqDoorbell: hw.doorbell(id, false),
		cqDoorbell: hw.doorbell(id, true),
		owner:      make([]int, depth),
		op:         make([]uint8, depth),
	}

	q.sqAddr, q.sq = hw.Region.Reserve(depth*sqeSize, pageSize)
	q.cqAddr, q.cq = hw.Region.Reserve(depth*cqeSize, pageSize)

	clear(q.sq)
	clear(q.cq)

	for cid := depth - 2; cid >= 0; cid-- {
		q.free = append(q.free, uint16(cid))
	}

	if id == 0 {
		return
	}

	q.prpAddr = make([]uint, depth)
	q.prp = make([][]byte, depth)

	for cid := range depth {
		q.prpAddr[cid], q.prp[cid] = hw.Region.Reserve(pageSize, pageSize)
	}

	return
}

// discard marks the queues of a previous initialization as unusable and
// detaches them, returning them for release once the controller is disabled.
func (hw *NVMe) discard() (stale []*queue) {
	stale = hw.io

	if hw.admin != nil {
		stale = append(stale, hw.admin)
	}

	for _, q := range stale {
		q.Lock()
		q.failed = errors.New("controller re-initialized")
		q.timer.Stop()
		q.Unlock()
	}

	hw.admin = nil
	hw.io = nil
	hw.Info = nil
	hw.Namespaces = nil

	return
}

func (q *queue) release() {
	r := q.hw.Region

	r.Release(q.sqAddr)
	r.Release(q.cqAddr)

	for _, addr := range q.prpAddr {
		r.Release(addr)
	}
}

// prps returns the PRP entries describing a physically contiguous buffer.
func (q *queue) prps(cid uint16, addr uint, length int) (prp1, prp2 uint64, err error) {
	if length == 0 {
		// no data transfer, addr (if any) is a queue base address
		return uint64(addr), 0, nil
	}

	if addr&3 != 0 {
		return 0, 0, errors.New("DMA buffer must be 4-byte aligned")
	}

	prp1 = uint64(addr)
	off := addr % pageSize
	first := addr - off
	pages := int((off + uint(length) + pageSize - 1) / pageSize)

	switch {
	case pages == 1:
	case pages == 2:
		prp2 = uint64(first + pageSize)
	case q.prp == nil || pages-1 > maxTransferPages:
		return 0, 0, errors.New("transfer too large for a single command")
	default:
		list := q.prp[cid]

		for i := 1; i < pages; i++ {
			addr := uint64(first + uint(i)*pageSize)
			binary.LittleEndian.PutUint64(list[(i-1)*8:], addr)
		}

		prp2 = uint64(q.prpAddr[cid])
	}

	return
}

// push writes a command in the submission queue, the doorbell is not rung.
func (q *queue) push(cmd *command, owner int) (err error) {
	n := len(q.free)

	if n == 0 {
		return errors.New("queue full")
	}

	cid := q.free[n-1]
	prp1, prp2, err := q.prps(cid, cmd.addr, cmd.length)

	if err != nil {
		return
	}

	q.free = q.free[:n-1]
	q.owner[cid] = owner
	q.op[cid] = cmd.opcode

	sqe := q.sq[int(q.sqTail)*sqeSize : int(q.sqTail+1)*sqeSize]

	var cdw0 uint32
	bits.SetN(&cdw0, CDW0_CID, 0xffff, uint32(cid))
	bits.SetN(&cdw0, CDW0_OPC, 0xff, uint32(cmd.opcode))

	binary.LittleEndian.PutUint32(sqe[SQE_CDW0:], cdw0)
	binary.LittleEndian.PutUint32(sqe[SQE_NSID:], cmd.nsid)
	binary.LittleEndian.PutUint64(sqe[SQE_CDW2:], 0)
	binary.LittleEndian.PutUint64(sqe[SQE_MPTR:], 0)
	binary.LittleEndian.PutUint64(sqe[SQE_PRP1:], prp1)
	binary.LittleEndian.PutUint64(sqe[SQE_PRP2:], prp2)

	for i, v := range cmd.cdw {
		binary.LittleEndian.PutUint32(sqe[SQE_CDW10+i*4:], v)
	}

	if q.sqTail++; int(q.sqTail) == q.depth {
		q.sqTail = 0
	}

	return
}

// ring notifies the controller of all pushed commands.
func (q *queue) ring() {
	reg.Write64As32(q.sqDoorbell, uint32(q.sqTail))
}

// collect consumes all available completion entries, invoking done for each
// one, and returns the number of collected entries.
func (q *queue) collect(done func(owner int, err error, result uint32)) (n int) {
	for {
		off := int(q.cqHead) * cqeSize
		dw3 := binary.LittleEndian.Uint32(q.cq[off+CQE_DW3:])

		if bits.GetN(&dw3, DW3_P, 1) != q.phase {
			break
		}

		result := binary.LittleEndian.Uint32(q.cq[off+CQE_DW0:])
		cid := uint16(bits.GetN(&dw3, DW3_CID, 0xffff))

		if int(cid) >= q.depth-1 {
			// should never happen, ignore spurious entry
		} else {
			var err error

			if bits.GetN(&dw3, DW3_SF, 0x7fff) != 0 {
				err = &StatusError{
					Opcode:     q.op[cid],
					Code:       uint8(bits.GetN(&dw3, DW3_SC, 0xff)),
					Type:       uint8(bits.GetN(&dw3, DW3_SCT, 0b111)),
					DoNotRetry: bits.Get(&dw3, DW3_DNR),
				}
			}

			owner := q.owner[cid]
			q.free = append(q.free, cid)
			done(owner, err, result)
		}

		if q.cqHead++; int(q.cqHead) == q.depth {
			q.cqHead = 0
			q.phase ^= 1
		}

		n++
	}

	if n > 0 {
		reg.Write64As32(q.cqDoorbell, uint32(q.cqHead))
	}

	return
}

// execute submits all commands, pipelined up to the queue depth, and waits
// for their completion. The done function is invoked for each completed
// command with its index. The queue lock must be held by the caller.
func (q *queue) execute(cmds []command, done func(i int, err error, result uint32)) (err error) {
	if q.failed != nil {
		return q.failed
	}

	next := 0
	pending := 0
	timeout := q.hw.Timeout

	for next < len(cmds) || pending > 0 {
		pushed := 0

		for next < len(cmds) && len(q.free) > 0 {
			if e := q.push(&cmds[next], next); e != nil {
				done(next, e, 0)
			} else {
				pushed++
			}

			next++
		}

		if pushed > 0 {
			q.ring()
			pending += pushed
		}

		if pending == 0 {
			break
		}

		if err = q.wait(done, &pending, timeout); err != nil {
			return
		}
	}

	return
}

// wait blocks until at least one completion is collected. The goroutine sleeps
// on the queue interrupt channel, completions are re-checked periodically only
// as a safety net against lost interrupts.
func (q *queue) wait(done func(int, error, uint32), pending *int, timeout time.Duration) error {
	var waited time.Duration

	for {
		// A wake up token might be stale, as completions signaled by
		// it could have been collected already, which only results in
		// an extra iteration. Interrupts are never lost as the token
		// is only consumed below, after the completion queue check.
		if n := q.collect(done); n > 0 {
			*pending -= n
			return nil
		}

		q.timer.Reset(recheck)

		select {
		case <-q.irq:
			q.timer.Stop()
			continue
		case <-q.timer.C:
		}

		if n := q.collect(done); n > 0 {
			*pending -= n
			return nil
		}

		if waited += recheck; waited >= timeout {
			q.failed = fmt.Errorf("queue %d command timeout", q.id)
			return q.failed
		}
	}
}

func (hw *NVMe) adminCommand(cmd command) (result uint32, err error) {
	q := hw.admin

	q.Lock()
	defer q.Unlock()

	var cmdErr error

	err = q.execute([]command{cmd}, func(_ int, e error, r uint32) {
		cmdErr = e
		result = r
	})

	if err == nil {
		err = cmdErr
	}

	return
}

// ioQueue returns an I/O queue, preferring an unlocked one, locked, or nil
// when the controller has no I/O queues.
func (hw *NVMe) ioQueue() *queue {
	n := uint32(len(hw.io))

	if n == 0 {
		return nil
	}

	start := hw.next.Add(1)

	for i := range n {
		if q := hw.io[(start+i)%n]; q.TryLock() {
			return q
		}
	}

	q := hw.io[start%n]
	q.Lock()

	return q
}
