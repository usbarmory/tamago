// AI Foundry ET-SoC-1 Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package minion

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/aifoundry/etsoc1"
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime/goos.Hwinit1).
func Init() {
	RV64.Init()

	// ET-Minion mtvec must be 4 KB aligned
	alignExceptionHandler()

	riscv64.IPI = etsoc1.IPI
	riscv64.ClearIPI = etsoc1.ClearIPI
}
