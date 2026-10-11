// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/analyzer"
)

const (
	// Memory manager status
	HSCH_MMGT = HSCH_BASE + 0x8da4
	RESET_CFG = HSCH_MMGT + 0x8
)

// Miscellaneous initialization registers
const (
	ASM_RAM_INIT        = ASM_BASE + 0x3d44
	ANA_AC_RAM_INIT     = ANA_AC_BASE + 0x31510
	REW_RAM_INIT        = REW_BASE + 0x16f28
	DSM_RAM_INIT        = DSM_BASE + 0x00
	EACL_RAM_INIT       = EACL_BASE + 0x71b8
	VCAP_SUPER_RAM_INIT = VCAP_SUPER_BASE + 0x460
	VOP_RAM_INIT        = VOP_BASE + 0x3ff0
	RAM_INIT            = 1

	ASM_STAT_CFG = ASM_BASE + 0x4780
	RESET        = 0

	ANA_AC_STAT_GLOBAL_CFG_PORT = ANA_AC_BASE + 0x31540
	ANA_AC_STAT_RESET           = ANA_AC_STAT_GLOBAL_CFG_PORT + 0x10
	STAT_CNT_CLR_SHOT           = 0

	PTP_CFG     = PTP_BASE + 0x200
	PTP_DOM_CFG = PTP_CFG + 0x0c
	PTP_ENA     = 9

	PTP_TOD_DOMAINS = PTP_BASE + 0x210
	CLK_PER_CFG0    = PTP_TOD_DOMAINS + 0x00
	CLK_PER_CFG1    = PTP_TOD_DOMAINS + 0x04

	UBOOT_CLK_PER = 0x18624dd2
)

// VLAN registers
const (
	MAC_FID = 1

	VLAN_BASE = ANA_L3_BASE

	VLAN_CFG_MAC_VID = VLAN_BASE + (analyzer.MAC_VID * 0x40) + 8
	VLAN_FID         = 8
	VLAN_LRN_DIS     = 3

	COMMON_BASE = ANA_L3_BASE + 0x5a840

	COMMON_VLAN_CTRL = COMMON_BASE + 0x4
	VLAN_ENA         = 0
)

// EnableSwitchCore resets and initializes the switching core.
func EnableSwitchCore() {
	// enable switch core
	reg.Write(RESET_CFG, 1)
}

func ramInit(addr uint32) {
	reg.Set(addr, RAM_INIT)
	reg.Wait(addr, RAM_INIT, 1, 0)
}

// InitSwitchCore resets, initializes and enables the switch core.
func InitSwitchCore() {
	// reset switch core, which a VCore reset leaves running
	SoftReset(VCoreProtection)
	reg.Wait(GCB_SOFT_RST, SOFT_SWC_RST, 1, 0)

	// initialize memories
	ramInit(ANA_AC_RAM_INIT)     // Analyzer Access Control
	ramInit(ASM_RAM_INIT)        // Assembler
	ramInit(DSM_RAM_INIT)        // Disassembler
	ramInit(EACL_RAM_INIT)       // Egress Access Control Lists
	ramInit(QSYS_RAM_INIT)       // Queue System Configuration
	ramInit(REW_RAM_INIT)        // Rewriter
	ramInit(VCAP_SUPER_RAM_INIT) // Versatile Content Aware Processor
	ramInit(VOP_RAM_INIT)        // Versatile OAM MEP Processor

	// reset counters
	reg.Set(ANA_AC_STAT_RESET, RESET)
	reg.Set(ASM_STAT_CFG, STAT_CNT_CLR_SHOT)

	// set internal bandwith for all ports to 1 Gbps
	// (10 ports per 32 bit word)
	for port := range 10 {
		reg.SetN(QSYS_CAL_AUTO0, 3*port, 0b111, GBPS_1)
		reg.SetN(QSYS_CAL_AUTO1, 3*port, 0b111, GBPS_1)
		reg.SetN(QSYS_CAL_AUTO2, 3*port, 0b111, GBPS_1)
		reg.SetN(QSYS_CAL_AUTO3, 3*port, 0b111, GBPS_1)
	}

	// time of day clock configuration
	reg.Write(CLK_PER_CFG0, UBOOT_CLK_PER)
	reg.Write(CLK_PER_CFG1, UBOOT_CLK_PER)

	// enable master counter
	reg.Set(PTP_DOM_CFG, PTP_ENA)

	// halt the calendar
	reg.SetN(QSYS_CAL_CTRL, CAL_CTRL_CAL_MODE, 0xf, MODE_HALT)

	// configure calendars
	for taxi := range uint32(5) {
		calendarConfig(taxi)
	}

	// enable automatic sequence mode
	reg.SetN(QSYS_CAL_CTRL, CAL_CTRL_CAL_MODE, 0xf, MODE_CAL_AUTO)

	// start frame analyzer
	ANA.Init()
	ANA.Insert([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, MAC_FID, analyzer.PGID_BROADCAST)

	// map PVID = FID
	var val uint32
	bits.SetN(&val, VLAN_FID, 0x1fff, analyzer.MAC_VID)

	// disable learning
	bits.Set(&val, VLAN_LRN_DIS)
	reg.Write(VLAN_CFG_MAC_VID, val)

	// enable VLANs
	reg.Set(COMMON_VLAN_CTRL, VLAN_ENA)

	// enable switch core
	EnableSwitchCore()
}
