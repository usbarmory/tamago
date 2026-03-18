// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import (
	_ "unsafe"
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint32 = 0x100

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime/goos.Hwinit1).
func Init() {
	ARM64.Init()
	ARM64.EnableCache()

	initTimers()
}
