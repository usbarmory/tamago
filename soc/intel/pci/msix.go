// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pci

import (
	"encoding/binary"

	"github.com/usbarmory/tamago/dma"
)

const msixEnable = 31

// CapabilityMSIX represents an MSI-X Capability Structure.
type CapabilityMSIX struct {
	CapabilityHeader
	MessageControl uint16
	TableOffset    uint32
	PBAOffset      uint32

	device *Device
	off    uint32
}

// Unmarshal decodes a PCI Capability common fields from the argument device
// configuration space at function 0 and the given register offset.
func (msix *CapabilityMSIX) Unmarshal(d *Device, off uint32) (err error) {
	val := d.Read(0, off)
	msix.Vendor = uint8(val & 0xff)
	msix.Next = uint8(val >> 8)
	msix.MessageControl = uint16(val >> 16)

	msix.TableOffset = d.Read(0, off+4)
	msix.PBAOffset = d.Read(0, off+8)

	msix.device = d
	msix.off = off

	return
}

// TableSize returns the number of entries in the MSI-X table.
func (msix *CapabilityMSIX) TableSize() int {
	return int(msix.MessageControl & 0x7ff) + 1
}

// EnableInterrupt configures an MSI-X interrupt entry.
func (msix *CapabilityMSIX) EnableInterrupt(n int, addr uint64, data uint32) {
	if n > msix.TableSize() || msix.device == nil {
		return
	}

	bir := int(msix.TableOffset & 0b11)
	bar := uint64(msix.device.BaseAddress(bir))
	table := bar + uint64(msix.TableOffset) & 0xfffffffc

	size := 16
	off := uint64(size * n)

	r, err := dma.NewRegion(uint(table+off), size, false)

	if err != nil {
		return
	}

	ptr, entry := r.Reserve(size, 0)
	defer dma.Release(ptr)

	binary.LittleEndian.PutUint64(entry[0:], addr)
	binary.LittleEndian.PutUint32(entry[8:], data)
	binary.LittleEndian.PutUint32(entry[12:], 0)

	msix.device.Write(0, msix.off, 1 << msixEnable)
}
