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
	"github.com/usbarmory/tamago/dma"
)

// SetHashDefaults initializes default values for a DCP work packet that
// performs hash operation.
func (pkt *WorkPacket) SetHashDefaults() {
	pkt.Control0 |= 1 << DCP_CTRL0_INTERRUPT_ENABL
	pkt.Control0 |= 1 << DCP_CTRL0_DECR_SEMAPHORE
	pkt.Control0 |= 1 << DCP_CTRL0_ENABLE_HASH

	pkt.Control1 |= HASH_SELECT_SHA256 << DCP_CTRL1_HASH_SELECT
}

func (hw *DCP) hash(buf []byte, mode int, size int, init bool, term bool) (sum []byte, err error) {
	sourceBufferAddress := dma.Alloc(buf, 4)
	defer dma.Free(sourceBufferAddress)

	pkt := &WorkPacket{}
	pkt.SetHashDefaults()
	pkt.SourceBufferAddress = uint32(sourceBufferAddress)
	pkt.BufferSize = uint32(len(buf))

	if init {
		pkt.Control0 |= 1 << DCP_CTRL0_HASH_INIT
	}

	if term {
		sum = make([]byte, size)

		payloadPointer := dma.Alloc(sum, 4)
		defer dma.Free(payloadPointer)

		pkt.Control0 |= 1 << DCP_CTRL0_HASH_TERM
		pkt.PayloadPointer = uint32(payloadPointer)

		defer func() {
			dma.Read(payloadPointer, 0, sum)

			for i, j := 0, len(sum)-1; i < j; i, j = i+1, j-1 {
				sum[i], sum[j] = sum[j], sum[i]
			}
		}()
	}

	pkt.Control1 |= uint32(mode) << DCP_CTRL1_HASH_SELECT

	ptr := dma.Alloc(pkt.Bytes(), 4)
	defer dma.Free(ptr)

	err = hw.cmd(ptr, 1)

	return
}
