// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package qspi

import (
	"encoding/binary"
	"fmt"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// WriteMemory writes buf to serial memory at the argument offset through the
// memory-mapped aperture.
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
			err = hw.invalidateTransfer(fmt.Errorf("write memory 0x%02x, %w", command.Instruction, err))
		}
	}()

	// the controller counts written bytes to detect the last access
	reg.Write(hw.Base+QSPI_WICR, uint32(command.Instruction))
	reg.Write(hw.Base+QSPI_WRACNT, uint32(len(buf)))

	if err := hw.changeInstructionFrame(frame); err != nil {
		return err
	}

	// clear status
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

func copyToMMAP(dst uint32, src []byte) {
	i := 0

	for ; i < len(src) && (dst+uint32(i))&3 != 0; i++ {
		reg.Write8(dst+uint32(i), src[i])
	}

	for ; i+4 <= len(src); i += 4 {
		reg.Write(dst+uint32(i), binary.LittleEndian.Uint32(src[i:]))
	}

	for ; i < len(src); i++ {
		reg.Write8(dst+uint32(i), src[i])
	}
}

func (session *Session) writeByte(instruction uint8, index int, value byte) (err error) {
	if err := session.hw.waitSet(QSPI_ISR, ISR_TDRE, "write data"); err != nil {
		return fmt.Errorf("write command 0x%02x byte %d, %w", instruction, index, err)
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "write data synchronization"); err != nil {
		return fmt.Errorf("write command 0x%02x byte %d, %w", instruction, index, err)
	}

	reg.Write(session.hw.Base+QSPI_TDR, uint32(value))

	return nil
}

func commandAddress(address uint32, addressBits int) (buf [4]byte, offset int) {
	offset = len(buf) - addressBits/8

	for i := len(buf) - 1; i >= offset; i-- {
		buf[i] = byte(address)
		address >>= 8
	}

	return
}

// WriteCommand issues a single width command, with optional address, followed
// by buf.
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

	var frame uint32
	bits.SetN(&frame, IFR_WIDTH, 0xf, uint32(command.Width))
	bits.Set(&frame, IFR_SMRM)
	bits.Set(&frame, IFR_INSTEN)

	if command.AddressBits != 0 || len(buf) != 0 {
		bits.Set(&frame, IFR_DATAEN)
	}

	reg.Write(session.hw.Base+QSPI_IAR, 0)
	reg.Write(session.hw.Base+QSPI_WICR, uint32(command.Instruction))
	reg.Write(session.hw.Base+QSPI_RICR, 0)

	if err := session.hw.changeInstructionFrame(frame); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", command.Instruction, err)
	}

	reg.Read(session.hw.Base + QSPI_ISR)

	if command.AddressBits == 0 && len(buf) == 0 {
		session.hw.command(CR_STTFR)

		if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
			return fmt.Errorf("write command 0x%02x, %w", command.Instruction, err)
		}

		return nil
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "write command synchronization"); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", command.Instruction, err)
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
		return fmt.Errorf("write command 0x%02x, %w", command.Instruction, err)
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "final write synchronization"); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", command.Instruction, err)
	}

	session.hw.command(CR_LASTXFER)

	if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", command.Instruction, err)
	}

	return nil
}
