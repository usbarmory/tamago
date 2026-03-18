// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan9696evb

import (
	"github.com/usbarmory/tamago/soc/microchip/analyzer"
	"github.com/usbarmory/tamago/soc/microchip/lan969x"
)

// https://microchip-ung.github.io/lan969x-industrial_reginfo/reginfo_LAN969x-Industrial.html

// Miscellaneous initialization registers
const (
	ASM_RAM_INIT        = lan969x.ASM_BASE + 0x3d44
	ANA_AC_RAM_INIT     = lan969x.ANA_AC_BASE + 0x31510
	REW_RAM_INIT        = lan969x.REW_BASE + 0x16f28
	DSM_RAM_INIT        = lan969x.DSM_BASE + 0x00
	EACL_RAM_INIT       = lan969x.EACL_BASE + 0x71b8
	VCAP_SUPER_RAM_INIT = lan969x.VCAP_SUPER_BASE + 0x460
	VOP_RAM_INIT        = lan969x.VOP_BASE + 0x3ff0
	RAM_INIT            = 1

	ASM_STAT_CFG = lan969x.ASM_BASE + 0x4780
	RESET        = 0

	ANA_AC_STAT_GLOBAL_CFG_PORT = lan969x.ANA_AC_BASE + 0x31540
	ANA_AC_STAT_RESET           = ANA_AC_STAT_GLOBAL_CFG_PORT + 0x10
	STAT_CNT_CLR_SHOT           = 0

	PTP_CFG     = lan969x.PTP_BASE + 0x200
	PTP_DOM_CFG = PTP_CFG + 0x0c
	PTP_ENA     = 9

	PTP_TOD_DOMAINS = lan969x.PTP_BASE + 0x210
	CLK_PER_CFG0    = PTP_TOD_DOMAINS + 0x00
	CLK_PER_CFG1    = PTP_TOD_DOMAINS + 0x04

	UBOOT_CLK_PER = 0x18624dd2
)

// Calendar registers
const (
	QSYS_RAM_INIT = lan969x.QSYS_BASE + 0x89c
	QSYS_CALCFG   = lan969x.QSYS_BASE + 0x874

	QSYS_CAL_AUTO0 = QSYS_CALCFG + 0x00
	QSYS_CAL_AUTO1 = QSYS_CALCFG + 0x04
	QSYS_CAL_AUTO2 = QSYS_CALCFG + 0x08
	QSYS_CAL_AUTO3 = QSYS_CALCFG + 0x0c
	GBPS_1         = 1

	QSYS_CAL_CTRL     = QSYS_CALCFG + 0x24
	CAL_CTRL_CAL_MODE = 11
	MODE_HALT         = 10
	MODE_CAL_AUTO     = 8

	DSM_CFG      = lan969x.DSM_BASE + 0x14
	TAXI_CAL_CFG = DSM_CFG + 0xc98
	CAL_SEL_STAT = 23
	CAL_SWITCH   = 22
	CAL_PGM_SEL  = 21
	CAL_IDX      = 15
	CAL_PGM_VAL  = 1
	CAL_PGM_ENA  = 0
)

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

	VLAN_BASE = lan969x.ANA_L3_BASE

	VLAN_CFG_MAC_VID = VLAN_BASE + (analyzer.MAC_VID * 0x40) + 8
	VLAN_FID         = 8
	VLAN_LRN_DIS     = 3

	COMMON_BASE = lan969x.ANA_L3_BASE + 0x5a840

	COMMON_VLAN_CTRL = COMMON_BASE + 0x4
	VLAN_ENA         = 0
)

// Queue Forwarding registers
const (
	SWITCH_PORT_MODE_BASE = lan969x.QFWD_BASE + 0
	SWITCH_PORT_MODE29    = SWITCH_PORT_MODE_BASE + (29 * 4) // D29
	SWITCH_PORT_MODE30    = SWITCH_PORT_MODE_BASE + (30 * 4) // D30 (CPU port 0)
	PORT_ENA              = 19
)

// Assembler registers
const (
	PORT_CFG_BASE   = lan969x.ASM_BASE + 0x4780 + 0x21c
	PORT_CFG30      = PORT_CFG_BASE + (30 * 4) // D30 (CPU port 0)
	NO_PREAMBLE_ENA = 9
	PAD_ENA         = 6
	INJ_FORMAT_CFG  = 2
)

// PHY registers
const (
	PHY_ADDR = 0x03

	PHY_CTRL    = 0x00
	CTRL_RESET  = 15
	CTRL_SPEED0 = 13
	CTRL_ANEG   = 12
	CTRL_DUPLEX = 8
	CTRL_SPEED1 = 6
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
