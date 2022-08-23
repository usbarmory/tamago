// NXP Ultra Secured Digital Host Controller (uSDHC) driver
// https://github.com/usbarmory/tamago
//
// IP: https://www.mobiveil.com/esdhc/
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usdhc

import (
	"fmt"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
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

	// data transfer direction
	WRITE = 0
	READ  = 1

	// response types
	RSP_NONE          = 0b00
	RSP_136           = 0b01
	RSP_48            = 0b10
	RSP_48_CHECK_BUSY = 0b11

	// SEND_CSD response contains CSD[127:8],
	CSD_RSP_OFF = -8

	DEFAULT_CMD_TIMEOUT = 10 * time.Millisecond
)

type cmdParams struct {
	// data transfer direction
	dtd uint32
	// response type
	res uint32
	// command index verification
	cic bool
	// CRC verification
	ccc bool
}

var cmds = map[uint32]cmdParams{
	// CMD0 - GO_IDLE_STATE - reset card
	0: {READ, RSP_NONE, false, false},
	// MMC: CMD1 - SEND_OP_COND - send operating conditions
	1: {READ, RSP_48, false, false},
	// CMD2 - ALL_SEND_CID - get unique card identification
	2: {READ, RSP_136, false, true},
	//  SD: CMD3 - SEND_RELATIVE_ADDR - get relative card address (RCA)
	// MMC: CMD3 -  SET_RELATIVE_ADDR - set relative card address (RCA
	3: {READ, RSP_48, true, true},
	// CMD6 - SWITCH - switch mode of operation
	6: {READ, RSP_48_CHECK_BUSY, true, true},
	// CMD7 - SELECT/DESELECT CARD - enter transfer state
	7: {READ, RSP_48_CHECK_BUSY, true, true},
	//  SD: CMD8 - SEND_IF_COND - read device data
	// MMC: CMD8 - SEND_EXT_CSD - read extended device data
	8: {READ, RSP_48, true, true},
	// CMD9 - SEND_CSD - read device data
	9: {READ, RSP_136, false, true},
	// SD: CMD11 - VOLTAGE_SWITCH - switch to 1.8V signaling
	11: {READ, RSP_48, true, true},
	// CMD12 - STOP_TRANSMISSION - stop transmission
	12: {READ, RSP_NONE, true, true},
	// CMD13 - SEND_STATUS - poll card status
	13: {READ, RSP_48, true, true},
	// CMD16 - SET_BLOCKLEN - define the block length
	16: {READ, RSP_48, true, true},
	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
	18: {READ, RSP_48, true, true},
	// CMD19 - send tuning block command, ignore responses
	19: {READ, RSP_48, true, true},
	// CMD23 - SET_BLOCK_COUNT - define read/write block count
	23: {READ, RSP_48, true, true},
	// CMD25 - WRITE_MULTIPLE_BLOCK - write consecutive blocks
	25: {WRITE, RSP_48, true, true},
	// SD: ACMD41 - SD_SEND_OP_COND - read capacity information
	41: {READ, RSP_48, false, false},
	// SD: CMD55 - APP_CMD - next command is application specific
	55: {READ, RSP_48, true, true},
}

// cmd sends an SD / MMC command as described in
// p349, 35.4.3 Send command to card flow chart, IMX6FG
func (hw *USDHC) cmd(index uint32, arg uint32, blocks uint32, timeout time.Duration) (err error) {
	params, ok := cmds[index]

	if !ok {
		return fmt.Errorf("CMD%d unsupported", index)
	}

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
	if blocks > 0 && !reg.WaitFor(timeout, hw.pres_state, PRES_STATE_CDIHB, 1, 0) {
		return fmt.Errorf("CMD%d data inhibit", index)
	}

	// clear interrupts status
	reg.Write(hw.int_status, 0xffffffff)

	if params.dtd == WRITE && reg.Get(hw.pres_state, PRES_STATE_WPSPL, 1) == 0 {
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

	if blocks > 0 {
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
	bits.SetTo(&xfr, CMD_XFR_TYP_CICEN, params.cic)
	// CRC verification
	bits.SetTo(&xfr, CMD_XFR_TYP_CCCEN, params.ccc)
	// dual data rate
	bits.SetTo(&mix, MIX_CTRL_DDR_EN, hw.card.DDR)

	// command completion
	int_status := INT_STATUS_CC

	if blocks > 0 {
		// transfer completion
		int_status = INT_STATUS_TC
		// enable data presence
		bits.Set(&xfr, CMD_XFR_TYP_DPSEL)
		// enable DMA
		bits.Set(&mix, MIX_CTRL_DMAEN)
		// enable automatic CMD12 to stop transactions
		bits.Set(&mix, MIX_CTRL_AC12EN)
		// multiple blocks
		bits.SetTo(&mix, MIX_CTRL_MSBSEL, blocks > 1)
		// block count
		bits.SetTo(&mix, MIX_CTRL_BCEN, blocks > 1)
	} else {
		bits.Clear(&xfr, CMD_XFR_TYP_DPSEL)
		bits.Clear(&mix, MIX_CTRL_AC12EN)
		bits.Clear(&mix, MIX_CTRL_BCEN)
		bits.Clear(&mix, MIX_CTRL_DMAEN)
		bits.Clear(&mix, MIX_CTRL_MSBSEL)
	}

	// set response type
	bits.SetN(&xfr, CMD_XFR_TYP_RSPTYP, 0b11, params.res)
	// set data transfer direction
	bits.SetN(&mix, MIX_CTRL_DTDSEL, 1, params.dtd)

	reg.Write(hw.mix_ctrl, mix)
	reg.Write(hw.cmd_xfr, xfr)

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
		if err := hw.cmd(12, 0, 0, hw.writeTimeout); err == nil {
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
		if err = hw.cmd(13, hw.rca, 0, hw.writeTimeout); err != nil {
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
