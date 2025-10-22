// NXP i.MX8MP configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx8mp

// Clock registers
const (
	CCM_ANALOG_DIGPROG = 0x30360800
	CCM_CCGR_BASE      = 0x30384000

	CCM_CCGR34 = CCM_CCGR_BASE + 0x4220 
	CCM_CCGR73 = CCM_CCGR_BASE + 0x4490
	CCM_CCGR74 = CCM_CCGR_BASE + 0x44a0
	CCM_CCGR83 = CCM_CCGR_BASE + 0x4530
	CCM_CCGR84 = CCM_CCGR_BASE + 0x4540
	CCM_CCGR85 = CCM_CCGR_BASE + 0x4550
)
