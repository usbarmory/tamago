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
	"crypto/aes"
	"errors"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/dma"
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

func cipher(buf []byte, index int, iv []byte, enc bool) (err error) {
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
	pkt.SetCipherDefaults()

	if enc {
		pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	}

	// use key RAM slot
	pkt.Control1 |= (uint32(index) & 0xff) << DCP_CTRL1_KEY_SELECT

	pkt.BufferSize = uint32(len(buf))
	pkt.SourceBufferAddress = dma.Alloc(buf, aes.BlockSize)

	pkt.DestinationBufferAddress = pkt.SourceBufferAddress
	defer dma.Free(pkt.SourceBufferAddress)

	pkt.PayloadPointer = dma.Alloc(iv, 4)
	defer dma.Free(pkt.PayloadPointer)

	ptr := dma.Alloc(pkt.Bytes(), 4)
	defer dma.Free(ptr)

	err = cmd(ptr, 1)

	if err != nil {
		return
	}

	dma.Read(pkt.DestinationBufferAddress, 0, buf)

	return
}

// Encrypt performs in-place buffer encryption using AES-128-CBC, the key can
// be selected with the index argument from one previously set with SetKey().
func Encrypt(buf []byte, index int, iv []byte) (err error) {
	return cipher(buf, index, iv, true)
}

// Decrypt performs in-place buffer decryption using AES-128-CBC, the key can
// be selected with the index argument from one previously set with SetKey().
func Decrypt(buf []byte, index int, iv []byte) (err error) {
	return cipher(buf, index, iv, false)
}

// CipherChain performs chained in-place buffer encryption/decryption using
// AES-128-CBC, the key can be selected with the index argument from one
// previously set with SetKey().
//
// The function expects a byte array with concatenated input data and a byte
// array with concatenated initialization vectors, the count and size arguments
// should reflect the number of slices, each to be ciphered and with the
// corresponding initialization vector slice.
func CipherChain(buf []byte, ivs []byte, count int, size int, index int, enc bool) (err error) {
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

	err = cmd(pkts, count)

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
