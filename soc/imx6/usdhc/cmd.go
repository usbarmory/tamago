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

package usdhc

import (
	"fmt"
	"time"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// CMD constants
const (
	GO_IDLE_STATE = 0

	// p127, 4.9.5 (Published RCA response), SD-PL-7.10
	RCA_ADDR   = 16
	RCA_STATUS = 0

	// p131, Table 4-42 : Card Status, SD-PL-7.10
	// p160, Table 68 - Device Status, JESD84-B51
	STATUS_CURRENT_STATE = 9
	STATUS_SWITCH_ERROR  = 7
	STATUS_APP_CMD       = 5
	CURRENT_STATE_IDENT  = 2
	CURRENT_STATE_TRAN   = 4

	WRITE = 0
	READ  = 1

	RSP_NONE          = 0b00
	RSP_136           = 0b01
	RSP_48            = 0b10
	RSP_48_CHECK_BUSY = 0b11

	// SEND_CSD response contains CSD[127:8],
	CSD_RSP_OFF = -8

	DEFAULT_CMD_TIMEOUT = 10 * time.Millisecond
)

// cmd sends an SD / MMC command as described in
// p349, 35.4.3 Send command to card flow chart, IMX6FG
func (hw *USDHC) cmd(index uint32, dtd uint32, arg uint32, res uint32, cic bool, ccc bool, dma bool, timeout time.Duration) (err error) {
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

	if dtd == WRITE && reg.Get(hw.pres_state, PRES_STATE_WPSPL, 1) == 0 {
		// The uSDHC merely reports on WP, it doesn't really act on it
		// despite IMX6ULLRM suggesting otherwise (e.g. p4017).
		return fmt.Errorf("card is write protected")
	}

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
	// clear special command types
	bits.SetN(&xfr, CMD_XFR_TYP_CMDTYP, 0b11, 0)

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

	if hw.card.DDR {
		// enable dual data rate
		bits.Set(&mix, MIX_CTRL_DDR_EN)
	} else {
		bits.Clear(&mix, MIX_CTRL_DDR_EN)
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
	} else {
		bits.Clear(&xfr, CMD_XFR_TYP_DPSEL)
		bits.Clear(&mix, MIX_CTRL_MSBSEL)
		bits.Clear(&mix, MIX_CTRL_AC12EN)
		bits.Clear(&mix, MIX_CTRL_BCEN)
		bits.Clear(&mix, MIX_CTRL_DMAEN)
	}

	if hw.rpmb {
		bits.Clear(&mix, MIX_CTRL_MSBSEL)
	}

	// set response type
	bits.SetN(&xfr, CMD_XFR_TYP_RSPTYP, 0b11, res)
	// set data transfer direction
	bits.SetN(&mix, MIX_CTRL_DTDSEL, 1, dtd)

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
		err = fmt.Errorf("CMD%d:timeout pres_state:%#x int_status:%#x", index,
			reg.Read(hw.pres_state),
			reg.Read(hw.int_status))
		// According to the IMX6FG flow chart we shouldn't return in
		// case of error, but still go ahead and check status.
	}

	// mask all interrupts
	reg.Write(hw.int_signal_en, 0)

	// read status
	status := reg.Read(hw.int_status)

	// p3997, 58.5.3.5.4 Auto CMD12 Error, IMX6ULLRM
	if (status >> 16) == ((1 << INT_STATUS_AC12E) >> 16) {
		// retry once CMD12 if the Auto one fails
		if err := hw.cmd(12, READ, 0, RSP_NONE, true, true, false, hw.writeTimeout); err == nil {
			bits.Clear(&status, INT_STATUS_AC12E)
		}
	}

	if (status >> 16) > 0 {
		msg := fmt.Sprintf("pres_state:%#x int_status:%#x", reg.Read(hw.pres_state), status)

		if bits.Get(&status, INT_STATUS_AC12E, 1) == 1 {
			msg += fmt.Sprintf(" AC12:%#x", reg.Read(hw.ac12_err_status))
		}

		err = fmt.Errorf("CMD%d:error %s", index, msg)
	}

	return
}

func (hw *USDHC) rsp(i int) uint32 {
	if i > 3 {
		return 0
	}

	return reg.Read(hw.cmd_rsp + uint32(i*4))
}

func (hw *USDHC) rspVal(pos int, mask int) (val uint32) {
	val = hw.rsp(pos/32) >> (pos % 32)
	val &= uint32(mask)
	return
}

func (hw *USDHC) waitState(state int, timeout time.Duration) (err error) {
	start := time.Now()

	for {
		// CMD13 - SEND_STATUS - poll card status
		if err = hw.cmd(13, READ, hw.rca, RSP_48, true, true, false, hw.writeTimeout); err != nil {
			if time.Since(start) >= timeout {
				return fmt.Errorf("error polling card status, %v", err)
			}

			continue
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
