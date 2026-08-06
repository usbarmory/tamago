// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mk2

import (
	"github.com/usbarmory/tamago/phy/ti/dp83825i"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/iomuxc"
)

// Ethernet PHY configuration constants.
//
// On the UA-MKII-LAN the ENET2 MAC is connected to a DP83825I PHY, this board
// package configures it at 100 Mbps / Full-duplex mode.
const (
	// ENET2 MUX
	IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_DATA0 = 0x020e00e4
	IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_DATA1 = 0x020e00e8
	IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_EN    = 0x020e00ec
	IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_DATA0 = 0x020e00f0
	IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_DATA1 = 0x020e00f4
	IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_EN    = 0x020e00f8
	IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_CLK   = 0x020e00fc
	IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_ER    = 0x020e0100

	// ENET2 PAD
	IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_DATA0 = 0x020e0370
	IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_DATA1 = 0x020e0374
	IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_EN    = 0x020e0378
	IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_DATA0 = 0x020e037c
	IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_DATA1 = 0x020e0380
	IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_EN    = 0x020e0384
	IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_CLK   = 0x020e0388
	IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_ER    = 0x020e038c

	// ENET2 SELECT INPUT
	IOMUXC_ENET2_REF_CLK2_SELECT_INPUT  = 0x020e057c
	IOMUXC_ENET2_MAC0_MDIO_SELECT_INPUT = 0x020e0580

	// MDIO (already defined as BT_SWDIO in ble.go)
	//IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO06
	//IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO06

	// MDC (already defined as BT_UART_RTS in ble.go)
	//IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07
	//IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07

	IOMUX_ALT0 = 0
	IOMUX_ALT1 = 1
	IOMUX_ALT4 = 4

	DAISY_ENET2_TX_CLK_ALT4     = 0b10
	DAISY_ENET2_GPIO1_IO06_ALT0 = 0

	PHY_ADDR = 0
)

func init() {
	if imx6ul.ENET1 != nil {
		// ENET1 is only used on emulated runs
		imx6ul.ENET1.EnablePHY = EnablePHY
		imx6ul.ENET1.RMII = true
	}

	if imx6ul.ENET2 != nil {
		// ENET2 is only used on UA-MKII-NET
		imx6ul.ENET2.EnablePHY = EnablePHY
		imx6ul.ENET2.RMII = true
	}
}

func ctl50() uint32 {
	return (iomuxc.SW_PAD_CTL_DSE_2_R0_6 << iomuxc.SW_PAD_CTL_DSE) |
		(iomuxc.SW_PAD_CTL_SPEED_50MHZ << iomuxc.SW_PAD_CTL_SPEED) |
		(1 << iomuxc.SW_PAD_CTL_PUE) | (1 << iomuxc.SW_PAD_CTL_PKE) |
		(iomuxc.SW_PAD_CTL_PUS_PULL_UP_100K << iomuxc.SW_PAD_CTL_PUS) |
		(1 << iomuxc.SW_PAD_CTL_HYS)
}

func ctl100() uint32 {
	return (iomuxc.SW_PAD_CTL_DSE_2_R0_6 << iomuxc.SW_PAD_CTL_DSE) |
		(iomuxc.SW_PAD_CTL_SPEED_100MHZ << iomuxc.SW_PAD_CTL_SPEED) |
		(1 << iomuxc.SW_PAD_CTL_PUE) | (1 << iomuxc.SW_PAD_CTL_PKE) |
		(iomuxc.SW_PAD_CTL_PUS_PULL_UP_100K << iomuxc.SW_PAD_CTL_PUS) |
		(1 << iomuxc.SW_PAD_CTL_HYS)
}

func configurePHYPads() {
	// 50 Mhz pad
	ctl50 := ctl50()
	// 100 Mhz pad
	ctl100 := ctl100()

	// [ALT0] ENET2_RDATA01
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_DATA0,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_DATA0,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_RDATA01
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_DATA1,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_DATA1,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_RX_EN
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_EN,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_EN,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_TDATA00
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_DATA0,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_DATA0,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_TDATA01
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_DATA1,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_DATA1,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_TX_EN
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_EN,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_EN,
		0, IOMUX_ALT0, ctl100)

	// [ALT4] ENET2_REF_CLK / SION ENABLED
	pad := configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_TX_CLK,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_TX_CLK,
		IOMUXC_ENET2_REF_CLK2_SELECT_INPUT,
		IOMUX_ALT4, ctl50)
	pad.Select(DAISY_ENET2_TX_CLK_ALT4)
	pad.SoftwareInput(true)

	// [ALT0] ENET2_RX_ER
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_ENET2_RX_ER,
		IOMUXC_SW_PAD_CTL_PAD_ENET2_RX_ER,
		0, IOMUX_ALT0, ctl100)

	// [ALT0] ENET2_MDIO
	pad = configurePad(
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO06,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO06,
		IOMUXC_ENET2_MAC0_MDIO_SELECT_INPUT,
		IOMUX_ALT1, ctl100)
	pad.Select(DAISY_ENET2_GPIO1_IO06_ALT0)

	// [ALT0] ENET2_MDC
	configurePad(
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07,
		0, IOMUX_ALT1, ctl100)
}

var phy *dp83825i.PHY

func EnablePHY(eth *enet.ENET) error {
	configurePHYPads()

	phy = &dp83825i.PHY{}

	return phy.Init(PHY_ADDR, eth)
}
