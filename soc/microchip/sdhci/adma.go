// Microchip Secure Digital Host Controller Interface support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sdhci

import "encoding/binary"

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

func admaTableSize(size int) int {
	return (size + admaMaxLength - 1) / admaMaxLength * admaDescriptorSize
}

func admaTable(address uint, size int) []byte {
	table := make([]byte, admaTableSize(size))

	for offset := 0; size > 0; offset += admaDescriptorSize {
		length := min(size, admaMaxLength)
		attribute := uint16(admaValid | admaTransfer)

		if length == size {
			attribute |= admaEnd
		}

		binary.LittleEndian.PutUint16(table[offset:], attribute)
		binary.LittleEndian.PutUint16(table[offset+2:], uint16(length))
		binary.LittleEndian.PutUint32(table[offset+4:], uint32(address))

		address += uint(length)
		size -= length
	}

	return table
}
