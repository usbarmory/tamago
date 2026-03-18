// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan9696evb

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/analyzer"
	"github.com/usbarmory/tamago/soc/microchip/lan969x"
)

func ramInit(addr uint32) {
	reg.Set(addr, RAM_INIT)
	reg.Wait(addr, RAM_INIT, 1, 0)
}

func initializeSwitchCore() {
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
	lan969x.ANA.Init()
	lan969x.ANA.Insert([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, MAC_FID, analyzer.PGID_BROADCAST)

	// map PVID = FID
	var val uint32
	bits.SetN(&val, VLAN_FID, 0x1fff, analyzer.MAC_VID)

	// disable learning
	bits.Set(&val, VLAN_LRN_DIS)
	reg.Write(VLAN_CFG_MAC_VID, val)

	// enable VLANs
	reg.Set(COMMON_VLAN_CTRL, VLAN_ENA)

	// enable switch core
	lan969x.EnableSwitchCore()
}
