// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// CPUIDFunction represents an AMD SEV-SNP CPUID_FUNCTION
// (SEV Secure Nested Paging Firmware ABI Specification - Table 16).
type CPUIDFunction struct {
	// Input parameters
	EAX_IN  uint32
	ECX_IN  uint32
	XCR0_IN uint64
	XSS_IN  uint64

	// Output parameters
	EAX uint32
	EBX uint32
	ECX uint32
	EDX uint32

	_ uint64
}

// CPUIDPage represents an AMD SEV-SNP CPUID Page
// (SEV Secure Nested Paging Firmware ABI Specification - Table 72).
type CPUIDPage struct {
	Count    uint32
	_        uint32
	_        uint64
	Function [64]CPUIDFunction
}

// Init initializes a CPUID Page instance deoding the contents of the argument
// memory location.
func (c *CPUIDPage) Init(addr uint, size int) (err error) {
	if addr == 0 {
		return errors.New("invalid address")
	}

	if size < pageSize {
		return errors.New("invalid size")
	}

	r, err := dma.NewRegion(addr, size, false)

	if err != nil {
		return
	}

	_, buf := r.Reserve(size, 0)

	if _, err = binary.Decode(buf, binary.LittleEndian, c); err != nil {
		return
	}

	return
}

// CPUID returns the processor capabilities as found in the AMD SEV-SNP CPUID
// Page, note that EAX_IN and ECX_IN are matched exactly against leaf and
// subleaf.
func (c *CPUIDPage) CPUID(leaf, subleaf uint32) (eax, ebx, ecx, edx uint32) {
	if c.Count == 0 || c.Count > 64 {
		return
	}

	for _, fn := range c.Function[0:c.Count] {
		if fn.EAX_IN == leaf && fn.ECX_IN == subleaf {
			return fn.EAX, fn.EBX, fn.ECX, fn.EDX
		}
	}

	return
}
