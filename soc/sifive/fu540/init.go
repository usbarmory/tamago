// SiFive FU540 initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fu540

import (
	_ "unsafe"
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint64 = 0x100

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup (e.g. runtime.hwinit).
func Init() {
	RV64.Init()
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return CLINT.Nanotime()
}
