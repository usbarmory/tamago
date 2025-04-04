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
)

// Capability IDs
//
// (PCI Code and ID Assignment Specification Revision 1.11
// 24 Jan 2019 - 2. Capability IDs).
const (
	Null           = 0x00
	Power          = 0x01
	AGP            = 0x02
	VPD            = 0x03
	SlotID         = 0x04
	MSI            = 0x05
	HotSwap        = 0x06
	PCIX           = 0x07
	HyperTransport = 0x08
	VendorSpecific = 0x09
	Debug          = 0x0a
	CompactPCI     = 0x0b
	HotPlug        = 0x0c
	Bridge         = 0x0d
	AGP8x          = 0x0e
	Secure         = 0x0f
	PCIe           = 0x10
	MSIX           = 0x11
	SATA           = 0x12
	AF             = 0x13
	EA             = 0x14
	FPB            = 0x15
)

// CapabilityHeader represents the common fields of PCI Capabilities entries.
type CapabilityHeader struct {
	Vendor uint8
	Next   uint8
}

// Unmarshal decodes a PCI Capability common fields from the argument device
// configuration space at function 0 and the given register offset.
func (hdr *CapabilityHeader) Unmarshal(d *Device, off uint32) (err error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, d.Read(0, off))
	_, err = binary.Decode(buf, binary.LittleEndian, hdr)
	return
}

// Capabilities is an iterator over the entries of the device Capabilities
// List.
func (d *Device) Capabilities() func(func(off uint32, hdr *CapabilityHeader) bool) {
	return func(yield func(uint32, *CapabilityHeader) bool) {
		off := d.Read(0, CapabilitiesOffset)

		for off != 0 {
			hdr := &CapabilityHeader{}

			if err := hdr.Unmarshal(d, off); err != nil {
				return
			}

			if !yield(off, hdr) {
				return
			}

			off = uint32(hdr.Next)
		}
	}
}
