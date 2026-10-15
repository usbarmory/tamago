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
func (s *Session) WriteMemory(cmd Command, offset uint32, buf []byte) (err error) {
	if s.hw == nil {
		return ErrSessionClosed
	}

	return s.hw.writeMemory(cmd, offset, buf)
}

func (hw *QSPI) writeMemory(cmd Command, offset uint32, buf []byte) (err error) {
	if !hw.ready {
		return ErrNotInitialized
	}

	if cmd.DummyCycles != 0 {
		return ErrInvalidCommand
	}

	if len(buf) == 0 {
		return nil
	}

	frame, err := hw.memoryFrame(cmd, offset, len(buf))

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = hw.invalidateTransfer(fmt.Errorf("write memory 0x%02x, %w", cmd.Instruction, err))
		}
	}()

	// the controller counts written bytes to detect the last access
	reg.Write(hw.Base+QSPI_WICR, uint32(cmd.Instruction))
	reg.Write(hw.Base+QSPI_WRACNT, uint32(len(buf)))

	if err = hw.changeInstructionFrame(frame); err != nil {
		return
	}

	// clear status
	reg.Read(hw.Base + QSPI_ISR)

	copyToMMAP(hw.MMAP+offset, buf)

	if err = hw.waitSet(QSPI_ISR, ISR_LWRA, "last write access"); err != nil {
		return
	}

	if err = hw.command(CR_LASTXFER); err != nil {
		return
	}

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

func (s *Session) writeByte(instruction uint8, index int, value byte) (err error) {
	if err := s.hw.waitSet(QSPI_ISR, ISR_TDRE, "write data"); err != nil {
		return fmt.Errorf("write command 0x%02x byte %d, %w", instruction, index, err)
	}

	if err := s.hw.synchronize(); err != nil {
		return fmt.Errorf("write command 0x%02x byte %d, %w", instruction, index, err)
	}

	reg.Write(s.hw.Base+QSPI_TDR, uint32(value))

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
func (s *Session) WriteCommand(cmd Command, address uint32, buf []byte) (err error) {
	if s.hw == nil {
		return ErrSessionClosed
	}

	if !s.hw.ready {
		return ErrNotInitialized
	}

	if err = validateSerialCommand(cmd, true); err != nil {
		return
	}

	if err = validateCommandAddress(cmd, address); err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = s.hw.invalidateTransfer(err)
		}
	}()

	var frame uint32
	bits.SetN(&frame, IFR_WIDTH, 0xf, uint32(cmd.Width))
	bits.Set(&frame, IFR_SMRM)
	bits.Set(&frame, IFR_INSTEN)

	if cmd.AddressBits != 0 || len(buf) != 0 {
		bits.Set(&frame, IFR_DATAEN)
	}

	reg.Write(s.hw.Base+QSPI_IAR, 0)
	reg.Write(s.hw.Base+QSPI_WICR, uint32(cmd.Instruction))
	reg.Write(s.hw.Base+QSPI_RICR, 0)

	if err = s.hw.changeInstructionFrame(frame); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", cmd.Instruction, err)
	}

	reg.Read(s.hw.Base + QSPI_ISR)

	if err = s.hw.command(CR_STTFR); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", cmd.Instruction, err)
	}

	if cmd.AddressBits == 0 && len(buf) == 0 {
		if err = s.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
			return fmt.Errorf("write command 0x%02x, %w", cmd.Instruction, err)
		}

		return nil
	}

	addressBytes, offset := commandAddress(address, cmd.AddressBits)
	index := 0

	for _, value := range addressBytes[offset:] {
		if err = s.writeByte(cmd.Instruction, index, value); err != nil {
			return err
		}

		index++
	}

	for _, value := range buf {
		if err = s.writeByte(cmd.Instruction, index, value); err != nil {
			return err
		}

		index++
	}

	if err = s.hw.waitSet(QSPI_ISR, ISR_TXEMPTY, "write drain"); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", cmd.Instruction, err)
	}

	if err = s.hw.command(CR_LASTXFER); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", cmd.Instruction, err)
	}

	if err = s.hw.waitSet(QSPI_ISR, ISR_CSRA, "write command completion"); err != nil {
		return fmt.Errorf("write command 0x%02x, %w", cmd.Instruction, err)
	}

	return nil
}
