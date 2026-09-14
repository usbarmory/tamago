// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package qspi

import (
	"fmt"

	"github.com/usbarmory/tamago/internal/reg"
)

// Session provides exclusive serial-memory access to a QSPI controller. A
// session is valid only for the duration of its [QSPI.Serialize] callback.
type Session struct {
	hw *QSPI
}

// Serialize runs operation with exclusive access to an initialized controller.
// The callback can group related serial-memory commands, such as write enable,
// page program, status polling, and write disable, without interleaving access
// from another goroutine. It must access the controller through Session because
// QSPI methods acquire the same lock. A controller transfer failure invalidates
// the controller; later commands require QSPI.Init outside the callback.
func (hw *QSPI) Serialize(operation func(*Session) error) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if !hw.ready {
		return ErrNotInitialized
	}

	if operation == nil {
		return ErrInvalidOperation
	}

	return operation(&Session{hw: hw})
}

func validateSerialCommand(command Command, write bool) (err error) {
	// Single is the only command-mode layout qualified by this driver.
	if command.Width != Single || command.DummyCycles > 31 {
		return ErrInvalidCommand
	}

	switch command.AddressBits {
	case 0, 8, 16, 24, 32:
	default:
		return ErrInvalidCommand
	}

	// Command-mode writes emit the address through the APB data path. Dummy
	// cycles are therefore unsupported for writes.
	if write && command.DummyCycles != 0 {
		return ErrInvalidCommand
	}

	return nil
}

func validateCommandAddress(command Command, address uint32) (err error) {
	if command.AddressBits == 0 {
		if address != 0 {
			return ErrCommandAddress
		}

		return nil
	}

	if command.AddressBits < 32 && address >= 1<<command.AddressBits {
		return ErrCommandAddress
	}

	return nil
}

// ReadCommand sends a single-width instruction, optional address, and dummy
// cycles through the serial-memory APB path, then fills buf. Each controller
// poll uses QSPI.Timeout. On error, a prefix of buf may have been replaced; the
// rest remains unchanged.
func (session *Session) ReadCommand(command Command, address uint32, buf []byte) (err error) {
	if !session.hw.ready {
		return ErrNotInitialized
	}

	if err := validateSerialCommand(command, false); err != nil {
		return err
	}

	if err := validateCommandAddress(command, address); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = session.hw.invalidateTransfer(err)
		}
	}()

	frame := uint32(command.Width)<<IFR_WIDTH | 1<<IFR_SMRM | 1<<IFR_APBTFRTYP | 1<<IFR_INSTEN

	switch command.AddressBits {
	case 8:
		frame |= 0 << IFR_ADDRL
	case 16:
		frame |= 1 << IFR_ADDRL
	case 24:
		frame |= 2 << IFR_ADDRL
	case 32:
		frame |= 3 << IFR_ADDRL
	}

	if command.AddressBits != 0 {
		frame |= 1 << IFR_ADDREN
	}

	if command.DummyCycles != 0 {
		frame |= uint32(command.DummyCycles) << IFR_NBDUM
	}

	if len(buf) != 0 {
		frame |= 1 << IFR_DATAEN
	}

	reg.Write(session.hw.Base+QSPI_IAR, address)
	reg.Write(session.hw.Base+QSPI_RICR, uint32(command.Instruction))

	// Frames without data use WICR even on the read path.
	reg.Write(session.hw.Base+QSPI_WICR, uint32(command.Instruction))

	if err := session.hw.changeInstructionFrame(frame); err != nil {
		return fmt.Errorf("qspi: read command 0x%02x: %w", command.Instruction, err)
	}

	reg.Read(session.hw.Base + QSPI_ISR)

	if len(buf) == 0 {
		session.hw.command(CR_STTFR)
		if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "read command completion"); err != nil {
			return fmt.Errorf("qspi: read command 0x%02x: %w", command.Instruction, err)
		}

		return nil
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "read command synchronization"); err != nil {
		return fmt.Errorf("qspi: read command 0x%02x: %w", command.Instruction, err)
	}

	session.hw.command(CR_STTFR)

	for i := range buf {
		if err := session.hw.waitSet(QSPI_ISR, ISR_RDRF, "read data"); err != nil {
			return fmt.Errorf("qspi: read command 0x%02x byte %d: %w", command.Instruction, i, err)
		}

		if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "read data synchronization"); err != nil {
			return fmt.Errorf("qspi: read command 0x%02x byte %d: %w", command.Instruction, i, err)
		}

		if i == len(buf)-1 {
			session.hw.command(CR_LASTXFER)
			if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "final read synchronization"); err != nil {
				return fmt.Errorf("qspi: read command 0x%02x: %w", command.Instruction, err)
			}
		}

		buf[i] = byte(reg.Read(session.hw.Base + QSPI_RDR))
	}

	if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "read command completion"); err != nil {
		return fmt.Errorf("qspi: read command 0x%02x: %w", command.Instruction, err)
	}

	return nil
}

func (session *Session) writeByte(instruction uint8, index int, value byte) (err error) {
	if err := session.hw.waitSet(QSPI_ISR, ISR_TDRE, "write data"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x byte %d: %w", instruction, index, err)
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "write data synchronization"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x byte %d: %w", instruction, index, err)
	}

	reg.Write(session.hw.Base+QSPI_TDR, uint32(value))

	return nil
}

func commandAddress(address uint32, bits int) (buf [4]byte, offset int) {
	offset = len(buf) - bits/8

	for i := len(buf) - 1; i >= offset; i-- {
		buf[i] = byte(address)
		address >>= 8
	}

	return
}

// WriteCommand sends a single-width instruction, the optional address most
// significant byte first, and buf through the serial-memory APB path. Each
// controller poll uses QSPI.Timeout. The method does not wait for a subsequent
// device operation. On error, a prefix may have reached the device, whose state
// is indeterminate; the transfer is never retried automatically.
func (session *Session) WriteCommand(command Command, address uint32, buf []byte) (err error) {
	if !session.hw.ready {
		return ErrNotInitialized
	}

	if err := validateSerialCommand(command, true); err != nil {
		return err
	}

	if err := validateCommandAddress(command, address); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = session.hw.invalidateTransfer(err)
		}
	}()

	frame := uint32(command.Width)<<IFR_WIDTH | 1<<IFR_SMRM | 1<<IFR_INSTEN
	if command.AddressBits != 0 || len(buf) != 0 {
		frame |= 1 << IFR_DATAEN
	}

	reg.Write(session.hw.Base+QSPI_IAR, 0)
	reg.Write(session.hw.Base+QSPI_WICR, uint32(command.Instruction))
	reg.Write(session.hw.Base+QSPI_RICR, 0)

	if err := session.hw.changeInstructionFrame(frame); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x: %w", command.Instruction, err)
	}

	reg.Read(session.hw.Base + QSPI_ISR)

	if command.AddressBits == 0 && len(buf) == 0 {
		session.hw.command(CR_STTFR)
		if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
			return fmt.Errorf("qspi: write command 0x%02x: %w", command.Instruction, err)
		}

		return nil
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "write command synchronization"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x: %w", command.Instruction, err)
	}

	session.hw.command(CR_STTFR)

	addressBytes, offset := commandAddress(address, command.AddressBits)
	index := 0

	for _, value := range addressBytes[offset:] {
		if err := session.writeByte(command.Instruction, index, value); err != nil {
			return err
		}

		index++
	}

	for _, value := range buf {
		if err := session.writeByte(command.Instruction, index, value); err != nil {
			return err
		}

		index++
	}

	if err := session.hw.waitSet(QSPI_ISR, ISR_TXEMPTY, "write drain"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x: %w", command.Instruction, err)
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "final write synchronization"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x: %w", command.Instruction, err)
	}

	session.hw.command(CR_LASTXFER)
	if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x: %w", command.Instruction, err)
	}

	return nil
}

// ReadCommand performs the same wire transaction, timeout handling, and
// partial-buffer behavior as [Session.ReadCommand] with exclusive controller
// access.
func (hw *QSPI) ReadCommand(command Command, address uint32, buf []byte) (err error) {
	return hw.Serialize(func(session *Session) error {
		return session.ReadCommand(command, address, buf)
	})
}

// WriteCommand performs the same wire transaction, timeout handling, and
// partial-transfer behavior as [Session.WriteCommand] with exclusive controller
// access.
func (hw *QSPI) WriteCommand(command Command, address uint32, buf []byte) (err error) {
	return hw.Serialize(func(session *Session) error {
		return session.WriteCommand(command, address, buf)
	})
}
