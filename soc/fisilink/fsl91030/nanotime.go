// Fisilink FSL91030 runtime nanotime hook
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// When built without the linknanotime tag this file provides the default
// runtime/goos.Nanotime implementation, reading the CLINT mtime register.
//
// Building with -tags linknanotime excludes this file, allowing a board
// package (e.g. milkv-vega-qemu) to supply its own nanotime via
// go:linkname. This is the standard TamaGo override pattern used by every
// SoC package that supports board-level timer overrides.
//
//go:build !linknanotime

package fsl91030

import (
	_ "unsafe"
)

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return RV64.GetTime()
}
