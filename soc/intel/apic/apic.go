// Intel Advanced Programmable Interrupt Controller (APIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package apic implements a driver for Intel Local (LAPIC) and I/O (IOAPIC)
// Advanced Programmable Interrupt Controllers adopting the following reference
// specifications:
//   - Intel® 64 and IA-32 Architectures Software Developer’s Manual - Volume 3A - Chapter 10
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package apic

const (
	// LAPIC and IOAPICs supported vectors
	MinVector = 16
	MaxVector = 255
)
