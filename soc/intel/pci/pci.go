// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pci implements a driver for Intel Peripheral Component Interconnect
// (PCI) controllers adopting the following reference
// specifications:
//   - PCI Local Bus Specification, revision 3.0, PCI Special Interest Group
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package pci

import (
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	CONFIG_ADDRESS = 0x0cf8
	CONFIG_DATA    = 0x0cfc
)

// Device represents a PCI device.
type Device struct {
	// Bus number
	Bus uint32
	// Vendor ID
	Vendor uint16
	// Device ID
	Device uint16

	// PCI Slot
	Slot uint32
	// Base Address #0 (BAR0)
	BaseAddress0 uint32
}

func (d *Device) read(bus uint32, slot uint32, fn uint32, offset uint32) uint32 {
	address := (bus << 16) | (slot << 11) | (fn << 8) | (offset & 0xfc) | 0x80000000

	reg.Out32(CONFIG_ADDRESS, address)
	return reg.In32(CONFIG_DATA) >> ((offset & 2) * 8)
}

// Probe probes a PCI device based on its Bus, Vendor and Device fields, on
// success the remaining fields are populated.
func (d *Device) Probe() (found bool) {
	for slot := uint32(0); slot <= 31; slot++ {
		val := d.read(d.Bus, slot, 0, 0)

		vendor := uint16(val)
		device := uint16(val >> 16)

		if vendor == 0xfff || vendor != d.Vendor || device != d.Device {
			continue
		}

		d.BaseAddress0 = d.read(d.Bus, slot, 0, 0x10)
		d.BaseAddress0 &= 0xfffffffc

		return true
	}

	return false
}
