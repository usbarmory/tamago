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
	"errors"

	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/iomuxc"
)

// KSZ8081RNB configuration constants.
//
// On the MCIMX6ULL-EVK the ENET MACs are each connected to an KSZ8081RNB PHY,
// this board package configures them at 100 Mbps / Full-duplex mode.
const (
	KSZ_CTRL    = 0x00
	CTRL_RESET  = 15
	CTRL_SPEED  = 13
	CTRL_DUPLEX = 8

	KSZ_INT = 0x1b

	KSZ_PHYCTRL2  = 0x1f
	CTRL2_HP_MDIX = 15
	CTRL2_RMII    = 7
	CTRL2_LED     = 4
)

const (
	// MUX
	IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_DATA0 = 0x020e00c4
	IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_DATA1 = 0x020e00c8
	IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_EN    = 0x020e00cc
	IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_DATA0 = 0x020e00d0
	IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_DATA1 = 0x020e00d4
	IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_EN    = 0x020e00d8
	IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_CLK   = 0x020e00dc
	IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_ER    = 0x020e00e0
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO06     = 0x020e0074
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07     = 0x020e0078

	// PAD
	IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_DATA0 = 0x020e0350
	IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_DATA1 = 0x020e0354
	IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_EN    = 0x020e0358
	IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_DATA0 = 0x020e035c
	IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_DATA1 = 0x020e0360
	IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_EN    = 0x020e0364
	IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_CLK   = 0x020e0368
	IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_ER    = 0x020e036c
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO06     = 0x020e0300
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07     = 0x020e0304

	// SELECT INPUT
	IOMUXC_ENET1_REF_CLK1_SELECT_INPUT  = 0x020e0574
	IOMUXC_ENET1_MAC0_MDIO_SELECT_INPUT = 0x020e0578

	IOMUX_ALT0 = 0
	IOMUX_ALT4 = 4

	DAISY_ENET1_TX_CLK_ALT4     = 0b10
	DAISY_ENET1_GPIO1_IO06_ALT0 = 0
)

func init() {
	imx6ul.ENET1.EnablePHY = EnablePHY
	imx6ul.ENET2.EnablePHY = EnablePHY

	imx6ul.ENET1.RMII = true
	imx6ul.ENET2.RMII = true
}

func configurePHYPad(mux uint32, pad uint32, daisy uint32, mode uint32, ctl uint32) (p *iomuxc.Pad) {
	p = &iomuxc.Pad{
		Mux:   mux,
		Pad:   pad,
		Daisy: daisy,
	}

	p.Mode(mode)
	p.Ctl(ctl)

	return
}

func configurePHYPads() {
	// 50 Mhz pad
	ctl50 := uint32((iomuxc.SW_PAD_CTL_DSE_2_R0_6 << iomuxc.SW_PAD_CTL_DSE) |
		(iomuxc.SW_PAD_CTL_SPEED_50MHZ << iomuxc.SW_PAD_CTL_SPEED) |
		(1 << iomuxc.SW_PAD_CTL_PUE) | (1 << iomuxc.SW_PAD_CTL_PKE) |
		(iomuxc.SW_PAD_CTL_PUS_PULL_UP_100K << iomuxc.SW_PAD_CTL_PUS) |
		(1 << iomuxc.SW_PAD_CTL_HYS))

	// 100 Mhz pad
	ctl100 := uint32((iomuxc.SW_PAD_CTL_DSE_2_R0_6 << iomuxc.SW_PAD_CTL_DSE) |
		(iomuxc.SW_PAD_CTL_SPEED_100MHZ << iomuxc.SW_PAD_CTL_SPEED) |
		(1 << iomuxc.SW_PAD_CTL_PUE) | (1 << iomuxc.SW_PAD_CTL_PKE) |
		(iomuxc.SW_PAD_CTL_PUS_PULL_UP_100K << iomuxc.SW_PAD_CTL_PUS) |
		(1 << iomuxc.SW_PAD_CTL_HYS))

	// [ALT0] ENET2_RDATA01
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_DATA0,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_DATA0,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_RDATA01
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_DATA1,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_DATA1,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_RX_EN
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_EN,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_EN,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_TDATA00
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_DATA0,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_DATA0,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_TDATA01
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_DATA1,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_DATA1,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_TX_EN
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_EN,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_EN,
		0, IOMUX_ALT0, ctl100)

	// [ALT4] ENET2_REF_CLK / SION ENABLED
	pad := configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_TX_CLK,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_TX_CLK,
		IOMUXC_ENET1_REF_CLK1_SELECT_INPUT,
		IOMUX_ALT4, ctl50)
	pad.Select(DAISY_ENET1_TX_CLK_ALT4)
	pad.SoftwareInput(true)

	// [ALT0] ENET2_RX_ER
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_ENET1_RX_ER,
		IOMUXC_SW_PAD_CTL_PAD_ENET1_RX_ER,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET1_MDIO
	pad = configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO06,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO06,
		IOMUXC_ENET1_MAC0_MDIO_SELECT_INPUT,
		IOMUX_ALT0, ctl100)
	pad.Select(DAISY_ENET1_GPIO1_IO06_ALT0)

	// [ALT0] ENET1_MDC
	configurePHYPad(
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07,
		0, IOMUX_ALT0, ctl100)
}

func EnablePHY(eth *enet.ENET) error {
	var pa int

	switch eth.Index {
	case 1:
		pa = 2
	case 2:
		pa = 1
	default:
		return errors.New("invalid index")
	}

	configurePHYPads()

	// Software reset
	eth.WriteMII(pa, KSZ_CTRL, (1 << CTRL_RESET))
	// HP Auto MDI/MDI-X mode, RMII 50MHz, LEDs: Activity/Link
	eth.WriteMII(pa, KSZ_PHYCTRL2, (1<<CTRL2_HP_MDIX)|(1<<CTRL2_RMII)|(1<<CTRL2_LED))
	// 100 Mbps, Full-duplex
	eth.WriteMII(pa, KSZ_CTRL, (1<<CTRL_SPEED)|(1<<CTRL_DUPLEX))
	// enable interrupts
	eth.WriteMII(pa, KSZ_INT, 0xff00)

	return nil
}
