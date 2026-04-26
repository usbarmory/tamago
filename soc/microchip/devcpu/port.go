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
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/analyzer"
)

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

	// control registers
	xtr_rd   uint32
	inj_ctrl uint32
	inj_wr   uint32
}

// Init initializes a CPU port module.
func (p *Port) Init() (err error) {
	p.Lock()
	defer p.Unlock()

	if p.Queue == 0 {
		return errors.New("invalid port instance")
	}

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

	// set manual injection/extraction for CPU queue
	reg.SetN(p.Queue+INJ_GRP_CFG+groupOffset, CFG_MODE, 0b11, 1)
	reg.SetN(p.Queue+XTR_GRP_CFG+groupOffset, CFG_MODE, 0b11, 1)

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

func (p *Port) recv(buf []byte) (ok bool) {
	switch val := reg.Read(p.xtr_rd); val {
	case RD_EOF_UNUSED_0, RD_EOF_UNUSED_1, RD_EOF_UNUSED_2, RD_EOF_UNUSED_3:
		return false
	case RD_EOF_TRUNCATED, RD_EOF_ABORTED, RD_ESCAPE:
		return false
	case RD_NOT_READY:
		return false
	default:
		binary.LittleEndian.PutUint32(buf, val)
	}

	return true
}

// Receive receives a single Ethernet frame from a port module.
func (p *Port) Receive(buf []byte) (n int, err error) {
	if !reg.Get(p.Queue+XTR_DATA_PRESENT, p.Group) {
		return
	}

	// skip internal frame header
	for i := 0; i < p.HeaderLength; i += 4 {
		reg.Read(p.xtr_rd)
	}

	length := len(buf)
	r := length % 4

	for i := 0; i < length-r; i += 4 {
		if p.recv(buf[i:]) {
			n += 4
		} else {
			return
		}
	}

	if r > 0 {
		if b := make([]byte, 4); p.recv(b) {
			copy(buf[length-r:], b)
			n += r

			// reach EOF
			p.recv(b)
		}
	}

	return
}

// Transmit transmits a single Ethernet frame to a port module.
func (p *Port) Transmit(buf []byte) (err error) {
	// signal Start Of Frame
	reg.Set(p.inj_ctrl, CTRL_SOF)

	// pad to word size
	if r := len(buf) % 4; r != 0 {
		buf = append(buf, make([]byte, 4-r)...)
	}

	for i := 0; i < len(buf); i += 4 {
		reg.Write(p.inj_wr, binary.LittleEndian.Uint32(buf[i:]))
	}

	// set valid bytes of last word
	reg.SetN(p.inj_ctrl, CTRL_VLD_BYTES, 0b11, uint32(len(buf)%4))

	// signal End Of Frame
	reg.Set(p.inj_ctrl, CTRL_EOF)

	// add dummy CRC
	reg.Write(p.inj_wr, 0)

	return
}
