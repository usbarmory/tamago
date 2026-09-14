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

// WriteMemory sends the instruction and an AddressBits-wide aperture offset,
// then buf with the configured Width, in one controller transaction through the
// memory-mapped aperture. Commands with dummy cycles are rejected. Each
// controller poll uses QSPI.Timeout. The method does not wait for a subsequent
// device operation. On error, a prefix may have reached the device, whose state
// is indeterminate; the transfer is never retried automatically.
func (session *Session) WriteMemory(command Command, offset uint32, buf []byte) (err error) {
	if session.hw == nil {
		return ErrSessionClosed
	}

	return session.hw.writeMemory(command, offset, buf)
}

func (hw *QSPI) writeMemory(command Command, offset uint32, buf []byte) (err error) {
	if !hw.ready {
		return ErrNotInitialized
	}

	if command.DummyCycles != 0 {
		return ErrInvalidCommand
	}

	if len(buf) == 0 {
		return nil
	}

	frame, err := hw.memoryFrame(command, offset, len(buf))
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = hw.invalidateTransfer(fmt.Errorf("qspi: write memory 0x%02x, %w", command.Instruction, err))
		}
	}()

	// the controller counts written bytes to detect the last access
	reg.Write(hw.Base+QSPI_WICR, uint32(command.Instruction))
	reg.Write(hw.Base+QSPI_WRACNT, uint32(len(buf)))

	if err := hw.changeInstructionFrame(frame); err != nil {
		return err
	}

	// consume stale status
	reg.Read(hw.Base + QSPI_ISR)

	copyToMMAP(hw.MMAP+offset, buf)

	if err := hw.waitSet(QSPI_ISR, ISR_LWRA, "last write access"); err != nil {
		return err
	}

	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "write synchronization"); err != nil {
		return err
	}

	hw.command(CR_LASTXFER)

	return hw.waitSet(QSPI_ISR, ISR_CSRA, "write completion")
}

// The caller validates the complete memory-mapped range before copying.
func copyToMMAP(dst uint32, src []byte) {
	i := 0

	for ; i < len(src) && (dst+uint32(i))&3 != 0; i++ {
		reg.Write8(dst+uint32(i), src[i])
	}

	for ; i+4 <= len(src); i += 4 {
		reg.Write(dst+uint32(i), uint32(src[i])|uint32(src[i+1])<<8|uint32(src[i+2])<<16|uint32(src[i+3])<<24)
	}

	for ; i < len(src); i++ {
		reg.Write8(dst+uint32(i), src[i])
	}
}

func (session *Session) writeByte(instruction uint8, index int, value byte) (err error) {
	if err := session.hw.waitSet(QSPI_ISR, ISR_TDRE, "write data"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x byte %d, %w", instruction, index, err)
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "write data synchronization"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x byte %d, %w", instruction, index, err)
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
	if session.hw == nil {
		return ErrSessionClosed
	}

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
		return fmt.Errorf("qspi: write command 0x%02x, %w", command.Instruction, err)
	}

	reg.Read(session.hw.Base + QSPI_ISR)

	if command.AddressBits == 0 && len(buf) == 0 {
		session.hw.command(CR_STTFR)

		if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
			return fmt.Errorf("qspi: write command 0x%02x, %w", command.Instruction, err)
		}

		return nil
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "write command synchronization"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x, %w", command.Instruction, err)
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
		return fmt.Errorf("qspi: write command 0x%02x, %w", command.Instruction, err)
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "final write synchronization"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x, %w", command.Instruction, err)
	}

	session.hw.command(CR_LASTXFER)

	if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
		return fmt.Errorf("qspi: write command 0x%02x, %w", command.Instruction, err)
	}

	return nil
}
