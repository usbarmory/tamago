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

	DEFAULT_CMD_TIMEOUT = 10 * time.Millisecond
)

// cmd sends an SD / MMC command as described in
// p142, 6.10.4 Detailed command description, JEDEC Standard No. 84-B51
// and
// p349, 35.4.3 Send command to card flow chart, IMX6FG
func (hw *usdhc) cmd(index uint32, dtd uint32, arg uint32, res uint32, cic bool, ccc bool, dma bool, timeout time.Duration) (err error) {
	if timeout == 0 {
		timeout = DEFAULT_CMD_TIMEOUT
	}

	// clear interrupt status
	reg.Write(hw.int_status, 0xffffffff)

	// enable interrupt status
	reg.Write(hw.int_status_en, 0xffffffff)

	// wait for command inhibit to be clear
	if !reg.WaitFor(timeout, hw.pres_state, PRES_STATE_CIHB, 1, 0) {
		return fmt.Errorf("CMD%d command inhibit", index)
	}

	// wait for data inhibit to be clear
	if dma && !reg.WaitFor(timeout, hw.pres_state, PRES_STATE_CDIHB, 1, 0) {
		return fmt.Errorf("CMD%d data inhibit", index)
	}

	// clear interrupts status
	reg.Write(hw.int_status, 0xffffffff)

	defer func() {
		if err != nil {
			reg.Clear(hw.pres_state, PRES_STATE_CIHB)
			reg.Clear(hw.pres_state, PRES_STATE_CDIHB)
			reg.Set(hw.sys_ctrl, SYS_CTRL_RSTC)
		}
	}()

	dmasel := uint32(DMASEL_NONE)

	if dma {
		dmasel = DMASEL_ADMA2
		reg.Write(hw.int_signal_en, 0xffffffff)
	}

	// select DMA mode
	reg.SetN(hw.prot_ctrl, PROT_CTRL_DMASEL, 0b11, dmasel)

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

	if hw.card.DDR {
		// enable dual data rate
		bits.Set(&mix, MIX_CTRL_DDR_EN)
	}

	if dma {
		// enable data presence
		bits.Set(&xfr, CMD_XFR_TYP_DPSEL)
		// enable multiple blocks
		bits.Set(&mix, MIX_CTRL_MSBSEL)
		// enable automatic CMD12 to stop transactions
		bits.Set(&mix, MIX_CTRL_AC12EN)
		// enable block count
		bits.Set(&mix, MIX_CTRL_BCEN)
		// enable DMA
		bits.Set(&mix, MIX_CTRL_DMAEN)
	}

	reg.Write(hw.mix_ctrl, mix)
	reg.Write(hw.cmd_xfr, xfr)

	// command completion
	int_status := INT_STATUS_CC

	if dma {
		// transfer completion
		int_status = INT_STATUS_TC
	}

	// wait for completion
	if !reg.WaitFor(timeout, hw.int_status, int_status, 1, 1) {
		err = fmt.Errorf("CMD%d timeout", index)
	}

	// mask all interrupts
	reg.Write(hw.int_signal_en, 0)

	// read status
	status := reg.Read(hw.int_status)

	// check for any error value
	if (status >> 16) > 0 {
		err = fmt.Errorf("CMD%d error status %x", index, status)
	}

	return
}

func (hw *usdhc) rsp(i uint32) uint32 {
	if i > 3 {
		return 0
	}

	return reg.Read(hw.cmd_rsp + i*4)
}

func (hw *usdhc) waitState(state int, timeout time.Duration) (err error) {
	start := time.Now()

	for {
		// CMD13 - SEND_STATUS - poll card status
		if err = hw.cmd(13, READ, hw.rca, RSP_48, true, true, false, 0); err != nil {
			return
		}

		curState := (hw.rsp(0) >> STATUS_CURRENT_STATE) & 0b1111

		if curState == uint32(state) {
			break
		}

		if time.Since(start) >= timeout {
			return fmt.Errorf("expected card state %d, got %d", state, curState)
		}
	}

	return
}
