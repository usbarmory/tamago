// NXP i.MX8MP configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx8mp

import (
	_ "unsafe"
)

// Timer registers (p31, Table 2-4, IMX8MPRM)
const SYS_CNT_BASE = 0x306a0000

func initTimers() {
	if !Native {
		// use QEMU fixed CNTFRQ value (62.5MHz)
		ARM64.InitGenericTimers(0, 62500000)
	} else {
		// U-Boot value for i.MX6 family (8.0MHz)
		ARM64.InitGenericTimers(SYS_CNT_BASE, 8000000)
	}
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return ARM64.GetTime()
}
