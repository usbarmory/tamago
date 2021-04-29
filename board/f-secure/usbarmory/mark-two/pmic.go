// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbarmory

import (
	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

// On the USB armory Mk II the PMIC watchdog input (WDI) is connected to the
// SoC external reset source (WDOG2_WDOG_RST_B_DEB) through
// ENET1_TX_EN/KPP_COL2.
const (
	IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_EN = 0x020e00d8
	IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_EN = 0x020e0364
	IOMUXC_ENET1_TX_EN_SELECT_INPUT   = 0x020e066c

	WDOG2_WDOG_RST_B_DEB_MODE = 8
)

func init() {
	var ctl uint32

	bits.Set(&ctl, imx6.SW_PAD_CTL_HYS)
	bits.Set(&ctl, imx6.SW_PAD_CTL_PUE)
	bits.Set(&ctl, imx6.SW_PAD_CTL_PKE)

	bits.SetN(&ctl, imx6.SW_PAD_CTL_PUS, 0b11, imx6.SW_PAD_CTL_PUS_PULL_UP_22K)
	bits.SetN(&ctl, imx6.SW_PAD_CTL_SPEED, 0b11, imx6.SW_PAD_CTL_SPEED_50MHZ)
	bits.SetN(&ctl, imx6.SW_PAD_CTL_DSE, 0b111, imx6.SW_PAD_CTL_DSE_2_R0_6)

	p, err := imx6.NewPad(IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_EN,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_EN,
		IOMUXC_ENET1_TX_EN_SELECT_INPUT)

	if err != nil {
		panic(err)
	}

	p.Mode(WDOG2_WDOG_RST_B_DEB_MODE)
	p.Ctl(ctl)
}

// Reset deasserts the PMIC watchdog signal (through the SoC external reset
// source) causing the USB armory Mk II board to power cycle (cold reset).
func Reset() {
	// enable software reset extension
	reg.Set16(imx6.WDOG2_WCR, imx6.WCR_SRE)

	// assert system reset signal
	reg.Clear16(imx6.WDOG2_WCR, imx6.WCR_SRS)
}
