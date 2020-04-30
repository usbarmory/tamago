// NXP Ultra Secured Digital Host Controller (uSDHC) driver
// https://github.com/f-secure-foundry/tamago
//
// IP: https://www.mobiveil.com/esdhc/
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package usdhc

import (
	"errors"
	"fmt"
	"time"

	"github.com/f-secure-foundry/tamago/internal/bits"
)

const (
	// p181, 7.1 OCR register, JESD84-B51
	MMC_OCR_BUSY        = 31
	MMC_OCR_ACCESS_MODE = 29
	MMC_OCR_VDD_HV_MAX  = 23
	MMC_OCR_VDD_HV_MIN  = 15
	MMC_OCR_VDD_LV      = 7

	// p62, 6.6.1 Command sets and extended settings, JESD84-B51
	MMC_SWITCH_ACCESS  = 24
	MMC_SWITCH_INDEX   = 16
	MMC_SWITCH_VALUE   = 8
	MMC_SWITCH_CMD_SET = 0

	// p193, 7.4 Extended CSD register, JESD84-B51
	EXT_CSD_BUS_WIDTH = 183

	ACCESS_WRITE_BYTE = 0b11

	ACCESS_MODE_BYTE   = 0b00
	ACCESS_MODE_SECTOR = 0b10

	MMC_DETECT_LOOP_CNT = 300
	MMC_DETECT_TIMEOUT  = 1 * time.Second
)

// p352, 35.4.6 MMC voltage validation flow chart, IMX6FG
func (hw *usdhc) voltageValidationMMC() (mmc bool, hc bool) {
	var arg uint32

	// CMD1 - SEND_OP_COND
	// p57, 6.4.2 Access mode validation, JESD84-B51

	// sector mode supported
	bits.SetN(&arg, MMC_OCR_ACCESS_MODE, 0b11, ACCESS_MODE_SECTOR)
	// set HV range
	bits.SetN(&arg, MMC_OCR_VDD_HV_MIN, 0x1ff, 0x1ff)

	// p46, 6.3.1 Device reset to Pre-idle state, JESD84-B51
	time.Sleep(1 * time.Millisecond)

	start := time.Now()

	for i := 0; i < 300; i++ {
		if err := hw.cmd(1, READ, arg, RSP_48, false, false); err != nil {
			return false, false
		}

		rsp := hw.rsp(0)

		if bits.Get(&rsp, MMC_OCR_BUSY, 1) == 0 && time.Since(start) < MMC_DETECT_TIMEOUT {
			continue
		}

		if bits.Get(&rsp, MMC_OCR_ACCESS_MODE, 0b11) == ACCESS_MODE_SECTOR {
			hc = true
		} else {
			hc = false
		}

		break
	}

	return true, hc
}

// p352, 35.4.7 MMC card initialization flow chart, IMX6FG
// p58, 6.4.4 Device identification process, JESD84-B51
func (hw *usdhc) initMMC() (err error) {
	var arg uint32
	var bus_width uint32

	// CMD2 - ALL_SEND_CID - get unique card identification
	err = hw.cmd(2, READ, arg, RSP_136, false, true)

	if err != nil {
		return
	}

	// This implementation re-uses the USDHC controller index as unique RCA
	// value.
	rca := uint32(hw.n << RCA_ADDR)

	// CMD3 - SET_RELATIVE_ADDR - set relative card address (RCA),
	err = hw.cmd(3, READ, rca, RSP_48, true, true)

	if err != nil {
		return
	}

	if state := (hw.rsp(0) >> STATUS_CURRENT_STATE) & 0b1111; state != CURRENT_STATE_IDENT {
		return fmt.Errorf("card not in ident state (%d)", state)
	}

	// clear clock
	hw.setClock(0, 0)
	// set operating frequency
	hw.setClock(DVS_OP, SDCLKFS_OP)

	// CMD7 - SELECT/DESELECT CARD - enter transfer state
	err = hw.cmd(7, READ, rca, RSP_48_CHECK_BUSY, true, true)

	if err != nil {
		return
	}

	// CMD13 - SEND_STATUS - poll card status
	err = hw.cmd(13, READ, rca, RSP_48, true, true)

	if state := (hw.rsp(0) >> STATUS_CURRENT_STATE) & 0b1111; state != CURRENT_STATE_TRAN {
		return fmt.Errorf("card not in tran state (%d)", state)
	}

	// p223, 7.4.67 BUS_WIDTH [183], JESD84-B51
	switch hw.width {
	case 4:
		bus_width = 0b01
	case 8:
		bus_width = 0b10
	default:
		return errors.New("unsupported MMC bus width")
	}

	arg = 0
	// write MMC_SWITCH_VALUE in register pointed in MMC_SWITCH_INDEX
	bits.SetN(&arg, MMC_SWITCH_ACCESS, 0b11, ACCESS_WRITE_BYTE)
	// set MMC_SWITCH_INDEX to BUS_WIDTH
	bits.SetN(&arg, MMC_SWITCH_INDEX, 0xff, EXT_CSD_BUS_WIDTH)
	// set bus width value
	bits.SetN(&arg, MMC_SWITCH_VALUE, 0xff, bus_width)

	// CMD6 - SWITCH - switch mode of operation
	return hw.cmd(6, READ, arg, RSP_48, true, true)
}

// MMC returns whether an MMC card has been detected.
func (hw *usdhc) MMC() bool {
	return hw.mmc
}
