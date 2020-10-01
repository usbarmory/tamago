// NXP Data Co-Processor (DCP) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/dma"
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

	SNVS_HPSR_REG     = 0x020cc014
	SSM_STATE         = 8
	SSM_STATE_TRUSTED = 0b1101
	SSM_STATE_SECURE  = 0b1111
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
	DCP_CTRL0_OTP_KEY         = 10
	DCP_CTRL0_CIPHER_INIT     = 9
	DCP_CTRL0_CIPHER_ENCRYPT  = 8
	DCP_CTRL0_ENABLE_CIPHER   = 5
	DCP_CTRL0_CHAIN           = 2
	DCP_CTRL0_DECR_SEMAPHORE  = 1
	DCP_CTRL0_INTERRUPT_ENABL = 0
	// p1070, 13.2.6.4.3 Control1 Field, MCIMX28RM
	DCP_CTRL1_KEY_SELECT    = 8
	DCP_CTRL1_CIPHER_MODE   = 4
	DCP_CTRL1_CIPHER_SELECT = 0
	// p1098, 13.3.11 DCP_PACKET2 field descriptions, MCIMX28RM
	AES128     = 0x0
	CBC        = 0x1
	UNIQUE_KEY = 0xfe
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

// SetDefaults initializes default values for the DCP work packet.
func (pkt *WorkPacket) SetDefaults() {
	pkt.Control0 |= 1 << DCP_CTRL0_DECR_SEMAPHORE
	pkt.Control0 |= 1 << DCP_CTRL0_ENABLE_CIPHER
	pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_INIT

	pkt.Control1 |= AES128 << DCP_CTRL1_CIPHER_SELECT
	pkt.Control1 |= CBC << DCP_CTRL1_CIPHER_MODE
}

// Bytes converts the DCP work packet structure to byte array format.
func (pkt *WorkPacket) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pkt)
	return buf.Bytes()
}

// Dcp represents a DCP module instance
type Dcp struct {
	sync.Mutex
}

// DCP (Data Co-Processor) module instance
var DCP = &Dcp{}

// Init initializes the DCP module.
func (hw *Dcp) Init() {
	hw.Lock()
	hw.Unlock()

	// soft reset DCP
	reg.Set(DCP_CTRL, CTRL_SFTRST)
	reg.Clear(DCP_CTRL, CTRL_SFTRST)

	// enable clocks
	reg.Clear(DCP_CTRL, CTRL_CLKGATE)

	// enable all channels with merged IRQs
	reg.Write(DCP_CHANNELCTRL, 0x000100ff)

	// enable all channel interrupts
	reg.SetN(DCP_CHANNELCTRL, 0, 0xff, 0xff)
}

// SNVS verifies whether the Secure Non Volatile Storage (SNVS) is available in
// Trusted or Secure state (indicating that Secure Boot is enabled).
//
// The unique OTPMK internal key is available only when Secure Boot (HAB) is
// enabled, otherwise a Non-volatile Test Key (NVTK), identical for each SoC,
// is used. The secure operation of the DCP and SNVS, in production
// deployments, should always be paired with Secure Boot activation.
func (hw *Dcp) SNVS() bool {
	ssm := reg.Get(SNVS_HPSR_REG, SSM_STATE, 0b1111)

	switch ssm {
	case SSM_STATE_TRUSTED, SSM_STATE_SECURE:
		return true
	default:
		return false
	}
}

func (hw *Dcp) cmd(ptr uint32, count int) (err error) {
	hw.Lock()
	defer hw.Unlock()

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

func (hw *Dcp) cipher(buf []byte, index int, iv []byte, enc bool) (err error) {
	if len(buf)%aes.BlockSize != 0 {
		return errors.New("invalid input size")
	}

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	if len(iv) != aes.BlockSize {
		return errors.New("invalid IV size")
	}

	pkt := &WorkPacket{}
	pkt.SetDefaults()
	pkt.Control0 |= 1 << DCP_CTRL0_INTERRUPT_ENABL

	if enc {
		pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	}

	// use key RAM slot
	pkt.Control1 |= (uint32(index) & 0xff) << DCP_CTRL1_KEY_SELECT

	pkt.BufferSize = uint32(len(buf))
	pkt.SourceBufferAddress = dma.Alloc(buf, aes.BlockSize)

	pkt.DestinationBufferAddress = pkt.SourceBufferAddress
	defer dma.Free(pkt.SourceBufferAddress)

	pkt.PayloadPointer = dma.Alloc(iv, 0)
	defer dma.Free(pkt.PayloadPointer)

	ptr := dma.Alloc(pkt.Bytes(), 0)
	defer dma.Free(ptr)

	err = hw.cmd(ptr, 1)

	if err != nil {
		return
	}

	dma.Read(pkt.DestinationBufferAddress, 0, buf)

	return
}

// DeriveKey derives a hardware unique key in a manner equivalent to PKCS#11
// C_DeriveKey with CKM_AES_CBC_ENCRYPT_DATA.
//
// The diversifier is AES-CBC encrypted using the internal OTPMK key (when SNVS
// is enabled).
//
// A negative index argument results in the derived key being computed and
// returned.
//
// An index argument equal or greater than 0 moves the derived key directly to
// the corresponding internal DCP key RAM slot (see SetKey()). This is
// accomplished through an iRAM reserved DMA buffer, to ensure that the key is
// never exposed to external RAM or the Go runtime. In this case no key is
// returned by the function.
func (hw *Dcp) DeriveKey(diversifier []byte, iv []byte, index int) (key []byte, err error) {
	if !hw.SNVS() {
		return nil, errors.New("SNVS unavailable, not in trusted or secure state")
	}

	if len(iv) != aes.BlockSize {
		return nil, errors.New("invalid IV size")
	}

	// prepare diversifier for in-place encryption
	key = pad(diversifier, false)

	region := dma.Default()

	if index >= 0 {
		// force use of iRAM if not already set as default DMA region
		if region.Start < iramStart || region.Start > iramStart+iramSize {
			region = &dma.Region{
				Start: iramStart,
				Size:  iramSize,
			}

			region.Init()
		}
	}

	pkt := &WorkPacket{}
	pkt.SetDefaults()
	pkt.Control0 |= 1 << DCP_CTRL0_INTERRUPT_ENABL

	// Use device-specific hardware key for encryption.
	pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	pkt.Control0 |= 1 << DCP_CTRL0_OTP_KEY
	pkt.Control1 |= UNIQUE_KEY << DCP_CTRL1_KEY_SELECT

	pkt.BufferSize = uint32(len(key))
	pkt.SourceBufferAddress = region.Alloc(key, aes.BlockSize)

	pkt.DestinationBufferAddress = pkt.SourceBufferAddress
	defer region.Free(pkt.SourceBufferAddress)

	pkt.PayloadPointer = region.Alloc(iv, 0)
	defer region.Free(pkt.PayloadPointer)

	ptr := region.Alloc(pkt.Bytes(), 0)
	defer region.Free(ptr)

	err = hw.cmd(ptr, 1)

	if err != nil {
		return
	}

	if index >= 0 {
		err = hw.setKeyData(index, nil, pkt.SourceBufferAddress)
	} else {
		region.Read(pkt.SourceBufferAddress, 0, key)
	}

	return
}

func (hw *Dcp) setKeyData(index int, key []byte, addr uint32) (err error) {
	var keyLocation uint32
	var subword uint32

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	if key != nil && len(key) > aes.BlockSize {
		return errors.New("invalid key size")
	}

	bits.SetN(&keyLocation, KEY_INDEX, 0b11, uint32(index))

	hw.Lock()
	defer hw.Unlock()

	for subword < 4 {
		off := subword * 4

		bits.SetN(&keyLocation, KEY_SUBWORD, 0b11, subword)
		reg.Write(DCP_KEY, keyLocation)

		if key != nil {
			k := key[off : off+4]
			reg.Write(DCP_KEYDATA, binary.LittleEndian.Uint32(k))
		} else {
			reg.Move(DCP_KEYDATA, addr+off)
		}

		subword++
	}

	return
}

// SetKey configures an AES-128 key in one of the 4 available slots of the DCP
// key RAM.
func (hw *Dcp) SetKey(index int, key []byte) (err error) {
	return hw.setKeyData(index, key, 0)
}

// Encrypt performs in-place buffer encryption using AES-128-CBC, the key can
// be selected with the index argument from one previously set with SetKey().
func (hw *Dcp) Encrypt(buf []byte, index int, iv []byte) (err error) {
	return hw.cipher(buf, index, iv, true)
}

// Decrypt performs in-place buffer decryption using AES-128-CBC, the key can
// be selected with the index argument from one previously set with SetKey().
func (hw *Dcp) Decrypt(buf []byte, index int, iv []byte) (err error) {
	return hw.cipher(buf, index, iv, false)
}

// CipherChain performs chained in-place buffer encryption/decryption using
// AES-128-CBC, the key can be selected with the index argument from one
// previously set with SetKey().
//
// The function expects a byte array with concatenated input data and a byte
// array with concatenated initialization vectors, the count and size arguments
// should reflect the number of slices, each to be ciphered and with the
// corresponding initialization vector slice.
func (hw *Dcp) CipherChain(buf []byte, ivs []byte, count int, size int, index int, enc bool) (err error) {
	if len(buf) != size*count || len(buf)%aes.BlockSize != 0 {
		return errors.New("invalid input size")
	}

	if len(ivs) != aes.BlockSize*count {
		return errors.New("invalid IV size")
	}

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	src := dma.Alloc(buf, aes.BlockSize)
	defer dma.Free(src)

	payloads := dma.Alloc(ivs, 0)
	defer dma.Free(payloads)

	pkts, pktBuf := dma.Reserve(WorkPacketLength*count, 0)
	defer dma.Release(pkts)

	pkt := &WorkPacket{}
	pkt.SetDefaults()
	pkt.Control0 |= 1 << DCP_CTRL0_CHAIN
	pkt.BufferSize = uint32(size)

	if enc {
		pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	}

	// use key RAM slot
	pkt.Control1 |= (uint32(index) & 0xff) << DCP_CTRL1_KEY_SELECT

	for i := 0; i < count; i++ {
		pkt.SourceBufferAddress = src + uint32(i*size)
		pkt.DestinationBufferAddress = pkt.SourceBufferAddress
		pkt.PayloadPointer = payloads + uint32(i*aes.BlockSize)

		if i < count-1 {
			pkt.NextCmdAddr = pkts + uint32((i+1)*WorkPacketLength)
		} else {
			bits.Clear(&pkt.Control0, DCP_CTRL0_CHAIN)
			bits.Set(&pkt.Control0, DCP_CTRL0_INTERRUPT_ENABL)
		}

		copy(pktBuf[i*WorkPacketLength:], pkt.Bytes())
	}

	err = hw.cmd(pkts, count)

	if err != nil {
		return
	}

	dma.Read(src, 0, buf)

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

	return append(buf, padding...)
}
