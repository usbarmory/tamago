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

func instructionFrame(command Command) (frame uint32, err error) {
	if command.Width > QuadCommand || command.DummyCycles > 31 {
		return 0, ErrInvalidCommand
	}

	var addressLength uint32

	switch command.AddressBits {
	case 8, 16, 24, 32:
		addressLength = uint32(command.AddressBits/8 - 1)
	default:
		return 0, ErrInvalidCommand
	}

	bits.SetN(&frame, IFR_WIDTH, 0xf, uint32(command.Width))
	bits.SetN(&frame, IFR_NBDUM, 0x1f, uint32(command.DummyCycles))
	bits.SetN(&frame, IFR_ADDRL, 0b11, addressLength)
	bits.Set(&frame, IFR_TFRTYP)
	bits.Set(&frame, IFR_DATAEN)
	bits.Set(&frame, IFR_ADDREN)
	bits.Set(&frame, IFR_INSTEN)

	return
}

// ReadMemory reads serial memory at the argument offset through the
// memory-mapped aperture, using DMA when configured.
func (session *Session) ReadMemory(command Command, offset uint32, buf []byte) (err error) {
	if session.hw == nil {
		return ErrSessionClosed
	}

	return session.hw.readMemory(command, offset, buf)
}

func (hw *QSPI) memoryFrame(command Command, offset uint32, size int) (frame uint32, err error) {
	if hw.MMAP == 0 {
		return 0, ErrInvalidInstance
	}

	end := uint64(offset) + uint64(size)
	if uint64(hw.MMAP)+end > 1<<32 {
		return 0, ErrRange
	}

	if hw.MMAPSize != 0 && end > uint64(hw.MMAPSize) {
		return 0, ErrRange
	}

	if frame, err = instructionFrame(command); err != nil {
		return
	}

	if end > uint64(1)<<command.AddressBits {
		return 0, ErrCommandAddress
	}

	return
}

func (hw *QSPI) readMemory(command Command, offset uint32, buf []byte) (err error) {
	if !hw.ready {
		return ErrNotInitialized
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
			err = hw.invalidateTransfer(err)
		}
	}()

	reg.Write(hw.Base+QSPI_RICR, uint32(command.Instruction))
	if err := hw.changeInstructionFrame(frame); err != nil {
		return err
	}

	if hw.DMA != nil && len(buf) >= dmaMinSize {
		if err = hw.readDMA(buf, hw.MMAP+offset); err != nil {
			return
		}
	} else {
		copyFromMMAP(buf, hw.MMAP+offset)
	}

	return hw.finishTransfer()
}

func (hw *QSPI) readDMA(buf []byte, src uint32) (err error) {
	for done := 0; done < len(buf); {
		n := min(len(buf)-done, dmaMaxSize)

		if err = hw.readDMAChunk(buf[done:done+n], src+uint32(done)); err != nil {
			return fmt.Errorf("DMA read, %w", err)
		}

		done += n
	}

	return
}

func (hw *QSPI) readDMAChunk(buf []byte, src uint32) (err error) {
	// returns a buffer reserved from Region as is, copies any other
	addr := hw.Region.Alloc(buf, 0)
	defer hw.Region.Free(addr)

	if err = hw.DMA.Copy(hw.DMAChannel, uint32(addr), src, len(buf)); err != nil {
		return
	}

	hw.Region.Read(addr, 0, buf)

	return
}

func copyFromMMAP(dst []byte, src uint32) {
	i := 0

	for ; i < len(dst) && (src+uint32(i))&7 != 0; i++ {
		dst[i] = reg.Read8(src + uint32(i))
	}

	for ; i+8 <= len(dst); i += 8 {
		binary.LittleEndian.PutUint64(dst[i:], reg.Read64(uint64(src)+uint64(i)))
	}

	for ; i < len(dst); i++ {
		dst[i] = reg.Read8(src + uint32(i))
	}
}

// ReadCommand issues a single width command, with optional address, and reads
// its response in buf.
func (session *Session) ReadCommand(command Command, address uint32, buf []byte) (err error) {
	if session.hw == nil {
		return ErrSessionClosed
	}

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

	var frame uint32
	bits.SetN(&frame, IFR_WIDTH, 0xf, uint32(command.Width))
	bits.SetN(&frame, IFR_NBDUM, 0x1f, uint32(command.DummyCycles))
	bits.Set(&frame, IFR_SMRM)
	bits.Set(&frame, IFR_APBTFRTYP)
	bits.Set(&frame, IFR_INSTEN)

	if command.AddressBits != 0 {
		bits.SetN(&frame, IFR_ADDRL, 0b11, uint32(command.AddressBits/8-1))
		bits.Set(&frame, IFR_ADDREN)
	}

	if len(buf) != 0 {
		bits.Set(&frame, IFR_DATAEN)
	}

	reg.Write(session.hw.Base+QSPI_IAR, address)
	reg.Write(session.hw.Base+QSPI_RICR, uint32(command.Instruction))

	// frames without data use WICR
	reg.Write(session.hw.Base+QSPI_WICR, uint32(command.Instruction))

	if err := session.hw.changeInstructionFrame(frame); err != nil {
		return fmt.Errorf("read command 0x%02x, %w", command.Instruction, err)
	}

	reg.Read(session.hw.Base + QSPI_ISR)

	if len(buf) == 0 {
		session.hw.command(CR_STTFR)

		if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "read command completion"); err != nil {
			return fmt.Errorf("read command 0x%02x, %w", command.Instruction, err)
		}

		return nil
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "read command synchronization"); err != nil {
		return fmt.Errorf("read command 0x%02x, %w", command.Instruction, err)
	}

	session.hw.command(CR_STTFR)

	for i := range buf {
		if err := session.hw.waitSet(QSPI_ISR, ISR_RDRF, "read data"); err != nil {
			return fmt.Errorf("read command 0x%02x byte %d, %w", command.Instruction, i, err)
		}

		if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "read data synchronization"); err != nil {
			return fmt.Errorf("read command 0x%02x byte %d, %w", command.Instruction, i, err)
		}

		if i == len(buf)-1 {
			session.hw.command(CR_LASTXFER)

			if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "final read synchronization"); err != nil {
				return fmt.Errorf("read command 0x%02x, %w", command.Instruction, err)
			}
		}

		buf[i] = byte(reg.Read(session.hw.Base + QSPI_RDR))
	}

	if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "read command completion"); err != nil {
		return fmt.Errorf("read command 0x%02x, %w", command.Instruction, err)
	}

	return nil
}
