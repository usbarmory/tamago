// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package qspi

import (
	"unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

const readChunkSize = 0x1000

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
// DummyCycles before receiving buf with the configured Width. Reads use 4 KiB
// controller transactions, while AddressBits bounds the complete device range.
// Each controller poll uses QSPI.Timeout. On error, a prefix of buf may have
// been replaced; the rest remains unchanged. Serial-memory geometry remains
// the caller's responsibility when MMAPSize is zero.
func (hw *QSPI) ReadMemory(command Command, offset uint32, buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	return hw.readMemory(command, offset, buf)
}

// ReadMemory performs the same wire transaction, bounds checks, timeout
// handling, and partial-buffer behavior as [QSPI.ReadMemory] while retaining
// the session's controller lock.
func (session *Session) ReadMemory(command Command, offset uint32, buf []byte) (err error) {
	return session.hw.readMemory(command, offset, buf)
}

func (hw *QSPI) readMemory(command Command, offset uint32, buf []byte) (err error) {
	if !hw.ready {
		return ErrNotInitialized
	}

	if len(buf) == 0 {
		return nil
	}

	if hw.MMAP == 0 {
		return ErrInvalidInstance
	}

	end := uint64(offset) + uint64(len(buf))
	if end > 1<<32 || uint64(hw.MMAP)+end > 1<<32 {
		return ErrRange
	}

	if hw.MMAPSize != 0 && end > uint64(hw.MMAPSize) {
		return ErrRange
	}

	frame, err := instructionFrame(command)
	if err != nil {
		return err
	}

	if end > uint64(1)<<command.AddressBits {
		return ErrCommandAddress
	}

	defer func() {
		if err != nil {
			err = hw.invalidateTransfer(err)
		}
	}()

	for done := 0; done < len(buf); {
		chunk := len(buf) - done
		if chunk > readChunkSize {
			chunk = readChunkSize
		}

		reg.Write(hw.Base+QSPI_RICR, uint32(command.Instruction))
		if err := hw.changeInstructionFrame(frame); err != nil {
			return err
		}

		address := hw.MMAP + offset + uint32(done)
		copyFromMMAP(buf[done:done+chunk], address)

		if err := hw.finishTransfer(); err != nil {
			return err
		}

		done += chunk
	}

	return nil
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
