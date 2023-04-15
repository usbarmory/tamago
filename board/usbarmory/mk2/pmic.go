// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mk2

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/iomuxc"
)

// On the USB armory Mk II the PMIC watchdog input (WDI) is connected to the
// SoC external reset source (WDOG2_WDOG_RST_B_DEB) through
// ENET1_TX_EN/KPP_COL2 (p4081, Table 59-1. WDOG External Signals, IMX6ULLRM).
const (
	IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_EN = 0x020e00d8
	IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_EN = 0x020e0364

	WDOG2_WDOG_RST_B_DEB_MODE = 8
)

func init() {
	var ctl uint32

	bits.Set(&ctl, iomuxc.SW_PAD_CTL_HYS)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_PUE)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_PKE)

	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_PUS, 0b11, iomuxc.SW_PAD_CTL_PUS_PULL_UP_22K)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_SPEED, 0b11, iomuxc.SW_PAD_CTL_SPEED_50MHZ)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_DSE, 0b111, iomuxc.SW_PAD_CTL_DSE_2_R0_6)

	p := &iomuxc.Pad{
		Mux: IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_EN,
		Pad: IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_EN,
	}

	p.Mode(WDOG2_WDOG_RST_B_DEB_MODE)
	p.Ctl(ctl)
}

// Reset asserts the PMIC watchdog signal (through the SoC external reset
// source) causing the USB armory Mk II board to power cycle (cold reset).
func Reset() {
	for {
		imx6ul.WDOG2.SoftwareReset()
	}
}
