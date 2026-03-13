// Custom GOOS support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !tiny

package goos

// Required constants.
const (
	// LogHeapArenaBytes defines the size of a runtime heap arena in log_2
	// bytes (see runtime.logHeapArenaBytes).
	LogHeapArenaBytes = (2+20)

	// LogPallocChunkPages defines the size of a runtime bitmap chunk in
	// log_2 bytes (see runtime.logPallocChunkPages).
	LogPallocChunkPages = 9

	// MinPhysPageSize is a lower-bound on the physical page size (see
	// runtime.minPhysPageSize).
	MinPhysPageSize = 4096
)
