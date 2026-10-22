// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
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
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	maxBuses   = 256
	maxDevices = 32
)

// Header Type 0x0 offsets
const (
	VendorID           = 0x00
	Command            = 0x04
	RevisionID         = 0x08
	Bar0               = 0x10
	CapabilitiesOffset = 0x34
)

const (
	CONFIG_ADDRESS = 0x0cf8
	CONFIG_DATA    = 0x0cfc
)

// Command register fields
// (PCI Local Bus Specification 3.0 - 6.2.2 Device Control).
const (
	COMMAND_INTERRUPT_DISABLE = 10
	COMMAND_BUS_MASTER        = 2
	COMMAND_MEMORY_SPACE      = 1
	COMMAND_IO_SPACE          = 0
)

// Device represents a PCI device.
type Device struct {
	// Bus number
	Bus uint32
	// Vendor ID
	Vendor uint16
	// Device ID
	Device uint16
	// Revision ID
	Revision uint8
	// Prog IF
	ProgIF uint8
	// Subclass
	Subclass uint8
	// Class Code
	ClassCode uint8

	// PCI Slot
	Slot uint32
}

func (d *Device) address(fn uint32, off uint32) uint32 {
	return 1<<31 | d.Bus<<16 | d.Slot<<11 | fn<<8 | off&0xfc
}

// Read reads the device configuration space for a given function and
// register offset.
func (d *Device) Read(fn uint32, off uint32) uint32 {
	reg.Out32(CONFIG_ADDRESS, d.address(fn, off))
	return reg.In32(CONFIG_DATA) >> ((off & 2) * 8)
}

// Write writes the device configuration space for a given function and
// register offset, the offset must be 32-bit aligned.
func (d *Device) Write(fn uint32, off uint32, val uint32) {
	if (off&2)*8 != 0 {
		return
	}

	reg.Out32(CONFIG_ADDRESS, d.address(fn, off))
	reg.Out32(CONFIG_DATA, val)
}

func (d *Device) probe() bool {
	if d.Bus > maxBuses {
		return false
	}

	val := d.Read(0, VendorID)

	if d.Vendor = uint16(val); d.Vendor == 0xffff {
		return false
	}

	d.Device = uint16(val >> 16)

	val = d.Read(0, RevisionID)

	d.Revision = uint8(val)
	d.ProgIF = uint8(val >> 8)
	d.Subclass = uint8(val >> 16)
	d.ClassCode = uint8(val >> 24)

	return true
}

// Probe probes a PCI device by vendor and device identifiers.
func Probe(bus int, vendor uint16, device uint16) *Device {
	d := &Device{
		Bus: uint32(bus),
	}

	for slot := range uint32(maxDevices) {
		d.Slot = slot

		if d.probe() && d.Vendor == vendor && d.Device == device {
			return d
		}
	}

	return nil
}

// ProbeFunction probes the nth (starting from 0) PCI device matching the
// argument functions.
func ProbeFunction(bus int, n int, class uint8, subclass uint8, progIF uint8) *Device {
	var i int

	d := &Device{
		Bus: uint32(bus),
	}

	for slot := range uint32(maxDevices) {
		d.Slot = slot

		if d.probe() && d.ClassCode == class && d.Subclass == subclass && d.ProgIF == progIF {
			if n == i {
				return d
			}

			i++
		}
	}

	return nil
}

// Devices returns all found PCI devices on a given bus.
func Devices(bus int) (devices []*Device) {
	for slot := range uint32(maxDevices) {
		d := &Device{
			Bus:  uint32(bus),
			Slot: slot,
		}

		if d.probe() {
			devices = append(devices, d)
		}
	}

	return
}

// SetCommand sets or clears a Command register bit for function 0.
func (d *Device) SetCommand(pos int, val bool) {
	cmd := d.Read(0, Command) & 0xffff
	bits.SetTo(&cmd, pos, val)
	d.Write(0, Command, cmd)
}
