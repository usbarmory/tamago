// NXP Data Co-Processor (DCP) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// DCP registers
const (
	DCP_BASE = 0x02280000

	DCP_CTRL     = DCP_BASE
	CTRL_SFTRST  = 31
	CTRL_CLKGATE = 30

	DCP_STAT     = DCP_BASE + 0x10
	DCP_STAT_CLR = DCP_BASE + 0x18
	DCP_STAT_IRQ = 0

	DCP_CHANNELCTRL = DCP_BASE + 0x0020

	DCP_KEY     = DCP_BASE + 0x0060
	KEY_INDEX   = 4
	KEY_SUBWORD = 0

	DCP_KEYDATA   = DCP_BASE + 0x0070
	DCP_CH0CMDPTR = DCP_BASE + 0x0100
	DCP_CH0SEMA   = DCP_BASE + 0x0110

	DCP_CH0STAT        = DCP_BASE + 0x0120
	CHxSTAT_ERROR_CODE = 16
	CHxSTAT_ERROR_MASK = 0b1111110

	DCP_CH0STAT_CLR = DCP_BASE + 0x0128
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

// The i.MX6 On-Chip RAM (OCRAM/iRAM) is used for secure key passing between
// SoC components.
const (
	iramStart = 0x00900000
	iramSize  = 0x20000
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

// Bytes converts the DCP work packet structure to byte array format.
func (pkt *WorkPacket) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pkt)
	return buf.Bytes()
}

var mux sync.Mutex

// Init initializes the DCP module.
func init() {
	mux.Lock()
	mux.Unlock()

	// soft reset DCP
	reg.Set(DCP_CTRL, CTRL_SFTRST)
	reg.Clear(DCP_CTRL, CTRL_SFTRST)

	// enable clocks
	reg.Clear(DCP_CTRL, CTRL_CLKGATE)

	// enable channel 0
	reg.Write(DCP_CHANNELCTRL, DCP_CHANNEL_0)
}

func cmd(ptr uint32, count int) (err error) {
	mux.Lock()
	defer mux.Unlock()

	// clear channel status
	reg.Write(DCP_CH0STAT_CLR, 0xffffffff)

	// set command address
	reg.Write(DCP_CH0CMDPTR, ptr)
	// activate channel
	reg.SetN(DCP_CH0SEMA, 0, 0xff, uint32(count))
	// wait for completion
	reg.Wait(DCP_STAT, DCP_STAT_IRQ, DCP_CHANNEL_0, 1)
	// clear interrupt register
	reg.Set(DCP_STAT_CLR, DCP_CHANNEL_0)

	chstatus := reg.Read(DCP_CH0STAT)

	// check for errors
	if bits.Get(&chstatus, 0, CHxSTAT_ERROR_MASK) != 0 {
		code := bits.Get(&chstatus, CHxSTAT_ERROR_CODE, 0xff)
		sema := reg.Read(DCP_CH0SEMA)
		err = fmt.Errorf("DCP channel 0 error, status:%#x error_code:%#x sema:%#x", chstatus, code, sema)
	}

	return
}
