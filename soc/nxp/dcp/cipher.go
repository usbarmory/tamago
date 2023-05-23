// NXP Data Co-Processor (DCP) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dcp

import (
	"crypto/aes"
	"errors"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
)

// SetCipherDefaults initializes default values for a DCP work packet that
// performs cipher operation.
func (pkt *WorkPacket) SetCipherDefaults() {
	pkt.Control0 |= 1 << DCP_CTRL0_INTERRUPT_ENABL
	pkt.Control0 |= 1 << DCP_CTRL0_DECR_SEMAPHORE
	pkt.Control0 |= 1 << DCP_CTRL0_ENABLE_CIPHER
	pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_INIT

	pkt.Control1 |= CIPHER_SELECT_AES128 << DCP_CTRL1_CIPHER_SELECT
	pkt.Control1 |= CIPHER_MODE_CBC << DCP_CTRL1_CIPHER_MODE
}

func (hw *DCP) cipher(buf []byte, index int, iv []byte, enc bool) (err error) {
	if len(buf)%aes.BlockSize != 0 {
		return errors.New("invalid input size")
	}

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	if len(iv) != aes.BlockSize {
		return errors.New("invalid IV size")
	}

	sourceBufferAddress := dma.Alloc(buf, aes.BlockSize)
	defer dma.Free(sourceBufferAddress)

	payloadPointer := dma.Alloc(iv, 4)
	defer dma.Free(payloadPointer)

	pkt := &WorkPacket{}
	pkt.SetCipherDefaults()

	if enc {
		pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	}

	// use key RAM slot
	pkt.Control1 |= (uint32(index) & 0xff) << DCP_CTRL1_KEY_SELECT
	pkt.SourceBufferAddress = uint32(sourceBufferAddress)
	pkt.DestinationBufferAddress = pkt.SourceBufferAddress
	pkt.BufferSize = uint32(len(buf))
	pkt.PayloadPointer = uint32(payloadPointer)

	ptr := dma.Alloc(pkt.Bytes(), 4)
	defer dma.Free(ptr)

	if err = hw.cmd(ptr, 1); err != nil {
		return
	}

	dma.Read(sourceBufferAddress, 0, buf)

	return
}

// Encrypt performs in-place buffer encryption using AES-128-CBC, the key can
// be selected with the index argument from one previously set with SetKey().
func (hw *DCP) Encrypt(buf []byte, index int, iv []byte) (err error) {
	return hw.cipher(buf, index, iv, true)
}

// Decrypt performs in-place buffer decryption using AES-128-CBC, the key can
// be selected with the index argument from one previously set with SetKey().
func (hw *DCP) Decrypt(buf []byte, index int, iv []byte) (err error) {
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
func (hw *DCP) CipherChain(buf []byte, ivs []byte, count int, size int, index int, enc bool) (err error) {
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

	payloads := dma.Alloc(ivs, 4)
	defer dma.Free(payloads)

	pkts, pktBuf := dma.Reserve(WorkPacketLength*count, 4)
	defer dma.Release(pkts)

	pkt := &WorkPacket{}
	pkt.SetCipherDefaults()
	pkt.Control0 |= 1 << DCP_CTRL0_CHAIN
	pkt.BufferSize = uint32(size)

	bits.Clear(&pkt.Control0, DCP_CTRL0_INTERRUPT_ENABL)

	if enc {
		pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	}

	// use key RAM slot
	pkt.Control1 |= (uint32(index) & 0xff) << DCP_CTRL1_KEY_SELECT

	for i := 0; i < count; i++ {
		pkt.SourceBufferAddress = uint32(src) + uint32(i*size)
		pkt.DestinationBufferAddress = pkt.SourceBufferAddress
		pkt.PayloadPointer = uint32(payloads) + uint32(i*aes.BlockSize)

		if i < count-1 {
			pkt.NextCmdAddr = uint32(pkts) + uint32((i+1)*WorkPacketLength)
		} else {
			bits.Clear(&pkt.Control0, DCP_CTRL0_CHAIN)
			bits.Set(&pkt.Control0, DCP_CTRL0_INTERRUPT_ENABL)
		}

		copy(pktBuf[i*WorkPacketLength:], pkt.Bytes())
	}

	if err = hw.cmd(pkts, count); err != nil {
		return
	}

	dma.Read(src, 0, buf)

	return
}
