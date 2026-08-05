// Microchip CPU port module (DEVCPU)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package devcpu implements a driver for the Microchip CPU port module
// (DEVCPU), responsible for exchanging frames between the internal CPU system
// and the switch core, adopting the following reference specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package devcpu

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/analyzer"
)

const (
	ExtractionTimeout     = 1 * time.Second
	InjectionTimeout      = 1 * time.Second
	InjectionPollInterval = 1 * time.Millisecond
	MinimumFrameSize      = 60
)

type Stats struct {
	RxTruncated         uint64
	RxFIFOTimeouts      uint64
	RxAborted           uint64
	RxEmpty             uint64
	RxOversized         uint64
	RxInvalid           uint64
	RxShortHeader       uint64
	TxWatermarkTimeouts uint64
	TxFIFOTimeouts      uint64
	TxAborts            uint64
	TxAbortTimeouts     uint64
}

// Port represents a CPU port module.
type Port struct {
	sync.Mutex

	// Port index
	Index int
	// Group index
	Group int
	// Interrupt ID
	IRQ int

	// Queue System Base register
	Queue uint32
	// Analyzer block
	Analyzer *analyzer.ANA
	// HeaderLength allows to override the default internal frame header
	// length [see IFH_LEN].
	HeaderLength int

	// Enable implements port enabling and MAC learning
	Enable func() error

	// MAC address (use SetMAC() for post Init() changes)
	MAC net.HardwareAddr
	// FID represents the VLAN filtering identifier
	FID int

	// Statistics
	Stats Stats

	rxMu sync.Mutex
	txMu sync.Mutex

	// control registers
	xtr_rd     uint32
	inj_ctrl   uint32
	inj_wr     uint32
	inj_status uint32
}

// Init initializes a CPU port module.
func (p *Port) Init() (err error) {
	p.Lock()
	defer p.Unlock()

	if p.Queue == 0 {
		return errors.New("invalid port instance")
	}

	if p.Group < 0 || p.Group > 1 {
		return errors.New("invalid port group")
	}

	p.Stats = Stats{}

	if p.MAC == nil {
		p.MAC = make([]byte, 6)
		rand.Read(p.MAC)
		// flag address as unicast and locally administered
		p.MAC[0] &= 0xfe
		p.MAC[0] |= 0x02
	} else if len(p.MAC) != 6 {
		return errors.New("invalid MAC")
	}

	if p.HeaderLength == 0 {
		p.HeaderLength = IFH_LEN
	}

	if p.Enable != nil {
		if err = p.Enable(); err != nil {
			return
		}
	}

	groupOffset := uint32(p.Group) * 4
	p.xtr_rd = p.Queue + XTR_RD + groupOffset
	p.inj_ctrl = p.Queue + INJ_CTRL + groupOffset
	p.inj_wr = p.Queue + INJ_WR + groupOffset
	p.inj_status = p.Queue + INJ_STATUS

	// set manual injection/extraction for CPU queue
	reg.Write(p.Queue+INJ_GRP_CFG+groupOffset, 1<<CFG_MODE|1<<CFG_BYTE_SWAP)
	reg.Write(p.Queue+XTR_GRP_CFG+groupOffset, 1<<CFG_MODE|1<<CFG_STATUS_WORD_POS|1<<CFG_BYTE_SWAP)
	reg.Write(p.Queue+XTR_FRM_PRUNING+groupOffset, 0)

	// abort stale frame
	if !p.abortInjection() {
		return errors.New("injection abort failed")
	}

	// add physical address to MAC table
	p.SetMAC(p.MAC)

	return
}

// SetMAC allows to change the controller physical address register after
// initialization.
func (p *Port) SetMAC(mac net.HardwareAddr) {
	if len(mac) != 6 {
		return
	}

	if len(p.MAC) != 0 {
		p.Analyzer.Delete(p.MAC, uint32(p.FID), analyzer.PGID_HOST)
	}

	p.Analyzer.Insert(mac, uint32(p.FID), analyzer.PGID_HOST)
	p.MAC = mac
}

func (p *Port) abortInjection() (complete bool) {
	var deadline time.Time

	for reg.Get(p.inj_status, INJ_STATUS_INJ_IN_PROGRESS+p.Group) {
		if deadline.IsZero() {
			reg.Set(p.inj_ctrl, CTRL_ABORT)
			deadline = time.Now().Add(InjectionTimeout)
		}

		if !time.Now().Before(deadline) {
			p.Stats.TxAbortTimeouts++
			return
		}

		runtime.Gosched()
	}

	return true
}

func (p *Port) read() (val uint32, err error) {
	var deadline time.Time

	for val = reg.Read(p.xtr_rd); val == RD_NOT_READY; val = reg.Read(p.xtr_rd) {
		if deadline.IsZero() {
			deadline = time.Now().Add(ExtractionTimeout)
		}

		if !time.Now().Before(deadline) {
			p.Stats.RxFIFOTimeouts++
			return 0, errors.New("extraction FIFO not ready")
		}

		runtime.Gosched()
	}

	return
}

func (p *Port) write(val uint32) (err error) {
	var deadline time.Time

	for !reg.Get(p.inj_status, INJ_STATUS_FIFO_RDY+p.Group) {
		if deadline.IsZero() {
			deadline = time.Now().Add(InjectionTimeout)
		}

		if !time.Now().Before(deadline) {
			p.Stats.TxFIFOTimeouts++
			return errors.New("injection FIFO not ready")
		}

		runtime.Gosched()
	}

	reg.Write(p.inj_wr, val)

	return
}

func (p *Port) recv(buf []byte) (eof bool, unused int, err error) {
	val, err := p.read()

	if err != nil {
		return
	}

	switch val {
	case RD_EOF_UNUSED_0:
		eof = true
	case RD_EOF_UNUSED_1:
		eof = true
		unused = 1
	case RD_EOF_UNUSED_2:
		eof = true
		unused = 2
	case RD_EOF_UNUSED_3:
		eof = true
		unused = 3
	case RD_EOF_TRUNCATED:
		p.read()
		p.Stats.RxTruncated++
		err = errors.New("truncated frame")
	case RD_EOF_ABORTED:
		p.Stats.RxAborted++
		err = errors.New("aborted frame")
	case RD_ESCAPE:
		if val, err = p.read(); err == nil {
			binary.LittleEndian.PutUint32(buf, val)
		}
	default:
		binary.LittleEndian.PutUint32(buf, val)
	}

	return
}

// Receive receives a single Ethernet frame from a port module.
func (p *Port) Receive(buf []byte) (n int, err error) {
	p.rxMu.Lock()
	defer p.rxMu.Unlock()

	if !reg.Get(p.Queue+XTR_DATA_PRESENT, p.Group) {
		return
	}

	// rx FIFO reads are 32-bits of frame data
	var scratch [4]byte

	// skip internal frame header
	for i := 0; i < p.HeaderLength; i += 4 {
		eof, _, err := p.recv(scratch[:])

		switch {
		case err != nil:
			return 0, err
		case eof:
			p.Stats.RxShortHeader++
			return 0, errors.New("short frame header")
		}
	}

	padded := 0

	for {
		dst := scratch[:]

		if len(buf)-padded >= len(scratch) {
			dst = buf[padded : padded+len(scratch)]
		}

		eof, unused, err := p.recv(dst)

		if err != nil {
			return 0, err
		}

		if eof {
			if unused > padded {
				p.Stats.RxInvalid++
				return 0, errors.New("invalid frame length")
			}

			switch length := padded - unused; {
			case length == 0:
				p.Stats.RxEmpty++
				return 0, errors.New("empty frame")
			case length > len(buf):
				p.Stats.RxOversized++
				return n, errors.New("frame exceeds receive buffer")
			default:
				return length, nil
			}
		}

		padded += len(scratch)
	}
}

func (p *Port) waitForInjection() (err error) {
	var deadline time.Time

	group := uint32(1 << p.Group)
	fifoReady := group << INJ_STATUS_FIFO_RDY
	watermarkReached := group << INJ_STATUS_WMARK_REACHED

	for {
		status := reg.Read(p.inj_status)
		watermarked := status&watermarkReached != 0

		if !watermarked && status&fifoReady != 0 {
			return
		}

		now := time.Now()

		if deadline.IsZero() {
			deadline = now.Add(InjectionTimeout)
		} else if !now.Before(deadline) {
			if watermarked {
				p.Stats.TxWatermarkTimeouts++
				return errors.New("injection watermark reached")
			}

			p.Stats.TxFIFOTimeouts++
			return errors.New("injection FIFO not ready")
		}

		if watermarked {
			time.Sleep(InjectionPollInterval)
		} else {
			runtime.Gosched()
		}
	}
}

// Transmit transmits a single Ethernet frame to a port module.
func (p *Port) Transmit(buf []byte) (err error) {
	p.txMu.Lock()
	defer p.txMu.Unlock()

	if err = p.waitForInjection(); err != nil {
		return
	}

	valid := len(buf) % 4
	written := 0

	if len(buf) < MinimumFrameSize {
		valid = 0
	}

	// signal Start Of Frame
	reg.Set(p.inj_ctrl, CTRL_SOF)

	defer func() {
		if err != nil {
			p.Stats.TxAborts++
			p.abortInjection()
		}
	}()

	// tx FIFO reads are 32-bits of frame data
	var scratch [4]byte

	for len(buf)-written >= len(scratch) {
		if err = p.write(binary.LittleEndian.Uint32(buf[written:])); err != nil {
			return
		}

		written += len(scratch)
	}

	if written < len(buf) {
		copy(scratch[:], buf[written:])

		if err = p.write(binary.LittleEndian.Uint32(scratch[:])); err != nil {
			return
		}

		written += len(scratch)
	}

	for written < MinimumFrameSize {
		if err = p.write(0); err != nil {
			return
		}

		written += len(scratch)
	}

	// set valid bytes of last word
	reg.SetN(p.inj_ctrl, CTRL_VLD_BYTES, 0b11, uint32(valid))

	// signal End Of Frame
	reg.Set(p.inj_ctrl, CTRL_EOF)

	// add dummy CRC
	err = p.write(0)

	return
}
