// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan9696evb

import (
	"github.com/usbarmory/tamago/soc/microchip/lan969x"
)

// https://microchip-ung.github.io/lan969x-industrial_reginfo/reginfo_LAN969x-Industrial.html

// Analyzer registers
const (
	PORT_BASE = lan969x.ANA_CL_BASE + 0x10000
	PORT29    = PORT_BASE + (29 * 512) // D29
	PORT30    = PORT_BASE + (30 * 512) // D30 (CPU port 0)

	FILTER_CTRL          = 0x04
	FORCE_FCS_UPDATE_ENA = 0

	VLAN_CTRL      = 0x20
	VLAN_POP_CNT   = 17
	VLAN_AWARE_ENA = 19
	PORT_VID       = 0

	FWD_CFG           = lan969x.ANA_L2_BASE + 0x89308
	CPU_DMAC_COPY_ENA = 6
)

// Queue Forwarding registers
const (
	SWITCH_PORT_MODE_BASE = lan969x.QFWD_BASE + 0
	SWITCH_PORT_MODE29    = SWITCH_PORT_MODE_BASE + (29 * 4) // D29
	SWITCH_PORT_MODE30    = SWITCH_PORT_MODE_BASE + (30 * 4) // D30 (CPU port 0)
	PORT_ENA              = 19
)

// Disassembler registers
const (
	DEV_TX_STOP_WM_CFG = lan969x.DSM_CFG + 0x550
	DEV_TX_CNT_CLR     = 0
)

// Assembler registers
const (
	PORT_CFG_BASE   = lan969x.ASM_BASE + 0x4780 + 0x21c
	PORT_CFG30      = PORT_CFG_BASE + (30 * 4) // D30 (CPU port 0)
	NO_PREAMBLE_ENA = 9
	PAD_ENA         = 6
	INJ_FORMAT_CFG  = 2
)

// XMII registers
const (
	XMIICFG0 = lan969x.HSIO_BASE + 0x74
	XMIICFG1 = lan969x.HSIO_BASE + 0x88

	XMII_CFG = 0x00

	GPIO_XMII_CFG = 1
	CFG_GPIO      = 0
	CFG_RGMII     = 1
	CFG_RMII      = 2

	RGMII_CFG    = 0x04
	TX_CLK_CFG   = 2
	RGMII_TX_RST = 1
	RGMII_RX_RST = 0

	DLL_CFG0 = 0x0c // rx
	DLL_CFG1 = 0x10 // tx

	DLL_ENA     = 19
	DLL_CLK_ENA = 18
	DLL_CLK_SEL = 15
	DLL_RST     = 0
)

// DEVRGMII registers
const (
	DEVRGMII1 = lan969x.DEVRGMII1

	DEV_CFG_STATUS = 0x00

	DEV_RST_CTRL = DEV_CFG_STATUS + 0x00

	SPEED_SEL  = 20
	SPEED_10M  = 0
	SPEED_100M = 1
	SPEED_1G   = 2

	MAC_TX_RST = 4
	MAC_RX_RST = 0

	MAC_CFG_STATUS = 0x24

	MAC_ENA_CFG = MAC_CFG_STATUS + 0x00
	RX_ENA      = 4
	TX_ENA      = 0

	MAC_MODE_CFG = MAC_CFG_STATUS + 0x04
	FDX_ENA      = 0

	MAC_IFG_CFG = MAC_CFG_STATUS + 0x18
	TX_IFG      = 8
	RX_IFG2     = 4
	RX_IFG1     = 0
)
