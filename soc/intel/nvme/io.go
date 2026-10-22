// NVM Express driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package nvme

import (
	"errors"
	"fmt"
	"io"

	"github.com/usbarmory/tamago/bits"
)

// NVM command set opcodes
const (
	NVM_FLUSH = 0x00
	NVM_WRITE = 0x01
	NVM_READ  = 0x02
)

// Read/Write Command Dword 12
// (NVM Express NVM Command Set Specification 1.0 - 3.2.4 Read command).
const (
	CDW12_LR     = 31
	CDW12_FUA    = 30
	CDW12_PRINFO = 26
	CDW12_NLB    = 0
)

const maxBlocksPerCommand = 1 << 16

// ErrUnaligned is returned when an I/O offset or length is not a multiple
// of the namespace block size.
var ErrUnaligned = errors.New("offset and length must be block aligned")

// ErrNotInitialized is returned when I/O is requested on a controller without
// I/O queues (e.g. after a failed initialization).
var ErrNotInitialized = errors.New("controller not initialized")

// Op represents an I/O request operation.
type Op uint8

const (
	Read Op = iota
	Write
	Flush
)

// Request represents a single I/O request for [Namespace.Do].
type Request struct {
	// Op is the requested operation.
	Op Op
	// LBA is the starting logical block address ([Read], [Write]).
	LBA uint64
	// Buf is the data buffer ([Read], [Write]), its length must be a
	// multiple of the namespace block size.
	//
	// Buffers reserved by the application within [NVMe.Region] (see
	// [dma.Region.Reserve]), or slices of them, are used directly by the
	// controller (zero copy), any other buffer is copied to and from a
	// temporary allocation (see [dma.Region.Alloc]).
	Buf []byte
	// FUA forces the write to non-volatile media before completion
	// ([Write]).
	FUA bool

	// Err holds the request completion result.
	Err error
}

// Namespace represents an NVMe namespace, a logical block device.
type Namespace struct {
	// ID is the namespace identifier.
	ID uint32
	// Blocks is the namespace size in logical blocks (NSZE).
	Blocks uint64
	// Capacity is the namespace capacity in logical blocks (NCAP).
	Capacity uint64
	// BlockSize is the logical block size in bytes.
	BlockSize int
	// Metadata is the per block metadata size, namespaces with metadata
	// are not supported for I/O.
	Metadata int

	hw        *NVMe
	maxBlocks int
}

// Size returns the namespace size in bytes.
func (ns *Namespace) Size() int64 {
	return int64(ns.Blocks) * int64(ns.BlockSize)
}

type transfer struct {
	req  *Request
	addr uint
}

// Do executes a batch of I/O requests: data transfers are split according to
// the controller maximum transfer size and all resulting commands are
// pipelined on a single queue pair. Requests are not ordered with respect to
// each other, a [Flush] request only covers writes completed before Do is
// invoked.
//
// The returned error is the first request error, each request Err field
// holds its own result.
func (ns *Namespace) Do(reqs ...*Request) (err error) {
	if ns.Metadata != 0 {
		return errors.New("namespaces with metadata are not supported")
	}

	bs := uint64(ns.BlockSize)

	var cmds []command
	var owners []*Request
	var transfers []transfer

	defer func() {
		// no-op for buffers reserved by the application
		for _, t := range transfers {
			ns.hw.Region.Free(t.addr)
		}
	}()

	for _, req := range reqs {
		req.Err = nil

		if req.Op == Flush {
			cmds = append(cmds, command{opcode: NVM_FLUSH, nsid: ns.ID})
			owners = append(owners, req)
			continue
		}

		var opcode uint8

		switch req.Op {
		case Read:
			opcode = NVM_READ
		case Write:
			opcode = NVM_WRITE
		default:
			req.Err = errors.New("invalid operation")
			continue
		}

		n := uint64(len(req.Buf))

		if n == 0 {
			continue
		}

		if n%bs != 0 {
			req.Err = ErrUnaligned
			continue
		}

		if blocks := n / bs; req.LBA >= ns.Blocks || blocks > ns.Blocks-req.LBA {
			req.Err = fmt.Errorf("LBA range out of bounds (%d+%d)", req.LBA, blocks)
			continue
		}

		addr := ns.hw.Region.Alloc(req.Buf, 0)
		transfers = append(transfers, transfer{req: req, addr: addr})

		if addr&3 != 0 {
			req.Err = errors.New("DMA buffer must be 4-byte aligned")
			continue
		}

		chunk := uint64(ns.maxBlocks) * bs

		for off := uint64(0); off < n; off += chunk {
			length := min(chunk, n-off)
			lba := req.LBA + off/bs

			cmd := command{
				opcode: opcode,
				nsid:   ns.ID,
				addr:   addr + uint(off),
				length: int(length),
			}

			// Starting LBA (CDW10-11), Number of Logical Blocks
			// (0's based)
			cmd.cdw[0] = uint32(lba)
			cmd.cdw[1] = uint32(lba >> 32)
			bits.SetN(&cmd.cdw[2], CDW12_NLB, 0xffff, uint32(length/bs-1))

			if req.FUA && opcode == NVM_WRITE {
				bits.Set(&cmd.cdw[2], CDW12_FUA)
			}

			cmds = append(cmds, cmd)
			owners = append(owners, req)
		}
	}

	if len(cmds) > 0 {
		q := ns.hw.ioQueue()

		if q == nil {
			return ErrNotInitialized
		}

		err = q.execute(cmds, func(i int, e error, _ uint32) {
			if req := owners[i]; e != nil && req.Err == nil {
				req.Err = e
			}
		})

		q.Unlock()

		if err != nil {
			// the controller might still access the buffers of
			// commands which did not complete, leave them allocated
			transfers = nil
			return
		}
	}

	// copies back unless reserved by the application
	for _, t := range transfers {
		if t.req.Op == Read && t.req.Err == nil {
			ns.hw.Region.Read(t.addr, 0, t.req.Buf)
		}
	}

	for _, req := range reqs {
		if req.Err != nil {
			return req.Err
		}
	}

	return
}

// ReadBlocks reads len(buf) bytes starting at the given logical block.
func (ns *Namespace) ReadBlocks(lba uint64, buf []byte) error {
	return ns.Do(&Request{Op: Read, LBA: lba, Buf: buf})
}

// WriteBlocks writes len(buf) bytes starting at the given logical block,
// with fua set the data is on non-volatile media when the function returns.
func (ns *Namespace) WriteBlocks(lba uint64, buf []byte, fua bool) error {
	return ns.Do(&Request{Op: Write, LBA: lba, Buf: buf, FUA: fua})
}

// Flush commits all completed writes to non-volatile media.
func (ns *Namespace) Flush() error {
	return ns.Do(&Request{Op: Flush})
}

func (ns *Namespace) rw(op Op, p []byte, off int64) (n int, err error) {
	bs := int64(ns.BlockSize)

	if off < 0 || off%bs != 0 || int64(len(p))%bs != 0 {
		return 0, ErrUnaligned
	}

	size := ns.Size()

	if off >= size {
		if op == Read {
			return 0, io.EOF
		}

		return 0, errors.New("write beyond end of namespace")
	}

	n = len(p)

	if remaining := size - off; int64(n) > remaining {
		n = int(remaining)
		p = p[:n]
	}

	if err = ns.Do(&Request{Op: op, LBA: uint64(off / bs), Buf: p}); err != nil {
		return 0, err
	}

	return
}

// WriteAt implements [io.WriterAt], off and len(p) must be multiples of the
// block size.
func (ns *Namespace) WriteAt(p []byte, off int64) (n int, err error) {
	want := len(p)

	if n, err = ns.rw(Write, p, off); err == nil && n < want {
		err = io.ErrShortWrite
	}

	return
}

// ReadAt implements [io.ReaderAt], off and len(p) must be multiples of the
// block size.
func (ns *Namespace) ReadAt(p []byte, off int64) (n int, err error) {
	want := len(p)

	if n, err = ns.rw(Read, p, off); err == nil && n < want {
		err = io.EOF
	}

	return
}
