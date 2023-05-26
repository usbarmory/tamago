// NXP Data Co-Processor (DCP) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package dcp implements a driver for the NXP Data Co-Processor (DCP)
// cryptographic accelerator adopting the following reference specifications:
//   - MCIMX28RM - i.MX28 Applications Processor Reference Manual - Rev 2 2013/08
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package dcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// DCP registers
const (
	DCP_CTRL     = 0x00
	CTRL_SFTRST  = 31
	CTRL_CLKGATE = 30

	DCP_STAT     = 0x10
	DCP_STAT_CLR = 0x18
	DCP_STAT_IRQ = 0

	DCP_CHANNELCTRL = 0x0020

	DCP_KEY     = 0x0060
	KEY_INDEX   = 4
	KEY_SUBWORD = 0

	DCP_KEYDATA   = 0x0070
	DCP_CH0CMDPTR = 0x0100
	DCP_CH0SEMA   = 0x0110

	DCP_CH0STAT        = 0x0120
	CHxSTAT_ERROR_CODE = 16
	CHxSTAT_ERROR_MASK = 0b1111110

	DCP_CH0STAT_CLR = 0x0128
)

// DCP channels
const (
	DCP_CHANNEL_0 = iota + 1
	DCP_CHANNEL_1
	DCP_CHANNEL_2
	DCP_CHANNEL_3
)

// DCP control packet settings
const (
	// p1068, 13.2.6.4.2 Control0 Field, MCIMX28RM

	DCP_CTRL0_HASH_TERM       = 13
	DCP_CTRL0_HASH_INIT       = 12
	DCP_CTRL0_OTP_KEY         = 10
	DCP_CTRL0_CIPHER_INIT     = 9
	DCP_CTRL0_CIPHER_ENCRYPT  = 8
	DCP_CTRL0_ENABLE_HASH     = 6
	DCP_CTRL0_ENABLE_CIPHER   = 5
	DCP_CTRL0_CHAIN           = 2
	DCP_CTRL0_DECR_SEMAPHORE  = 1
	DCP_CTRL0_INTERRUPT_ENABL = 0

	// p1070, 13.2.6.4.3 Control1 Field, MCIMX28RM
	// p1098, 13.3.11 DCP_PACKET2 field descriptions, MCIMX28RM

	DCP_CTRL1_HASH_SELECT = 16
	HASH_SELECT_SHA1      = 0x00
	HASH_SELECT_CRC32     = 0x01
	HASH_SELECT_SHA256    = 0x02

	DCP_CTRL1_KEY_SELECT  = 8
	KEY_SELECT_UNIQUE_KEY = 0xfe

	DCP_CTRL1_CIPHER_MODE = 4
	CIPHER_MODE_CBC       = 0x01

	DCP_CTRL1_CIPHER_SELECT = 0
	CIPHER_SELECT_AES128    = 0x00
)

const WorkPacketLength = 32

// WorkPacket represents a DCP work packet
// (p1067, 13.2.6.4 Work Packet Structure, MCIMX28RM).
type WorkPacket struct {
	NextCmdAddr              uint32
	Control0                 uint32
	Control1                 uint32
	SourceBufferAddress      uint32
	DestinationBufferAddress uint32
	BufferSize               uint32
	PayloadPointer           uint32
	Status                   uint32
}

// DCP represents the Data Co-Processor instance.
type DCP struct {
	sync.Mutex

	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int

	// DeriveKeyMemory represents the DMA memory region used for exchanging DCP
	// derived keys when the derivation index points to an internal DCP key RAM
	// slot. The memory region must be initialized before DeriveKey().
	//
	// It is recommended to use a DMA region within the internal RAM (e.g.
	// i.MX6 On-Chip OCRAM/iRAM) to avoid exposure to external RAM.
	//
	// The DeriveKey() function uses DeriveKeyMemory only if the default
	// DMA region start does not overlap with it.
	DeriveKeyMemory *dma.Region

	// control registers
	ctrl        uint32
	stat        uint32
	stat_clr    uint32
	chctrl      uint32
	key         uint32
	keydata     uint32
	ch0cmdptr   uint32
	ch0sema     uint32
	ch0stat     uint32
	ch0stat_clr uint32
}

// Bytes converts the DCP work packet structure to byte array format.
func (pkt *WorkPacket) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pkt)
	return buf.Bytes()
}

// Init initializes the DCP module.
func (hw *DCP) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid DCP instance")
	}

	hw.ctrl = hw.Base + DCP_CTRL
	hw.stat = hw.Base + DCP_STAT
	hw.stat_clr = hw.Base + DCP_STAT_CLR
	hw.chctrl = hw.Base + DCP_CHANNELCTRL
	hw.key = hw.Base + DCP_KEY
	hw.keydata = hw.Base + DCP_KEYDATA
	hw.ch0cmdptr = hw.Base + DCP_CH0CMDPTR
	hw.ch0sema = hw.Base + DCP_CH0SEMA
	hw.ch0stat = hw.Base + DCP_CH0STAT
	hw.ch0stat_clr = hw.Base + DCP_CH0STAT_CLR

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// soft reset DCP
	reg.Set(hw.ctrl, CTRL_SFTRST)
	reg.Clear(hw.ctrl, CTRL_SFTRST)

	// enable DCP
	reg.Clear(hw.ctrl, CTRL_CLKGATE)

	// enable channel 0
	reg.Write(hw.chctrl, DCP_CHANNEL_0)
}

func (hw *DCP) cmd(ptr uint, count int) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if reg.Read(hw.chctrl) != DCP_CHANNEL_0 {
		return errors.New("co-processor is not initialized")
	}

	// clear channel status
	reg.Write(hw.ch0stat_clr, 0xffffffff)

	// set command address
	reg.Write(hw.ch0cmdptr, uint32(ptr))
	// activate channel
	reg.SetN(hw.ch0sema, 0, 0xff, uint32(count))
	// wait for completion
	reg.Wait(hw.stat, DCP_STAT_IRQ, DCP_CHANNEL_0, 1)
	// clear interrupt register
	reg.Set(hw.stat_clr, DCP_CHANNEL_0)

	chstatus := reg.Read(hw.ch0stat)

	// check for errors
	if bits.Get(&chstatus, 0, CHxSTAT_ERROR_MASK) != 0 {
		code := bits.Get(&chstatus, CHxSTAT_ERROR_CODE, 0xff)
		sema := reg.Read(hw.ch0sema)
		return fmt.Errorf("DCP channel 0 error, status:%#x error_code:%#x sema:%#x", chstatus, code, sema)
	}

	return
}
