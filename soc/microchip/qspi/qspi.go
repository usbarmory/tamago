// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package qspi implements serial-memory support for Microchip QSPI controllers.
//
// The following specification is adopted:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package qspi

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// QSPI registers
const (
	QSPI_CR     = 0x00
	CR_LASTXFER = 24
	CR_STTFR    = 9
	CR_UPDCFG   = 8
	CR_SWRST    = 7
	CR_STPCAL   = 4
	CR_DLLOFF   = 3
	CR_DLLON    = 2
	CR_QSPIDIS  = 1
	CR_QSPIEN   = 0

	QSPI_MR  = 0x04
	MR_DLYCS = 24
	MR_SMM   = 0

	QSPI_RDR = 0x08
	QSPI_TDR = 0x0c

	QSPI_ISR    = 0x10
	ISR_CSRA    = 15
	ISR_TXEMPTY = 2
	ISR_TDRE    = 1
	ISR_RDRF    = 0

	QSPI_SCR  = 0x20
	SCR_DLYBS = 16

	QSPI_SR    = 0x24
	SR_DLOCK   = 5
	SR_HIDLE   = 4
	SR_RBUSY   = 3
	SR_QSPIENS = 1
	SR_SYNCBSY = 0

	QSPI_IAR  = 0x30
	QSPI_WICR = 0x34

	QSPI_IFR      = 0x38
	IFR_APBTFRTYP = 24
	IFR_SMRM      = 23
	IFR_NBDUM     = 16
	IFR_TFRTYP    = 12
	IFR_ADDRL     = 10
	IFR_DATAEN    = 7
	IFR_ADDREN    = 5
	IFR_INSTEN    = 4
	IFR_WIDTH     = 0

	QSPI_RICR = 0x3c

	QSPI_WPMR  = 0xe4
	WPMR_WPKEY = 8
	WPMR_KEY   = 0x515350
)

// Generic Clock Configuration register fields
const (
	GCK_ENA            = 0
	GCK_SRC_SEL        = 8
	GCK_SRC_SEL_MASK   = 0x3
	GCK_PRESCALER      = 16
	GCK_PRESCALER_MASK = 0xff
)

// Timeout is the default timeout for QSPI operations.
const Timeout = 200 * time.Millisecond

var (
	// ErrInvalidInstance is returned when required controller fields are unset.
	ErrInvalidInstance = errors.New("qspi: invalid controller instance")
	// ErrNotInitialized is returned when a controller operation is attempted
	// before successful initialization.
	ErrNotInitialized = errors.New("qspi: controller is not initialized")
	// ErrInvalidCommand is returned when a command has unsupported fields.
	ErrInvalidCommand = errors.New("qspi: invalid command")
	// ErrInvalidOperation is returned when a nil controller operation is passed.
	ErrInvalidOperation = errors.New("qspi: invalid operation")
	// ErrCommandAddress is returned when an address cannot be represented by a
	// command's address phase.
	ErrCommandAddress = errors.New("qspi: command address out of range")
	// ErrRange is returned when a memory-mapped read exceeds the configured
	// aperture or the 32-bit address space.
	ErrRange = errors.New("qspi: read outside memory-mapped aperture")
)

// Width configures the bus width used for command phases.
type Width uint32

const (
	// Single uses one data line for the instruction, address, and data.
	Single Width = iota
	// DualOutput uses one line for the instruction and address, and two for data.
	DualOutput
	// QuadOutput uses one line for the instruction and address, and four for data.
	QuadOutput
	// DualIO uses one line for the instruction, and two for the address and data.
	DualIO
	// QuadIO uses one line for the instruction, and four for the address and data.
	QuadIO
	// DualCommand uses two lines for the instruction, address, and data.
	DualCommand
	// QuadCommand uses four lines for the instruction, address, and data.
	QuadCommand
)

// Command describes a serial-memory command.
type Command struct {
	// Instruction is the command opcode.
	Instruction uint8
	// AddressBits is one of 0, 8, 16, 24, and 32. A zero value omits the
	// address phase.
	AddressBits int
	// DummyCycles is the number of clock cycles inserted before data transfer.
	DummyCycles uint8
	// Width configures the bus width used by each command phase.
	Width Width
}

// QSPI represents a Microchip QSPI controller instance.
type QSPI struct {
	sync.Mutex

	// Base is the controller register base address.
	Base uint32
	// MMAP is the base address of the memory-mapped serial-memory aperture.
	MMAP uint32
	// MMAPSize optionally bounds memory-mapped reads in bytes. A zero value
	// leaves the serial-memory geometry to the caller.
	MMAPSize uint32
	// GCK is the Generic Clock Configuration register address.
	GCK uint32
	// ParentClock is the Generic Clock source frequency in Hz.
	ParentClock uint32
	// TargetClock is the requested Generic Clock frequency in Hz.
	TargetClock uint32
	// Timeout bounds each controller poll. One transfer may use multiple polls.
	Timeout time.Duration

	ready        bool
	needsCleanup bool
}

// Ready reports whether the controller is initialized and has not been
// invalidated by a transfer failure.
func (hw *QSPI) Ready() bool {
	hw.Lock()
	defer hw.Unlock()

	return hw.ready
}

// Init initializes a QSPI controller in serial-memory mode. It does not issue
// commands to the attached device. Repeated calls preserve a ready controller.
// A failed attempt leaves the controller unready and retains any required
// cleanup for the next Init or Shutdown call.
func (hw *QSPI) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.ready {
		return nil
	}

	if hw.Base == 0 || hw.GCK == 0 {
		return ErrInvalidInstance
	}

	prescaler, err := genericClockPrescaler(hw.ParentClock, hw.TargetClock)
	if err != nil {
		return err
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	if hw.needsCleanup {
		if err := hw.shutdown(); err != nil {
			return err
		}
	}

	// Retain cleanup responsibility before the first controller write.
	hw.needsCleanup = true

	// Clear inherited control-register protection before issuing commands.
	reg.Write(hw.Base+QSPI_WPMR, WPMR_KEY<<WPMR_WPKEY)

	// Stop the DLL before changing its functional clock.
	hw.command(CR_DLLOFF)
	if err := hw.waitClear(QSPI_SR, SR_DLOCK, "DLL disable"); err != nil {
		return err
	}

	hw.enableGenericClock(prescaler)

	hw.commandMask(1<<CR_DLLON | 1<<CR_STPCAL)
	if err := hw.waitSet(QSPI_SR, SR_DLOCK, "DLL calibration"); err != nil {
		return err
	}

	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "controller synchronization"); err != nil {
		return err
	}

	hw.command(CR_QSPIDIS)
	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "controller disable"); err != nil {
		return err
	}

	hw.command(CR_SWRST)
	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "controller reset"); err != nil {
		return err
	}

	// Re-arm DLL calibration after reset.
	hw.commandMask(1<<CR_DLLON | 1<<CR_STPCAL)
	if err := hw.waitSet(QSPI_SR, SR_DLOCK, "DLL recalibration"); err != nil {
		return err
	}

	// Configure serial-memory mode and conservative chip-select delays.
	reg.Write(hw.Base+QSPI_MR, 7<<MR_DLYCS|1<<MR_SMM)
	reg.Write(hw.Base+QSPI_SCR, 2<<SCR_DLYBS)

	if err := hw.updateConfiguration(); err != nil {
		return err
	}

	hw.command(CR_QSPIEN)

	if err := hw.waitSet(QSPI_SR, SR_QSPIENS, "controller enable"); err != nil {
		return err
	}

	hw.ready = true

	return nil
}

// Shutdown ends any active transfer and resets the controller, including after
// a failed initialization or transfer. It does not issue a separate command to
// the attached memory.
func (hw *QSPI) Shutdown() (err error) {
	hw.Lock()
	defer hw.Unlock()

	return hw.shutdown()
}

func (hw *QSPI) shutdown() (err error) {
	hw.ready = false

	if !hw.needsCleanup {
		return nil
	}

	firstErr := hw.finishTransfer()

	// Never issue synchronized commands while SYNCBSY remains asserted.
	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "controller synchronization"); err != nil {
		return errors.Join(firstErr, err)
	}

	hw.command(CR_QSPIDIS)
	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "controller disable"); err != nil {
		return errors.Join(firstErr, err)
	}

	hw.command(CR_SWRST)
	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "controller reset"); err != nil {
		return errors.Join(firstErr, err)
	}

	hw.needsCleanup = false

	return firstErr
}

func (hw *QSPI) invalidateTransfer(transferErr error) (err error) {
	if resetErr := hw.shutdown(); resetErr != nil {
		return fmt.Errorf("%w (controller recovery failed: %v)", transferErr, resetErr)
	}

	return transferErr
}

func (hw *QSPI) command(bit int) {
	hw.commandMask(1 << bit)
}

func genericClockPrescaler(parentHz uint32, targetHz uint32) (prescaler uint32, err error) {
	if parentHz == 0 {
		return 0, errors.New("qspi: invalid parent clock frequency")
	}

	if targetHz == 0 {
		return 0, errors.New("qspi: invalid target clock frequency")
	}

	divider := (uint64(parentHz) + uint64(targetHz) - 1) / uint64(targetHz)
	if divider > GCK_PRESCALER_MASK+1 {
		return 0, errors.New("qspi: clock prescaler out of range")
	}

	prescaler = uint32(divider - 1)

	return
}

func (hw *QSPI) enableGenericClock(prescaler uint32) {
	reg.Clear(hw.GCK, GCK_ENA)
	reg.SetN(hw.GCK, GCK_SRC_SEL, GCK_SRC_SEL_MASK, 0)
	reg.SetN(hw.GCK, GCK_PRESCALER, GCK_PRESCALER_MASK, prescaler)
	reg.Set(hw.GCK, GCK_ENA)
}

func (hw *QSPI) commandMask(mask uint32) {
	reg.Write(hw.Base+QSPI_CR, mask)
}

func (hw *QSPI) waitSet(offset uint32, bit int, operation string) (err error) {
	if !reg.WaitFor(hw.Timeout, hw.Base+offset, bit, 1, 1) {
		return fmt.Errorf("qspi: %s timeout", operation)
	}

	return nil
}

func (hw *QSPI) waitClear(offset uint32, bit int, operation string) (err error) {
	if !reg.WaitFor(hw.Timeout, hw.Base+offset, bit, 1, 0) {
		return fmt.Errorf("qspi: %s timeout", operation)
	}

	return nil
}

func (hw *QSPI) updateConfiguration() (err error) {
	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "configuration synchronization"); err != nil {
		return err
	}

	hw.command(CR_UPDCFG)

	return hw.waitClear(QSPI_SR, SR_SYNCBSY, "configuration update")
}

func (hw *QSPI) changeInstructionFrame(frame uint32) (err error) {
	reg.Write(hw.Base+QSPI_IFR, frame)

	// Dummy status read required after changing the instruction frame.
	reg.Read(hw.Base + QSPI_SR)

	if err := hw.waitClear(QSPI_SR, SR_RBUSY, "read completion"); err != nil {
		return err
	}

	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "instruction synchronization"); err != nil {
		return err
	}

	hw.command(CR_LASTXFER)
	if err := hw.waitSet(QSPI_SR, SR_HIDLE, "controller idle"); err != nil {
		return err
	}

	return hw.updateConfiguration()
}

func (hw *QSPI) finishTransfer() (err error) {
	if err := hw.waitClear(QSPI_SR, SR_RBUSY, "read completion"); err != nil {
		return err
	}

	if err := hw.waitClear(QSPI_SR, SR_SYNCBSY, "transfer synchronization"); err != nil {
		return err
	}

	hw.command(CR_LASTXFER)

	return hw.waitSet(QSPI_SR, SR_HIDLE, "controller idle")
}
