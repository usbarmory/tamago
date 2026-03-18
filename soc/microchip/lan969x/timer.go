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

// SYS_CNT CTRL (p500 TABLE 3-326, DS00005048E)
const SYS_CNT_BASE = 0xe8000000

func initTimers() {
	// Use default frequency, set in bootloaders
	ARM64.InitGenericTimers(SYS_CNT_BASE, 0)
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return ARM64.GetTime()
}
