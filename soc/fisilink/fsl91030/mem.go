// Fisilink FSL91030 memory layout
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package fsl91030

import (
	_ "unsafe"
)

// FSL91030 DRAM starts at 0x41000000 (240 MB available)
// From device tree: nuclei_ux600fd.dts

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = 0x41000000
