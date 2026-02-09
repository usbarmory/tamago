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
// setup (e.g. runtime/goos.Hwinit1).
func Init() {
	RV64.Init()
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return CLINT.Nanotime()
}
