// AI Foundry ET-SoC-1 initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package etsoc1

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/riscv64"
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime/goos.Hwinit1).
func Init() {
	// FIXME
	rv64 := riscv64.CPU{}
	rv64.Init()
}
