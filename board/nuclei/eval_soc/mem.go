// Nuclei EvalSoC emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package eval_soc

import (
	_ "unsafe"
)

// The Nuclei QEMU nuclei_evalsoc machine limits usable RAM to 200 MB to avoid
// an overlap with the emulator internal address decoding; the FSL91030 DRAM
// base (0x41000000) is provided by the SoC package.
//
// Applications can override ramSize with the `linkramsize` build tag.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x0c800000 // 200 MB
