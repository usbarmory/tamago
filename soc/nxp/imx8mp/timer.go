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

// CNTFID0 (p215, 4.11.4.1.6.4, IMX8MPRM)
const CNTFID0_FREQ = REF_FREQ / 3

func initTimers() {
	if !Native {
		ARM64.InitGenericTimers(0, CNTFID0_FREQ)
	} else {
		ARM64.InitGenericTimers(SYS_CNT_BASE, CNTFID0_FREQ)
	}
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return ARM64.GetTime()
}
