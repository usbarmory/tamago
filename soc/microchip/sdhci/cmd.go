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

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

var (
	// CommandInhibitTimeout controls how long a command waits for the host
	// command and data paths to become idle.
	CommandInhibitTimeout = 200 * time.Millisecond
	// CommandTimeout controls how long a command waits for completion.
	CommandTimeout = 500 * time.Millisecond
)

type cmdParams struct {
	responseType uint16
	commandType  uint16
	openDrain    bool
	indexCheck   bool
	crcCheck     bool
	dataPresent  bool
}

var cmds = map[uint16]cmdParams{
	// CMD0 - GO_IDLE_STATE - reset card
	0: {responseType: CR_RESPTYP_NORESP, openDrain: true},
	// CMD1 - SEND_OP_COND - send operating conditions
	1: {responseType: CR_RESPTYP_RL48, openDrain: true},
	// CMD2 - ALL_SEND_CID - get unique card identification
	2: {responseType: CR_RESPTYP_RL136, openDrain: true, crcCheck: true},
	// CMD3 - SET_RELATIVE_ADDR - set relative card address
	3: {responseType: CR_RESPTYP_RL48, openDrain: true, indexCheck: true, crcCheck: true},
	// CMD6 - SWITCH - switch mode of operation
	6: {responseType: CR_RESPTYP_RL48BSY, indexCheck: true, crcCheck: true},
	// CMD7 - SELECT/DESELECT_CARD - enter or leave transfer state
	7: {responseType: CR_RESPTYP_RL48BSY, indexCheck: true, crcCheck: true},
	// CMD8 - SEND_EXT_CSD - read extended device data
	8: {responseType: CR_RESPTYP_RL48, indexCheck: true, crcCheck: true, dataPresent: true},
	// CMD9 - SEND_CSD - read device data
	9: {responseType: CR_RESPTYP_RL136, crcCheck: true},
	// CMD12 - STOP_TRANSMISSION - stop a multi-block transfer
	12: {responseType: CR_RESPTYP_RL48BSY, commandType: CR_CMDTYP_ABORT, crcCheck: true},
	// CMD13 - SEND_STATUS - read card status
	13: {responseType: CR_RESPTYP_RL48, indexCheck: true, crcCheck: true},
	// CMD17 - READ_SINGLE_BLOCK - read one block
	17: {responseType: CR_RESPTYP_RL48, indexCheck: true, crcCheck: true, dataPresent: true},
	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
	18: {responseType: CR_RESPTYP_RL48, indexCheck: true, crcCheck: true, dataPresent: true},
	// CMD24 - WRITE_BLOCK - write one block
	24: {responseType: CR_RESPTYP_RL48, indexCheck: true, crcCheck: true, dataPresent: true},
}

func commandValue(index uint16, params cmdParams) (command uint16) {
	bits.SetN16(&command, CR_RESPTYP, CR_RESPTYP_MASK, params.responseType)
	bits.SetN16(&command, CR_CMDTYP, CR_CMDTYP_MASK, params.commandType)
	bits.SetTo16(&command, CR_CMDICEN, params.indexCheck)
	bits.SetTo16(&command, CR_CMDCCEN, params.crcCheck)
	bits.SetTo16(&command, CR_DPSEL, params.dataPresent)
	bits.SetN16(&command, CR_CMDIDX, CR_CMDIDX_MASK, index)

	return
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

	// clear pending status
	reg.Write16(hw.eistr, allInterrupts)
	reg.Write16(hw.nistr, allInterrupts)

	usesDataLine := params.responseType == CR_RESPTYP_RL48BSY || params.dataPresent

	var inhibitMask uint32
	bits.Set(&inhibitMask, PSR_CMDINHC)

	if params.commandType != CR_CMDTYP_ABORT && usesDataLine {
		bits.Set(&inhibitMask, PSR_CMDINHD)
	}

	if !reg.WaitFor(CommandInhibitTimeout, hw.psr, 0, int(inhibitMask), 0) {
		var resetMask uint16
		bits.Set16(&resetMask, SRR_SWRSTCMD)

		if usesDataLine {
			bits.Set16(&resetMask, SRR_SWRSTDAT)
		}

		if resetErr := hw.reset(uint8(resetMask), ControllerSetupTimeout); resetErr != nil {
			return 0, 0, false, fmt.Errorf("sdhci: command inhibit timeout (recovery failed: %v)", resetErr)
		}

		return 0, 0, false, fmt.Errorf("CMD%d command inhibit timeout", index)
	}

	// configure command signaling
	mode := uint16(reg.Read8(hw.mc1r))
	bits.SetN16(&mode, MC1R_CMDTYP, MC1R_CMDTYP_MASK, 0)
	bits.SetTo16(&mode, MC1R_OPD, params.openDrain)
	bits.Set16(&mode, MC1R_FCD)
	reg.Write8(hw.mc1r, uint8(mode))

	// issue card command
	reg.Write(hw.arg1r, argument)
	reg.Write16(hw.cr, commandValue(index, params))

	value, ignored, err = hw.pollStatusIgnoring(1<<NISTR_CMDC, CommandTimeout, ignoredErrors)
	if err != nil {
		var resetMask uint16
		bits.Set16(&resetMask, SRR_SWRSTCMD)

		if usesDataLine {
			bits.Set16(&resetMask, SRR_SWRSTDAT)
		}

		if resetErr := hw.reset(uint8(resetMask), ControllerSetupTimeout); resetErr != nil {
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

		if bits.Get16(&status, NISTR_ERRINT) {
			errorStatus := reg.Read16(hw.eistr)

			// clear detailed errors before the summary
			reg.Write16(hw.eistr, errorStatus)
			reg.Write16(hw.nistr, 1<<NISTR_ERRINT)

			if errorStatus&^ignoredErrors != 0 {
				if bits.Get16(&errorStatus, EISTR_ADMA) {
					return 0, ignored, fmt.Errorf("sdhci: ADMA2 interrupt 0x%04x (%s)", errorStatus|ignored, hw.dumpRegisters())
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
			return 0, ignored, fmt.Errorf("sdhci: status 0x%04x timeout (%s)", expected, hw.dumpRegisters())
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

	if resetErr := hw.reset(1<<SRR_SWRSTCMD|1<<SRR_SWRSTDAT, ControllerSetupTimeout); resetErr != nil {
		return fmt.Errorf("%w (controller recovery failed: %v)", err, resetErr)
	}

	return fmt.Errorf("%w (controller requires reinitialization)", err)
}

func (hw *SDHCI) invalidateStop(transferErr error, stopErr error) error {
	recoveryErr := fmt.Errorf("CMD12 STOP_TRANSMISSION failed: %w", stopErr)

	if transferErr != nil {
		recoveryErr = fmt.Errorf("%w (transfer error: %v)", recoveryErr, transferErr)
	}

	return hw.invalidateTransfer(recoveryErr)
}

func (hw *SDHCI) stopTransmission(transferErr error) error {
	status, ignored, _, stopErr := hw.runCommand(12, 0, EISTR_DAT_LINE_ERROR_MASK)
	if stopErr != nil {
		return hw.invalidateStop(transferErr, stopErr)
	}

	if ignored != 0 {
		if stopErr = hw.reset(1<<SRR_SWRSTDAT, ControllerSetupTimeout); stopErr != nil {
			return hw.invalidateStop(transferErr, stopErr)
		}
	}

	if stopErr = checkR1(status); stopErr != nil {
		return hw.invalidateStop(transferErr, stopErr)
	}

	if stopErr = hw.waitState(CURRENT_STATE_TRAN, ControllerSetupTimeout); stopErr != nil {
		return hw.invalidateStop(transferErr, stopErr)
	}

	return transferErr
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

		if bits.Get(&status, STATUS_READY_FOR_DATA) && bits.GetN(&status, STATUS_CURRENT_STATE, STATUS_CURRENT_STATE_MASK) == uint32(state) {
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
