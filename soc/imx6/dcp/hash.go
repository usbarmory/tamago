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
	"errors"

	"github.com/f-secure-foundry/tamago/dma"
)

// SetHashDefaults initializes default values for a DCP work packet that
// performs hash operation.
func (pkt *WorkPacket) SetHashDefaults() {
	pkt.Control0 |= 1 << DCP_CTRL0_INTERRUPT_ENABL
	pkt.Control0 |= 1 << DCP_CTRL0_DECR_SEMAPHORE
	pkt.Control0 |= 1 << DCP_CTRL0_ENABLE_HASH

	pkt.Control1 |= HASH_SELECT_SHA256 << DCP_CTRL1_HASH_SELECT
}

func hash(buf []byte, mode uint32, init bool, term bool) (sum []byte, err error) {
	var size int

	pkt := &WorkPacket{}
	pkt.SetHashDefaults()

	pkt.BufferSize = uint32(len(buf))

	pkt.SourceBufferAddress = dma.Alloc(buf, 4)
	defer dma.Free(pkt.SourceBufferAddress)

	switch mode {
	case HASH_SELECT_CRC32:
		size = 4
	case HASH_SELECT_SHA1:
		size = 20
	case HASH_SELECT_SHA256:
		size = 32
	default:
		return nil, errors.New("invalid hash mode")
	}

	if init {
		pkt.Control0 |= 1 << DCP_CTRL0_HASH_INIT
	}

	if term {
		sum = make([]byte, size)

		pkt.PayloadPointer = dma.Alloc(sum, 4)
		defer dma.Free(pkt.PayloadPointer)

		pkt.Control0 |= 1 << DCP_CTRL0_HASH_TERM
	}

	pkt.Control1 |= mode << DCP_CTRL1_HASH_SELECT

	ptr := dma.Alloc(pkt.Bytes(), 4)
	defer dma.Free(ptr)

	err = cmd(ptr, 1)

	if err != nil {
		return
	}

	dma.Read(pkt.PayloadPointer, 0, sum)

	for i, j := 0, len(sum)-1; i < j; i, j = i+1, j-1 {
		sum[i], sum[j] = sum[j], sum[i]
	}

	return
}
