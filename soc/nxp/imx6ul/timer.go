// NXP i.MX6UL timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6ul

import (
	_ "unsafe"
)

// Timer registers (p178, Table 2-3, IMX6ULLRM)
const SYS_CNT_BASE = 0x021dc000

func initTimers() {
	if !Native {
		// use QEMU fixed CNTFRQ value (62.5MHz)
		ARM.InitGenericTimers(0, 62500000)
	} else {
		// U-Boot value for i.MX6 family (8.0MHz)
		ARM.InitGenericTimers(SYS_CNT_BASE, 8000000)
	}
}

//go:linkname nanotime1 runtime/goos.Nanotime
func nanotime1() int64 {
	return ARM.GetTime()
}
