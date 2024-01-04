// NXP i.MX6UL timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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
	switch Family {
	case IMX6UL, IMX6ULL:
		if !Native {
			// use QEMU fixed CNTFRQ value (62.5MHz)
			ARM.InitGenericTimers(0, 62500000)
		} else {
			// U-Boot value for i.MX6 family (8.0MHz)
			ARM.InitGenericTimers(SYS_CNT_BASE, 8000000)
		}
	default:
		ARM.InitGlobalTimers()
	}
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(ARM.TimerFn()*ARM.TimerMultiplier + ARM.TimerOffset)
}
