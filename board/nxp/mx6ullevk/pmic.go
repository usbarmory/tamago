// MCIMX6ULL-EVK support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mx6ullevk

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/iomuxc"
)

// On the MCIMX6ULL-EVK the SoC LCD_RESET output signal, which is tied to the
// global watchdog, is connected to power reset logic
// (p4081, Table 59-1. WDOG External Signals, IMX6ULLRM).
const (
	IOMUXC_SW_MUX_CTL_PAD_LCD_RESET = 0x020e0114
	IOMUXC_SW_PAD_CTL_PAD_LCD_RESET = 0x020e03a0

	WDOG1_WDOG_ANY_MODE = 4
)

func init() {
	var ctl uint32

	bits.Set(&ctl, iomuxc.SW_PAD_CTL_HYS)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_PUE)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_PKE)

	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_PUS, 0b11, iomuxc.SW_PAD_CTL_PUS_PULL_UP_22K)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_SPEED, 0b11, iomuxc.SW_PAD_CTL_SPEED_50MHZ)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_DSE, 0b111, iomuxc.SW_PAD_CTL_DSE_2_R0_6)

	p := iomuxc.Pad{
		Mux: IOMUXC_SW_MUX_CTL_PAD_LCD_RESET,
		Pad: IOMUXC_SW_PAD_CTL_PAD_LCD_RESET,
	}

	p.Mode(WDOG1_WDOG_ANY_MODE)
	p.Ctl(ctl)
}

// Reset asserts the global watchdog signal which causes the MCIMX6ULL-EVK
// board to power cycle (cold reset).
func Reset() {
	imx6ul.WDOG1.Reset()
}
