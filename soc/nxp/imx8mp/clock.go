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

	CCM_CCGR10 = CCM_CCGR_BASE + 0x00a0
	CCM_CCGR34 = CCM_CCGR_BASE + 0x0220
	CCM_CCGR71 = CCM_CCGR_BASE + 0x0470
	CCM_CCGR73 = CCM_CCGR_BASE + 0x0490
	CCM_CCGR74 = CCM_CCGR_BASE + 0x04a0
	CCM_CCGR83 = CCM_CCGR_BASE + 0x0530
	CCM_CCGR84 = CCM_CCGR_BASE + 0x0540
	CCM_CCGR85 = CCM_CCGR_BASE + 0x0550
)

// Clocks at boot time
// (p749, Table 6-4. Clock root setting by ROM, IMX8MPRM)
const (
	ARM_FREQ = 1000000000 // 1GHz
	AHB_FREQ = 133000000  // 133MHz
	REF_FREQ =  24000000  // 24MHz
)

// ARMFreq returns the ARM core frequency.
func ARMFreq() (hz uint32) {
	return ARM_FREQ // TODO: default is assumed
}

// GetPeripheralClock returns the IPG_CLK_ROOT frequency,
// (p253, Figure 5-3. CCM Clock Tree Root Slics, IMX8MPRM).
func GetPeripheralClock() uint32 {
	// IPG_CLK_ROOT derived from AHB_CLK_ROOT which is 133 MHz
	return AHB_FREQ // TODO: default is assumed
}

// GetUARTClock returns the UART_CLK_ROOT frequency,
// (p253, Figure 5-3. CCM Clock Tree Root Slics, IMX8MPRM).
func GetUARTClock() uint32 {
	return REF_FREQ // TODO: default is assumed
}
