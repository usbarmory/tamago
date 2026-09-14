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

func (descriptor *admaDescriptor) init(address uint, size int) {
	for size > 0 {
		length := min(size, admaMaxLength)

		descriptor.Attribute = admaValid | admaTransfer
		descriptor.Length = uint16(length)
		descriptor.Address = uint32(address)

		if length == size {
			descriptor.Attribute |= admaEnd
			return
		}

		address += uint(length)
		size -= length
		descriptor.next = &admaDescriptor{}
		descriptor = descriptor.next
	}
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

func admaTable(address uint, size int) []byte {
	if size <= 0 {
		return nil
	}

	descriptor := &admaDescriptor{}
	descriptor.init(address, size)
	return descriptor.Bytes()
}
