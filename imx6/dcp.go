// NXP Data Co-Processor (DCP) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/f-secure-foundry/tamago/imx6/internal/mem"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	HW_DCP_BASE uint32 = 0x02280000

	HW_DCP_CTRL         = HW_DCP_BASE + 0x00
	HW_DCP_CTRL_SFTRST  = 31
	HW_DCP_CTRL_CLKGATE = 30

	HW_DCP_STAT     = HW_DCP_BASE + 0x10
	HW_DCP_STAT_CLR = HW_DCP_BASE + 0x18
	HW_DCP_STAT_IRQ = 0

	HW_DCP_CHANNELCTRL = HW_DCP_BASE + 0x0020
	HW_DCP_CH0CMDPTR   = HW_DCP_BASE + 0x0100
	HW_DCP_CH0SEMA     = HW_DCP_BASE + 0x0110
	HW_DCP_CH0STAT     = HW_DCP_BASE + 0x0120
	HW_DCP_CH0STAT_CLR = HW_DCP_BASE + 0x0128

	SNVS_HPSR_REG               uint32 = 0x020cc014
	SNVS_HPSR_SSM_STATE                = 8
	SNVS_HPSR_SSM_STATE_TRUSTED        = 0b1101
	SNVS_HPSR_SSM_STATE_SECURE         = 0b1111
)

const (
	// p1068, 13.2.6.4.2 Control0 Field, MCIMX28RM
	DCP_CTRL0_OTP_KEY         = 10
	DCP_CTRL0_CIPHER_INIT     = 9
	DCP_CTRL0_CIPHER_ENCRYPT  = 8
	DCP_CTRL0_ENABLE_CIPHER   = 5
	DCP_CTRL0_DECR_SEMAPHORE  = 1
	DCP_CTRL0_INTERRUPT_ENABL = 0
	// p1070, 13.2.6.4.3 Control1 Field, MCIMX28RM
	DCP_CTRL1_KEY_SELECT    = 8
	DCP_CTRL1_CIPHER_MODE   = 4
	DCP_CTRL1_CIPHER_SELECT = 0
	// p1098, 13.3.11 HW_DCP_PACKET2 field descriptions, MCIMX28RM
	AES128     = 0x0
	CBC        = 0x1
	UNIQUE_KEY = 0xfe
)

const (
	DCP_CHANNEL_1 = iota
	DCP_CHANNEL_2
	DCP_CHANNEL_3
	DCP_CHANNEL_4
	DCP_CHANNEL_MAX
)

// p1067, 13.2.6.4 Work Packet Structure, MCIMX28RM
type WorkPacket struct {
	NextCmdAddr              uint32
	Control0                 uint32
	Control1                 uint32
	SourceBufferAddress      uint32
	DestinationBufferAddress uint32
	BufferSize               uint32
	PayloadPointer           uint32
	Status                   uint32
	Pad_cgo_0                [4]byte
}

type dcp struct {
	sync.Mutex
}

var DCP = &dcp{}

// Init initializes the DCP module.
func (hw *dcp) Init() {
	hw.Lock()
	// note: cannot defer during initialization

	// soft reset DCP
	reg.Set(HW_DCP_CTRL, HW_DCP_CTRL_SFTRST)
	reg.Clear(HW_DCP_CTRL, HW_DCP_CTRL_SFTRST)

	// enable clocks
	reg.Clear(HW_DCP_CTRL, HW_DCP_CTRL_CLKGATE)

	// enable all channels with merged IRQs
	reg.Write(HW_DCP_CHANNELCTRL, 0x000100ff)

	// enable all channel interrupts
	reg.SetN(HW_DCP_CHANNELCTRL, 0, 0xff, 0xff)

	hw.Unlock()
}

// SNVS verifies whether the Secure Non Volatile Storage (SNVS) is available in
// Trusted or Secure state (indicating that Secure Boot is enabled).
//
// The unique OTPMK internal key is available only when Secure Boot (HAB) is
// enabled, otherwise a Non-volatile Test Key (NVTK), identical for each SoC,
// is used. The secure operation of the DCP and SNVS, in production
// deployments, should always be paired with Secure Boot activation.
func (hw *dcp) SNVS() bool {
	ssm := reg.Get(SNVS_HPSR_REG, SNVS_HPSR_SSM_STATE, 0b1111)

	switch ssm {
	case SNVS_HPSR_SSM_STATE_TRUSTED, SNVS_HPSR_SSM_STATE_SECURE:
		return true
	default:
		return false
	}
}

// DeriveKey derives a hardware unique key in a manner equivalent to PKCS#11
// C_DeriveKey with CKM_AES_CBC_ENCRYPT_DATA.
//
// The diversifier is AES-CBC encrypted using the internal OTPMK key (when SNVS
// is enabled).
func (hw *dcp) DeriveKey(diversifier []byte, iv []byte) (key []byte, err error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("invalid IV size")
	}

	if len(diversifier) > aes.BlockSize {
		return nil, errors.New("invalid diversifier size")
	}

	if !hw.SNVS() {
		err = errors.New("SNVS unavailable, not in trusted or secure state")
		return
	}

	// p1057, 13.1.1 DCP Limitations for Software, MCIMX28RM
	// * buffer size must be aligned to 16 bytes for AES operations
	diversifier = pad(diversifier, false)
	key = make([]byte, len(diversifier))

	// p1057, 13.1.1 DCP Limitations for Software, MCIMX28RM
	// * any byte alignment is supported but 4 bytes one leads to better
	//   performance
	workPacket := WorkPacket{}

	workPacket.Control0 |= (1 << DCP_CTRL0_INTERRUPT_ENABL)
	workPacket.Control0 |= (1 << DCP_CTRL0_DECR_SEMAPHORE)
	workPacket.Control0 |= (1 << DCP_CTRL0_ENABLE_CIPHER)
	workPacket.Control0 |= (1 << DCP_CTRL0_CIPHER_ENCRYPT)
	workPacket.Control0 |= (1 << DCP_CTRL0_CIPHER_INIT)
	// Use device-specific hardware key, payload does not contain the key.
	workPacket.Control0 |= (1 << DCP_CTRL0_OTP_KEY)

	workPacket.Control1 |= (AES128 << DCP_CTRL1_CIPHER_SELECT)
	workPacket.Control1 |= (CBC << DCP_CTRL1_CIPHER_MODE)
	workPacket.Control1 |= (UNIQUE_KEY << DCP_CTRL1_KEY_SELECT)

	hw.Lock()
	defer hw.Unlock()

	workPacket.BufferSize = uint32(len(diversifier))
	workPacket.SourceBufferAddress = mem.Alloc(diversifier, 0)
	defer mem.Free(workPacket.SourceBufferAddress)

	workPacket.DestinationBufferAddress = mem.Alloc(key, 0)
	defer mem.Free(workPacket.DestinationBufferAddress)

	// p1073, Table 13-12. DCP Payload Field, MCIMX28RM
	workPacket.PayloadPointer = mem.Alloc(iv, 0)
	defer mem.Free(workPacket.PayloadPointer)

	// clear channel status
	reg.Write(HW_DCP_CH0STAT_CLR, 0xffffffff)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &workPacket)

	pkt := mem.Alloc(buf.Bytes(), 0)
	defer mem.Free(pkt)

	reg.Write(HW_DCP_CH0CMDPTR, pkt)
	reg.Set(HW_DCP_CH0SEMA, 0)

	// channel 0 is used
	log.Printf("imx6_dcp: waiting for key derivation")
	reg.Wait(HW_DCP_STAT, HW_DCP_STAT_IRQ, 0b1, 1)
	reg.Set(HW_DCP_STAT_CLR, 1)

	if chstatus := reg.Get(HW_DCP_CH0STAT, 1, 0b111111); chstatus != 0 {
		code := reg.Get(HW_DCP_CH0STAT, 16, 0xff)
		return nil, fmt.Errorf("DCP channel 0 error, status:%#x error_code:%#x", chstatus, code)
	}

	key = mem.Read(workPacket.DestinationBufferAddress, 0, len(key))

	return
}

func pad(buf []byte, extraBlock bool) []byte {
	padLen := 0
	r := len(buf) % aes.BlockSize

	if r != 0 {
		padLen = aes.BlockSize - r
	} else if extraBlock {
		padLen = aes.BlockSize
	}

	padding := []byte{(byte)(padLen)}
	padding = bytes.Repeat(padding, padLen)
	buf = append(buf, padding...)

	return buf
}
