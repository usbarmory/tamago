// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package dma provides primitives for direct memory allocation and alignment,
// it is primarily used in bare metal device driver operation to avoid passing
// Go pointers for DMA purposes.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package dma

import (
	"fmt"
	"runtime"
)

// NewRegion initializes a memory region for DMA buffer allocation.
//
// To avoid unforseen consequences the caller must ensure that allocated
// regions do not overlap among themselves or with the global one (see Init()).
//
// To allow allocation of DMA buffers within Go runtime memory the unsafe flag
// must be set.
func NewRegion(addr uint, size int, unsafe bool) (r *Region, err error) {
	start := uint(addr)
	end := uint(start) + uint(size)

	// returns uint32/uint64 depending on platform
	rs, re := runtime.MemRegion()
	ramStart := uint(rs)
	ramEnd := uint(re)

	if !unsafe &&
		(ramStart >= start && ramStart < end ||
			ramEnd > start && ramEnd < end ||
			start >= ramStart && end < ramEnd) {
		return nil, fmt.Errorf("DMA within Go runtime memory (%#x-%#x) is not allowed", ramStart, ramEnd)
	}

	r = &Region{}
	r.Init(start, uint(size))

	return
}

// Init initializes the global memory region for DMA buffer allocation, used
// throughout the tamago package for all DMA allocations.
//
// Additional DMA regions for application use can be allocated through
// NewRegion().
func Init(start uint, size int) (err error) {
	dma, err = NewRegion(start, size, false)
	return
}

// Reserve is the equivalent of Region.Reserve() on the global DMA region.
func Reserve(size int, align int) (addr uint, buf []byte) {
	return dma.Reserve(size, align)
}

// Reserved is the equivalent of Region.Reserved() on the global DMA region.
func Reserved(buf []byte) (res bool, addr uint) {
	return dma.Reserved(buf)
}

// Alloc is the equivalent of Region.Alloc() on the global DMA region.
func Alloc(buf []byte, align int) (addr uint) {
	return dma.Alloc(buf, align)
}

// Read is the equivalent of Region.Read() on the global DMA region.
func Read(addr uint, off int, buf []byte) {
	dma.Read(addr, off, buf)
}

// Write is the equivalent of Region.Write() on the global DMA region.
func Write(addr uint, off int, buf []byte) {
	dma.Write(addr, off, buf)
}

// Free is the equivalent of Region.Free() on the global DMA region.
func Free(addr uint) {
	dma.Free(addr)
}

// Release is the equivalent of Region.Release() on the global DMA region.
func Release(addr uint) {
	dma.Release(addr)
}
