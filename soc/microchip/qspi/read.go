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
	"unsafe"

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

	frame = uint32(command.Width)<<IFR_WIDTH |
		uint32(command.DummyCycles)<<IFR_NBDUM |
		addressLength<<IFR_ADDRL |
		1<<IFR_TFRTYP |
		1<<IFR_DATAEN |
		1<<IFR_ADDREN |
		1<<IFR_INSTEN

	return
}

// ReadMemory sends the instruction, an AddressBits-wide aperture offset, and
// DummyCycles before receiving buf with the configured Width in one controller
// transaction, while AddressBits bounds the complete device range.
// Each controller poll uses QSPI.Timeout. On error, a prefix of buf may have
// been replaced and the rest remains unchanged, except that with QSPI.DMA the
// contents of buf are undefined. Serial-memory geometry remains the caller's
// responsibility when MMAPSize is zero.
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

	if hw.DMA != nil {
		if err = hw.readDirect(buf, hw.MMAP+offset); err != nil {
			return
		}
	} else {
		copyFromMMAP(buf, hw.MMAP+offset)
	}

	return hw.finishTransfer()
}

func (hw *QSPI) readDirect(dst []byte, src uint32) (err error) {
	line := hw.Cache.DataCacheLineSize()
	addr := uint(uintptr(unsafe.Pointer(&dst[0])))
	head := min(int(-addr&uint(line-1)), len(dst))
	body := (len(dst) - head) &^ (line - 1)

	if body == 0 || uint64(addr)+uint64(head+body) > 1<<32 {
		copyFromMMAP(dst, src)
		return
	}

	copyFromMMAP(dst[:head], src)

	for done := 0; done < body; {
		n := min(body-done, dmaMaxSize)
		start := addr + uint(head+done)

		// discard lines before the transfer and any refilled during it
		hw.Cache.InvalidateDataCacheRange(start, n)
		err = hw.DMA(uint32(start), src+uint32(head+done), n)
		hw.Cache.InvalidateDataCacheRange(start, n)

		if err != nil {
			return fmt.Errorf("qspi: DMA read, %w", err)
		}

		done += n
	}

	copyFromMMAP(dst[head+body:], src+uint32(head+body))

	return
}

// The caller validates the complete memory-mapped range before copying.
func copyFromMMAP(dst []byte, src uint32) {
	const size = 0x40

	// Align Device memory for arm64 128-bit SIMD loads.
	prefix := min(int(-src&0xf), len(dst))
	copyWordsFromMMAP(dst[:prefix], src)
	dst = dst[prefix:]
	src += uint32(prefix)

	var ptr unsafe.Pointer
	ptr = unsafe.Add(ptr, src)
	mem := unsafe.Slice((*byte)(ptr), len(dst))

	// Fixed-size copies avoid memmove realigning Device accesses.
	for len(dst) >= size {
		copy(dst[:size], mem[:size])
		dst, mem = dst[size:], mem[size:]
		src += size
	}

	copyWordsFromMMAP(dst, src)
}

func copyWordsFromMMAP(dst []byte, src uint32) {
	i := 0

	for ; i < len(dst) && (src+uint32(i))&3 != 0; i++ {
		dst[i] = reg.Read8(src + uint32(i))
	}

	for ; i+4 <= len(dst); i += 4 {
		value := reg.Read(src + uint32(i))
		dst[i+0] = byte(value)
		dst[i+1] = byte(value >> 8)
		dst[i+2] = byte(value >> 16)
		dst[i+3] = byte(value >> 24)
	}

	for ; i < len(dst); i++ {
		dst[i] = reg.Read8(src + uint32(i))
	}
}

// ReadCommand sends a single-width instruction, optional address, and dummy
// cycles through the serial-memory APB path, then fills buf. Each controller
// poll uses QSPI.Timeout. On error, a prefix of buf may have been replaced; the
// rest remains unchanged.
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

	frame := uint32(command.Width)<<IFR_WIDTH |
		uint32(command.DummyCycles)<<IFR_NBDUM |
		1<<IFR_SMRM |
		1<<IFR_APBTFRTYP |
		1<<IFR_INSTEN

	if command.AddressBits != 0 {
		frame |= uint32(command.AddressBits/8-1)<<IFR_ADDRL | 1<<IFR_ADDREN
	}

	if len(buf) != 0 {
		frame |= 1 << IFR_DATAEN
	}

	reg.Write(session.hw.Base+QSPI_IAR, address)
	reg.Write(session.hw.Base+QSPI_RICR, uint32(command.Instruction))

	// Frames without data use WICR even on the read path.
	reg.Write(session.hw.Base+QSPI_WICR, uint32(command.Instruction))

	if err := session.hw.changeInstructionFrame(frame); err != nil {
		return fmt.Errorf("qspi: read command 0x%02x, %w", command.Instruction, err)
	}

	reg.Read(session.hw.Base + QSPI_ISR)

	if len(buf) == 0 {
		session.hw.command(CR_STTFR)

		if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "read command completion"); err != nil {
			return fmt.Errorf("qspi: read command 0x%02x, %w", command.Instruction, err)
		}

		return nil
	}

	if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "read command synchronization"); err != nil {
		return fmt.Errorf("qspi: read command 0x%02x, %w", command.Instruction, err)
	}

	session.hw.command(CR_STTFR)

	for i := range buf {
		if err := session.hw.waitSet(QSPI_ISR, ISR_RDRF, "read data"); err != nil {
			return fmt.Errorf("qspi: read command 0x%02x byte %d, %w", command.Instruction, i, err)
		}

		if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "read data synchronization"); err != nil {
			return fmt.Errorf("qspi: read command 0x%02x byte %d, %w", command.Instruction, i, err)
		}

		if i == len(buf)-1 {
			session.hw.command(CR_LASTXFER)

			if err := session.hw.waitClear(QSPI_SR, SR_SYNCBSY, "final read synchronization"); err != nil {
				return fmt.Errorf("qspi: read command 0x%02x, %w", command.Instruction, err)
			}
		}

		buf[i] = byte(reg.Read(session.hw.Base + QSPI_RDR))
	}

	if err := session.hw.waitSet(QSPI_ISR, ISR_CSRA, "read command completion"); err != nil {
		return fmt.Errorf("qspi: read command 0x%02x, %w", command.Instruction, err)
	}

	return nil
}
