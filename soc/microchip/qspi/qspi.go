// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package qspi implements a driver for the Microchip Quad SPI (QSPI)
// controller, in serial memory mode, adopting the following specifications:
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

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/gck"
	"github.com/usbarmory/tamago/soc/microchip/xdmac"
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
	ISR_LWRA    = 11
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

	QSPI_WRACNT = 0x54

	QSPI_WPMR  = 0xe4
	WPMR_WPKEY = 8
	WPMR_KEY   = 0x515350
)

// Timeout is the default timeout for QSPI operations.
const Timeout = 200 * time.Millisecond

const (
	// maximum DMA transfer size
	dmaMaxSize = 1 << 20
	// minimum DMA transfer size
	dmaMinSize = 64
)

const (
	// minimum chip select inactive time
	chipSelectInactive = 70 * time.Nanosecond
	// delay from chip select to first clock edge
	clockDelay = 20 * time.Nanosecond
)

// QSPI errors
var (
	ErrInvalidInstance  = errors.New("invalid controller instance")
	ErrNotInitialized   = errors.New("controller is not initialized")
	ErrInvalidCommand   = errors.New("invalid command")
	ErrInvalidOperation = errors.New("invalid operation")
	ErrSessionClosed    = errors.New("session used outside its callback")
	ErrCommandAddress   = errors.New("command address out of range")
	ErrRange            = errors.New("access outside memory-mapped aperture")
)

// Width represents the number of lines used by the instruction, address and
// data phases of a command.
type Width uint32

// Command widths
const (
	Single      Width = iota // 1-1-1
	DualOutput               // 1-1-2
	QuadOutput               // 1-1-4
	DualIO                   // 1-2-2
	QuadIO                   // 1-4-4
	DualCommand              // 2-2-2
	QuadCommand              // 4-4-4
)

// Command represents a serial memory command.
type Command struct {
	// Instruction code
	Instruction uint8
	// Address size (0, 8, 16, 24 or 32 bits)
	AddressBits int
	// Clock cycles between address and data
	DummyCycles uint8
	// Command width
	Width Width
}

// QSPI represents a QSPI controller instance.
type QSPI struct {
	sync.Mutex

	// Base register
	Base uint32
	// Memory-mapped serial memory
	MMAP uint32
	// Memory-mapped serial memory size (optional)
	MMAPSize uint32
	// Generic clock configuration register
	GCK uint32
	// Generic clock source frequency (Hz)
	ParentClock uint32
	// Generic clock frequency (Hz)
	TargetClock uint32
	// Register polling timeout (default: Timeout)
	Timeout time.Duration

	// DMA controller for memory-mapped reads (optional)
	DMA *xdmac.XDMAC
	// DMA controller channel
	DMAChannel int
	// DMA region for memory-mapped reads (default: dma.Default())
	Region *dma.Region

	ready bool
}

// Ready returns whether the controller is initialized.
func (hw *QSPI) Ready() bool {
	return hw.ready
}

// Init initializes the QSPI controller in serial memory mode.
func (hw *QSPI) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.ready {
		return nil
	}

	if hw.Base == 0 || hw.GCK == 0 {
		return ErrInvalidInstance
	}

	if hw.DMA != nil && hw.Region == nil {
		hw.Region = dma.Default()
	}

	if hw.DMA != nil && (hw.Region == nil || hw.Region.End() > 1<<32) {
		return ErrInvalidInstance
	}

	prescaler, err := gck.Prescaler(hw.ParentClock, hw.TargetClock)
	if err != nil {
		return err
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	// end any transfer left active by an earlier boot stage
	if err := hw.shutdown(); err != nil {
		return err
	}

	// disable write protection
	var protection uint32
	bits.SetN(&protection, WPMR_WPKEY, 0xffffff, WPMR_KEY)
	reg.Write(hw.Base+QSPI_WPMR, protection)

	// stop the DLL before changing its clock
	hw.command(CR_DLLOFF)

	if err := hw.waitClear(QSPI_SR, SR_DLOCK, "DLL disable"); err != nil {
		return err
	}

	gck.Enable(hw.GCK, prescaler)

	// STPCAL is undocumented but set by the vendor TF-A
	var control uint32
	bits.Set(&control, CR_DLLON)
	bits.Set(&control, CR_STPCAL)
	hw.commandMask(control)

	if err := hw.waitSet(QSPI_SR, SR_DLOCK, "DLL lock"); err != nil {
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

	gckHz := hw.ParentClock / (prescaler + 1)

	var mode uint32
	bits.Set(&mode, MR_SMM)
	bits.SetN(&mode, MR_DLYCS, 0xff, gckCycles(chipSelectInactive, gckHz, 0xff))
	reg.Write(hw.Base+QSPI_MR, mode)

	var serialClock uint32
	bits.SetN(&serialClock, SCR_DLYBS, 0xff, gckCycles(clockDelay, gckHz, 0xff))
	reg.Write(hw.Base+QSPI_SCR, serialClock)

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

// Shutdown ends any active transfer and resets the controller.
func (hw *QSPI) Shutdown() (err error) {
	hw.Lock()
	defer hw.Unlock()

	return hw.shutdown()
}

func (hw *QSPI) shutdown() (err error) {
	hw.ready = false

	// a disabled or unclocked controller has no transfer to end
	if !reg.Get(hw.Base+QSPI_SR, SR_QSPIENS) || !reg.Get(hw.Base+QSPI_SR, SR_DLOCK) {
		return nil
	}

	firstErr := hw.finishTransfer()

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

func gckCycles(delay time.Duration, gckHz uint32, mask uint32) uint32 {
	periods := (uint64(delay)*uint64(gckHz) + uint64(time.Second) - 1) / uint64(time.Second)
	return uint32(min(periods, uint64(mask)))
}

func (hw *QSPI) commandMask(mask uint32) {
	reg.Write(hw.Base+QSPI_CR, mask)
}

func (hw *QSPI) waitSet(offset uint32, bit int, operation string) (err error) {
	if !reg.WaitFor(hw.Timeout, hw.Base+offset, bit, 1, 1) {
		return fmt.Errorf("%s timeout", operation)
	}

	return nil
}

func (hw *QSPI) waitClear(offset uint32, bit int, operation string) (err error) {
	if !reg.WaitFor(hw.Timeout, hw.Base+offset, bit, 1, 0) {
		return fmt.Errorf("%s timeout", operation)
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

	// dummy read required after an instruction frame change
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
