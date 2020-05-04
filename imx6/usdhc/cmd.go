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
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	GO_IDLE_STATE = 0

	// p127, 4.9.5 (Published RCA response), SD-PL-7.10
	RCA_ADDR   = 16
	RCA_STATUS = 0

	// p131, Table 4-42 : Card Status, SD-PL-7.10
	// p160, Table 68 - Device Status, JESD84-B51
	STATUS_CURRENT_STATE = 9
	STATUS_APP_CMD       = 5
	CURRENT_STATE_IDENT  = 2
	CURRENT_STATE_TRAN   = 4

	WRITE = 0
	READ  = 1

	RSP_NONE          = 0b00
	RSP_136           = 0b01
	RSP_48            = 0b10
	RSP_48_CHECK_BUSY = 0b11

	CMD_TIMEOUT = 10 * time.Millisecond
)

func (hw *usdhc) rsp(i uint32) uint32 {
	if i > 3 {
		return 0
	}

	return reg.Read(hw.cmd_rsp + i*4)
}

// cmd sends an SD / MMC command as described in
// p142, 6.10.4 Detailed command description, JEDEC Standard No. 84-B51
// and
// p349, 35.4.3 Send command to card flow chart, IMX6FG
func (hw *usdhc) cmd(index uint32, dtd uint32, arg uint32, res uint32, cic bool, ccc bool) (err error) {
	// clear interrupts status
	reg.Write(hw.int_status, 0xffffffff)

	// enable interrupt status
	reg.Write(hw.int_status_en, 0xffffffff)

	// wait for command inhibit to be clear
	if !reg.WaitFor(CMD_TIMEOUT, hw.pres_state, PRES_STATE_CIHB, 1, 0) {
		return errors.New("command inhibit")
	}

	// TODO
	data := false

	// wait for data inhibit to be clear
	if data && !reg.WaitFor(CMD_TIMEOUT, hw.pres_state, PRES_STATE_CDIHB, 1, 0) {
		return errors.New("data inhibit")
	}

	defer func() {
		if err != nil {
			reg.Clear(hw.pres_state, PRES_STATE_CIHB)
			reg.Clear(hw.pres_state, PRES_STATE_CDIHB)
			reg.Set(hw.sys_ctrl, SYS_CTRL_RSTC)
		}
	}()

	// disable DMA
	reg.ClearN(hw.prot_ctrl, PROT_CTRL_DMASEL, 0b11)

	// command configuration

	// set command arguments
	reg.Write(hw.cmd_arg, arg)

	xfr := reg.Read(hw.cmd_xfr)
	mix := reg.Read(hw.mix_ctrl)

	// set command index
	bits.SetN(&xfr, CMD_XFR_TYP_CMDINX, 0b111111, index)

	// command index verification
	if cic {
		bits.Set(&xfr, CMD_XFR_TYP_CICEN)
	} else {
		bits.Clear(&xfr, CMD_XFR_TYP_CICEN)
	}

	// CRC verification
	if ccc {
		bits.Set(&xfr, CMD_XFR_TYP_CCCEN)
	} else {
		bits.Clear(&xfr, CMD_XFR_TYP_CCCEN)
	}

	// set response type
	bits.SetN(&xfr, CMD_XFR_TYP_RSPTYP, 0b11, res)
	// set data transfer direction
	bits.SetN(&mix, MIX_CTRL_DTDSEL, 1, dtd)

	if hw.ddr {
		bits.Set(&mix, MIX_CTRL_DDR_EN)
	}

	reg.Write(hw.mix_ctrl, mix)
	reg.Write(hw.cmd_xfr, xfr)

	// wait for completion
	if !reg.WaitFor(CMD_TIMEOUT, hw.int_status, INT_STATUS_CC, 1, 1) {
		err = errors.New("command timeout")
	} else {
		// mask all interrupts
		reg.Write(hw.int_signal_en, 0)
	}

	// read status
	status := reg.Read(hw.int_status)

	// check for any error value
	if (status >> 16) > 0 {
		err = fmt.Errorf("CMD%d error, interrupt status %x", index, status)
	}

	return
}
