// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pci

import (
	"fmt"

	"github.com/usbarmory/tamago/bits"
)

// Base Address Register fields
// (PCI Local Bus Specification 3.0 - 6.2.5.1 Address Maps).
const (
	BAR_ADDR         = 4 // bits 31:4 (memory BARs)
	BAR_PREFETCHABLE = 3
	BAR_TYPE         = 1 // bits 2:1 (memory BARs)
	BAR_IO           = 0

	BAR_TYPE_32 = 0b00
	BAR_TYPE_64 = 0b10
)

func (d *Device) bar(n int) (off, addr uint32, err error) {
	if n < 0 || n > 5 {
		return 0, 0, fmt.Errorf("invalid BAR%d", n)
	}

	off = Bar0 + uint32(n)*4
	addr = d.Read(0, off)

	return
}

// BaseAddress returns a device Base Address register (BAR).
func (d *Device) BaseAddress(n int) uint {
	off, bar, err := d.bar(n)

	if err != nil {
		return 0
	}

	switch bits.GetN(&bar, BAR_TYPE, 0b11) {
	case BAR_TYPE_32:
		return uint(bar)
	case BAR_TYPE_64:
		return uint(d.Read(0, off+4))<<32 | uint(bar)&0xfffffff0
	}

	return 0
}

// MemoryBAR returns the address and size of a memory Base Address Register.
func (d *Device) MemoryBAR(n int) (addr uint64, size uint64, err error) {
	off, bar, err := d.bar(n)

	if err != nil {
		return
	}

	if bits.Get(&bar, BAR_IO) {
		return 0, 0, fmt.Errorf("BAR%d is not a memory BAR", n)
	}

	is64 := bits.GetN(&bar, BAR_TYPE, 0b11) == BAR_TYPE_64

	if is64 && n == 5 {
		return 0, 0, fmt.Errorf("invalid 64-bit BAR%d", n)
	}

	cmd := d.Read(0, Command) & 0xffff
	defer d.Write(0, Command, cmd)

	d.SetCommand(COMMAND_MEMORY_SPACE, false)

	// size by writing all ones and reading back the address bits

	d.Write(0, off, 0xffffffff)
	val := d.Read(0, off)

	mask := uint64(bits.GetN(&val, BAR_ADDR, 0xfffffff) << BAR_ADDR)
	d.Write(0, off, bar)

	if is64 {
		hi := d.Read(0, off+4)

		d.Write(0, off+4, 0xffffffff)
		mask |= uint64(d.Read(0, off+4)) << 32

		d.Write(0, off+4, hi)
		addr = uint64(hi) << 32
	} else {
		mask |= 0xffffffff00000000
	}

	addr |= uint64(bits.GetN(&bar, BAR_ADDR, 0xfffffff) << BAR_ADDR)
	size = ^mask + 1

	if addr == 0 || mask == 0 {
		return 0, 0, fmt.Errorf("BAR%d not assigned", n)
	}

	return
}
