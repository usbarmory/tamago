// Microchip Secure Digital Host Controller Interface support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sdhci

import (
	"bytes"
	"encoding/binary"
)

const (
	// ADMA2 data and descriptor addresses are word aligned.
	admaAlignment = 4

	admaDescriptorSize = 8
	// Largest non-zero, word-aligned value in the 16-bit length field.
	admaMaxLength = 65532

	// Largest partial cache line bounced at either end of a direct read.
	admaMaxEdge = 128

	admaValid    = 1 << 0
	admaEnd      = 1 << 1
	admaTransfer = 0b10 << 4
)

type admaDescriptor struct {
	Attribute uint16
	Length    uint16
	Address   uint32

	next *admaDescriptor
}

type admaSegment struct {
	address uint
	size    int
}

// Bytes converts the descriptor chain to its ADMA2 byte representation.
func (descriptor *admaDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	for descriptor != nil {
		binary.Write(buf, binary.LittleEndian, descriptor.Attribute)
		binary.Write(buf, binary.LittleEndian, descriptor.Length)
		binary.Write(buf, binary.LittleEndian, descriptor.Address)
		descriptor = descriptor.next
	}

	return buf.Bytes()
}

func admaTableSize(size int) int {
	return (size + admaMaxLength - 1) / admaMaxLength * admaDescriptorSize
}

func admaTable(segments ...admaSegment) []byte {
	var first, last *admaDescriptor

	for _, segment := range segments {
		address, size := segment.address, segment.size

		for size > 0 {
			length := min(size, admaMaxLength)
			descriptor := &admaDescriptor{
				Attribute: admaValid | admaTransfer,
				Length:    uint16(length),
				Address:   uint32(address),
			}

			if last == nil {
				first = descriptor
			} else {
				last.next = descriptor
			}

			last = descriptor
			address += uint(length)
			size -= length
		}
	}

	if first == nil {
		return nil
	}

	last.Attribute |= admaEnd

	return first.Bytes()
}
