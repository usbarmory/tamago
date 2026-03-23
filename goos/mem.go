// Custom GOOS support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !tiny

package goos

const bits = 32 << (^uint(0) >> 63) / 8

// Required constants.
const (
	// ArenaBaseOffset is the pointer value that corresponds to index 0 in
	// the heap arena map (see runtime.arenaBaseOffset).
	ArenaBaseOffset = 0

	// HeapAddrBits defines the number of bits in a heap address (see
	// runtime.heapAddrBits).
	HeapAddrBits = (8-bits)*3 + bits*5 // 32-bit:32 64-bit:40

	// LogHeapArenaBytes defines the size of a runtime heap arena in log_2
	// bytes (see runtime.logHeapArenaBytes).
	LogHeapArenaBytes = (2 + 20)

	// LogPallocChunkPages defines the size of a runtime bitmap chunk in
	// log_2 bytes (see runtime.logPallocChunkPages).
	LogPallocChunkPages = 9

	// MinPhysPageSize is a lower-bound on the physical page size (see
	// runtime.minPhysPageSize).
	MinPhysPageSize = 4096

	// StackSystem is a number of additional bytes to add to each stack
	// below the usual guard area.
	StackSystem = 0
)
