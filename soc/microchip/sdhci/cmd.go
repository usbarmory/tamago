// Microchip Secure Digital Host Controller Interface support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sdhci

import (
	"fmt"
	"runtime"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	commandInhibitTimeout = 200 * time.Millisecond
	commandTimeout        = 500 * time.Millisecond
)

type cmdParams struct {
	response uint16
	mc1r     byte
}

var cmds = map[uint16]cmdParams{
	// CMD0 - GO_IDLE_STATE - reset card
	0: {response: CR_RESPTYP_NORESP << CR_RESPTYP, mc1r: 1 << MC1R_OPD},
	// CMD1 - SEND_OP_COND - send operating conditions
	1: {response: CR_RESPTYP_RL48 << CR_RESPTYP, mc1r: 1 << MC1R_OPD},
	// CMD2 - ALL_SEND_CID - get unique card identification
	2: {response: CR_RESPTYP_RL136<<CR_RESPTYP | 1<<CR_CMDCCEN, mc1r: 1 << MC1R_OPD},
	// CMD3 - SET_RELATIVE_ADDR - set relative card address
	3: {response: CR_RESPTYP_RL48<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN, mc1r: 1 << MC1R_OPD},
	// CMD6 - SWITCH - switch mode of operation
	6: {response: CR_RESPTYP_RL48BSY<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN},
	// CMD7 - SELECT/DESELECT_CARD - enter or leave transfer state
	7: {response: CR_RESPTYP_RL48BSY<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN},
	// CMD8 - SEND_EXT_CSD - read extended device data
	8: {response: CR_RESPTYP_RL48<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN | 1<<CR_DPSEL},
	// CMD9 - SEND_CSD - read device data
	9: {response: CR_RESPTYP_RL136<<CR_RESPTYP | 1<<CR_CMDCCEN},
	// CMD12 - STOP_TRANSMISSION - stop a multi-block transfer
	12: {response: CR_RESPTYP_RL48BSY<<CR_RESPTYP | 1<<CR_CMDCCEN | CR_CMDTYP_ABORT<<CR_CMDTYP},
	// CMD13 - SEND_STATUS - read card status
	13: {response: CR_RESPTYP_RL48<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN},
	// CMD17 - READ_SINGLE_BLOCK - read one block
	17: {response: CR_RESPTYP_RL48<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN | 1<<CR_DPSEL},
	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
	18: {response: CR_RESPTYP_RL48<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN | 1<<CR_DPSEL},
	// CMD24 - WRITE_BLOCK - write one block
	24: {response: CR_RESPTYP_RL48<<CR_RESPTYP | 1<<CR_CMDICEN | 1<<CR_CMDCCEN | 1<<CR_DPSEL},
}

func (hw *SDHCI) cmd(index uint16, argument uint32) (uint32, error) {
	value, _, _, err := hw.runCommand(index, argument, 0)

	return value, err
}

func (hw *SDHCI) runCommand(index uint16, argument uint32, ignoredErrors uint16) (value uint32, ignored uint16, issued bool, err error) {
	params, ok := cmds[index]
	if !ok {
		return 0, 0, false, fmt.Errorf("sdhci: unsupported command CMD%d", index)
	}

	reg.Write16(hw.eistr, allInterrupts)
	reg.Write16(hw.nistr, allInterrupts)

	responseType := params.response >> CR_RESPTYP & CR_RESPTYP_MASK
	commandType := params.response >> CR_CMDTYP & CR_CMDTYP_ABORT
	usesDataLine := responseType == CR_RESPTYP_RL48BSY || params.response&1<<CR_DPSEL != 0
	inhibitMask := 1 << PSR_CMDINHC

	if commandType != CR_CMDTYP_ABORT && usesDataLine {
		inhibitMask |= 1 << PSR_CMDINHD
	}

	if !reg.WaitFor(commandInhibitTimeout, hw.psr, 0, inhibitMask, 0) {
		mask := uint8(1 << SRR_SWRSTCMD)

		if usesDataLine {
			mask |= 1 << SRR_SWRSTDAT
		}

		if resetErr := hw.reset(mask, controllerSetupTimeout); resetErr != nil {
			return 0, 0, false, fmt.Errorf("sdhci: command inhibit timeout (recovery failed: %v)", resetErr)
		}

		return 0, 0, false, fmt.Errorf("CMD%d command inhibit timeout", index)
	}

	mc1r := reg.Read8(hw.mc1r) &^ (MC1R_CMDTYP_MASK<<MC1R_CMDTYP | 1<<MC1R_OPD)
	reg.Write8(hw.mc1r, mc1r|params.mc1r|1<<MC1R_FCD)
	reg.Write(hw.arg1r, argument)
	reg.Write16(hw.cr, index<<CR_CMDIDX|params.response)

	value, ignored, err = hw.pollStatusIgnoring(1<<NISTR_CMDC, commandTimeout, ignoredErrors)
	if err != nil {
		mask := uint8(1 << SRR_SWRSTCMD)

		if usesDataLine {
			mask |= 1 << SRR_SWRSTDAT
		}

		if resetErr := hw.reset(mask, controllerSetupTimeout); resetErr != nil {
			err = fmt.Errorf("%w (command recovery failed: %v)", err, resetErr)
		}
	}

	return value, ignored, true, err
}

func (hw *SDHCI) pollStatus(expected uint16, timeout time.Duration) (uint32, error) {
	value, _, err := hw.pollStatusIgnoring(expected, timeout, 0)

	return value, err
}

func (hw *SDHCI) pollStatusIgnoring(expected uint16, timeout time.Duration, ignoredErrors uint16) (value uint32, ignored uint16, err error) {
	deadline := time.Now().Add(timeout)

	for {
		status := reg.Read16(hw.nistr)

		if status&1<<NISTR_ERRINT != 0 {
			errorStatus := reg.Read16(hw.eistr)

			// Clear detailed errors before the summary.
			reg.Write16(hw.eistr, errorStatus)
			reg.Write16(hw.nistr, 1<<NISTR_ERRINT)

			if errorStatus&^ignoredErrors != 0 {
				if errorStatus&1<<EISTR_ADMA != 0 {
					return 0, ignored, fmt.Errorf("sdhci: ADMA2 error 0x%02x interrupt=0x%04x", reg.Read8(hw.aesr), errorStatus|ignored)
				}

				return 0, ignored, fmt.Errorf("sdhci: interrupt error 0x%04x", errorStatus|ignored)
			}

			ignored |= errorStatus
		}

		if status&expected != 0 {
			reg.Write16(hw.nistr, expected)
			return reg.Read(hw.rr), ignored, nil
		}

		if time.Now().After(deadline) {
			return 0, ignored, fmt.Errorf("sdhci: status timeout waiting for 0x%04x nistr=0x%04x eistr=0x%04x psr=0x%08x tmr=0x%04x bcr=0x%04x ccr=0x%04x", expected, status, reg.Read16(hw.eistr), reg.Read(hw.psr), reg.Read16(hw.tmr), reg.Read16(hw.bcr), reg.Read16(hw.ccr))
		}

		runtime.Gosched()
	}
}

func (hw *SDHCI) response136() [4]uint32 {
	return [4]uint32{
		reg.Read(hw.rr),
		reg.Read(hw.rr + 4),
		reg.Read(hw.rr + 8),
		reg.Read(hw.rr + 12),
	}
}

func (hw *SDHCI) invalidateTransfer(err error) error {
	hw.controllerReady = false
	hw.ready = false

	if resetErr := hw.reset(1<<SRR_SWRSTCMD|1<<SRR_SWRSTDAT, controllerSetupTimeout); resetErr != nil {
		return fmt.Errorf("%w (controller recovery failed: %v)", err, resetErr)
	}

	return fmt.Errorf("%w (controller requires reinitialization)", err)
}

func (hw *SDHCI) stopTransmission(transferErr error) error {
	status, ignored, _, stopErr := hw.runCommand(12, 0, EISTR_DAT_LINE_ERROR_MASK)

	if stopErr == nil && ignored != 0 {
		stopErr = hw.reset(1<<SRR_SWRSTDAT, controllerSetupTimeout)
	}

	if stopErr == nil {
		stopErr = checkR1(status)
	}

	if stopErr == nil {
		stopErr = hw.waitState(CURRENT_STATE_TRAN, controllerSetupTimeout)
	}

	if stopErr == nil {
		return transferErr
	}

	recoveryErr := fmt.Errorf("CMD12 STOP_TRANSMISSION failed: %w", stopErr)

	if transferErr != nil {
		recoveryErr = fmt.Errorf("%w (transfer error: %v)", recoveryErr, transferErr)
	}

	return hw.invalidateTransfer(recoveryErr)
}

func (hw *SDHCI) waitState(state int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		status, err := hw.cmd(13, uint32(hw.card.RCA)<<16)
		if err != nil {
			return fmt.Errorf("CMD13 SEND_STATUS failed: %w", err)
		}

		if err := checkR1(status); err != nil {
			return fmt.Errorf("CMD13 SEND_STATUS: %w", err)
		}

		if status&(1<<STATUS_READY_FOR_DATA) != 0 && (status>>STATUS_CURRENT_STATE)&STATUS_CURRENT_STATE_MASK == uint32(state) {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("sdhci: card ready timeout status=0x%08x", status)
		}

		runtime.Gosched()
	}
}

func checkR1(response uint32) error {
	if response&r1ErrorMask != 0 {
		return fmt.Errorf("eMMC R1 error status=0x%08x", response)
	}

	return nil
}
