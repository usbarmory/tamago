// NXP Ultra Secured Digital Host Controller (uSDHC) driver
// https://github.com/f-secure-foundry/tamago
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
	// p101, 4.3.13 Send Interface Condition Command (CMD8), SD-PL-7.10
	CMD8_ARG_VHS           = 8
	CMD8_ARG_CHECK_PATTERN = 0

	VHS_HIGH      = 0b0001
	VHS_LOW       = 0b0010
	CHECK_PATTERN = 0b10101010

	// p59, 4.2.3.1 Initialization Command (ACMD41), SD-PL-7.10
	// p198, 5.1 OCR register, SD-PL-7.10
	SD_OCR_BUSY       = 31
	SD_OCR_HCS        = 30
	SD_OCR_XPC        = 28
	SD_OCR_S18R       = 24
	SD_OCR_VDD_HV_MAX = 23
	SD_OCR_VDD_HV_MIN = 15
	SD_OCR_VDD_LV     = 7

	SD_DETECT_LOOP_CNT = 300
	SD_DETECT_TIMEOUT  = 1 * time.Second
)

// p350, 35.4.4 SD voltage validation flow chart, IMX6FG
func (hw *usdhc) voltageValidationSD() (sd bool, hc bool) {
	var arg uint32
	var hv bool

	// CMD8 - SEND_EXT_CSD - read device data
	// p101, 4.3.13 Send Interface Condition Command (CMD8), SD-PL-7.10

	bits.SetN(&arg, CMD8_ARG_VHS, 0b1111, VHS_HIGH)
	bits.SetN(&arg, CMD8_ARG_CHECK_PATTERN, 0xff, CHECK_PATTERN)

	if hw.cmd(8, READ, arg, RSP_48, true, true) == nil && hw.rsp(0) == arg {
		// HC/LC HV SD 2.x
		hc = true
		hv = true
	} else {
		arg = VHS_LOW<<CMD8_ARG_VHS | CHECK_PATTERN

		if hw.cmd(8, READ, arg, RSP_48, true, true) == nil && hw.rsp(0) == arg {
			// LC SD 1.x
			hc = true
		} else {
			// LC SD 2.x
			hv = true
		}
	}

	// ACMD41 - SD_SEND_OP_COND - read capacity information
	// p59, 4.2.3.1 Initialization Command (ACMD41), SD-PL-7.10
	//
	// The ACMD41 full argument is the OCR, despite the standard
	// confusingly naming OCR only bits 23-08 of it (which instead
	// represents part of OCR register voltage window).
	arg = 0

	if hc {
		// SDHC or SDXC supported
		bits.Set(&arg, SD_OCR_HCS)
		// Maximum Performance
		bits.Set(&arg, SD_OCR_XPC)
	}

	if hv {
		// set HV range
		bits.SetN(&arg, SD_OCR_VDD_HV_MIN, 0b111111111, 0b111111111)
	} else {
		bits.Set(&arg, SD_OCR_VDD_LV)
	}

	start := time.Now()

	for i := 0; i < SD_DETECT_LOOP_CNT; i++ {
		// CMD55 - APP_CMD - next command is application specific
		if hw.cmd(55, READ, 0, RSP_48, true, true) != nil {
			return false, false
		}

		if err := hw.cmd(41, READ, arg, RSP_48, false, false); err != nil {
			return false, false
		}

		rsp := hw.rsp(0)

		if bits.Get(&rsp, SD_OCR_BUSY, 0b1) == 0 && time.Since(start) < SD_DETECT_TIMEOUT {
			continue
		}

		if bits.Get(&rsp, SD_OCR_HCS, 0b1) == 1 {
			hc = true
		} else {
			hc = false
		}

		break
	}

	return true, hc
}

// SD returns whether an SD card has been detected.
func (hw *usdhc) SD() bool {
	return hw.sd
}
