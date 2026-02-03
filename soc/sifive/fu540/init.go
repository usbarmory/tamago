// SiFive FU540 initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fu540

import (
	_ "unsafe"
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime.hwinit1).
func Init() {
	RV64.Init()
}

//go:linkname nanotime1 runtime/goos.Nanotime
func nanotime1() int64 {
	return CLINT.Nanotime()
}
