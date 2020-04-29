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
	"time"

	"github.com/f-secure-foundry/tamago/imx6/internal/bits"
)

const (
	// p181, 7.1 OCR register, JESD84-B51
	MMC_OCR_BUSY        = 31
	MMC_OCR_ACCESS_MODE = 29
	MMC_OCR_VDD_HV_MAX  = 23
	MMC_OCR_VDD_HV_MIN  = 15
	MMC_OCR_VDD_LV      = 7

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

// p351, 35.4.7 MMC card initialization flow chart, IMX6FG
func (hw *usdhc) initMMC() (err error) {
	// TODO
	return
}

// MMC returns whether an MMC card has been detected.
func (hw *usdhc) MMC() bool {
	return hw.mmc
}
