// Microchip Secure Digital Host Controller Interface support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sdhci implements ADMA2 eMMC support for the Microchip Secure Digital
// Host Controller Interface found on LAN969x SoCs.
//
// The following specifications are adopted:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//   - SD Host Controller Simplified Specification - Version 3.00
//   - JESD84-B51 - Embedded Multi-Media Card (eMMC) Electrical Standard (5.1) - 2015/02
//
// The driver supports sector-addressed eMMC devices in 8-bit legacy mode. It
// initializes the controller and card, reports card metadata, and transfers
// full 512-byte blocks. DMA allocations use dma.Default() unless callers provide
// a controller-specific region. The region must be controller-accessible,
// non-cacheable, and below 4 GiB. Multi-block reads are stopped explicitly with
// CMD12; failed stop recovery invalidates the instance until initialization.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go, see https://github.com/usbarmory/tamago.
package sdhci

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// SDMMC registers
const (
	SDMMC_BSR = 0x04
	SDMMC_BCR = 0x06

	SDMMC_ARG1R = 0x08

	SDMMC_TMR  = 0x0c
	TMR_DMAEN  = 0
	TMR_BCEN   = 1
	TMR_DTDSEL = 4
	TMR_MSBSEL = 5

	SDMMC_CR = 0x0e

	CR_RESPTYP         = 0
	CR_RESPTYP_MASK    = 0x3
	CR_RESPTYP_NORESP  = 0x0
	CR_RESPTYP_RL136   = 0x1
	CR_RESPTYP_RL48    = 0x2
	CR_RESPTYP_RL48BSY = 0x3
	CR_CMDCCEN         = 3
	CR_CMDICEN         = 4
	CR_DPSEL           = 5
	CR_CMDTYP          = 6
	CR_CMDTYP_MASK     = 0x3
	CR_CMDTYP_ABORT    = 0x3
	CR_CMDIDX          = 8
	CR_CMDIDX_MASK     = 0x3f

	SDMMC_RR0 = 0x10

	SDMMC_PSR   = 0x24
	PSR_CMDINHC = 0
	PSR_CMDINHD = 1

	SDMMC_HC1R       = 0x28
	HC1R_DW_4BIT     = 1
	HC1R_DMASEL      = 3
	HC1R_DMASEL_MASK = 0x3
	HC1R_EXTDW       = 5

	DMASEL_ADMA2_32 = 0b10

	SDMMC_PCR        = 0x29
	PCR_SDBPWR       = 0
	PCR_SDBVSEL      = 1
	PCR_SDBVSEL_MASK = 0x7
	PCR_SDBVSEL_3V3  = 0x7

	SDMMC_CCR                = 0x2c
	CCR_INTCLKEN             = 0
	CCR_INTCLKS              = 1
	CCR_SDCLKEN              = 2
	CCR_CLKGSEL              = 5
	CCR_SDCLKFSEL_UPPER      = 6
	CCR_SDCLKFSEL_UPPER_MASK = 0x3
	CCR_SDCLKFSEL_LOWER      = 8
	CCR_SDCLKFSEL_LOWER_MASK = 0xff
	CCR_SDCLKFSEL_MASK       = 0x3ff

	SDMMC_TCR = 0x2e

	SDMMC_SRR    = 0x2f
	SRR_SWRSTALL = 0
	SRR_SWRSTCMD = 1
	SRR_SWRSTDAT = 2

	SDMMC_NISTR  = 0x30
	NISTR_CMDC   = 0
	NISTR_TRFC   = 1
	NISTR_ERRINT = 15

	SDMMC_EISTR                      = 0x32
	EISTR_ADMA                       = 9
	EISTR_DAT_LINE_ERROR_MASK uint16 = 0x0070

	SDMMC_NISTER = 0x34
	SDMMC_EISTER = 0x36

	SDMMC_CAPR  = 0x40
	CAPR_ADMA2  = 19
	SDMMC_AESR  = 0x54
	SDMMC_ASAR0 = 0x58
	SDMMC_ASAR1 = 0x5c

	SDMMC_MC1R       = 0x204
	MC1R_CMDTYP      = 0
	MC1R_CMDTYP_MASK = 0x3
	MC1R_OPD         = 4
	MC1R_FCD         = 7
)

// Generic Clock Configuration register fields
const (
	GCK_ENA            = 0
	GCK_SRC_SEL        = 8
	GCK_SRC_SEL_MASK   = 0x3
	GCK_PRESCALER      = 16
	GCK_PRESCALER_MASK = 0xff
)

// SDHCI constants
const (
	// data transfer direction
	WRITE = 0
	READ  = 1

	// Maximum SDHCI data timeout exponent; 0xf is reserved.
	dataTimeoutCounter = 0x0e
	allInterrupts      = 0xffff

	mmcIdentificationClockHz = 400_000
	mmcLegacyClockHz         = 25_000_000

	r1ErrorMask uint32 = 0xfff9a080

	// On arm64 Go `copy` built-in cannot be safely used on device memory
	// as LDP/STP instructions require 8-byte alignment, for this reason
	// all `copy` against DMA buffers are forced to 8-byte aligned slices.
	copyAlign    = 8
	dmaAlignment = max(admaAlignment, copyAlign)

	// BlockSize is the size of a sector-addressed eMMC block.
	BlockSize            = MMC_DEFAULT_BLOCK_SIZE
	maxBlocksPerTransfer = 0xffff
)

var (
	// ControllerSetupTimeout controls controller reset and setup waits.
	ControllerSetupTimeout = 500 * time.Millisecond
	// ClockSetupTimeout controls clock idle and stabilization waits.
	ClockSetupTimeout = 100 * time.Millisecond
	// WriteTimeout controls card write and cache synchronization waits.
	WriteTimeout = 30 * time.Second

	// ErrNotInitialized indicates that Detect has not completed successfully.
	ErrNotInitialized = errors.New("sdhci: eMMC card is not initialized")
	// ErrRange indicates that a transfer exceeds the detected card or LBA range.
	ErrRange = errors.New("sdhci: LBA range overflows")
	// ErrAlignment indicates that a transfer is empty or not block-aligned.
	ErrAlignment = errors.New("sdhci: buffer size must be a non-zero multiple of 512")
)

// CardInfo describes the detected eMMC card.
type CardInfo struct {
	// Relative Card Address
	RCA uint16
	// Operation Conditions register
	OCR uint32
	// Card Identification register
	CID [4]uint32
	// Extended CSD revision
	ExtCSDRev byte
	// Device type
	DeviceType byte
	// Write cache state
	CacheEnabled bool

	// Block size
	BlockSize int
	// Capacity
	Blocks int
}

// SDHCI represents a Microchip SDHCI controller bound to an eMMC device.
type SDHCI struct {
	sync.Mutex

	// Base register
	Base uint32
	// Generic Clock Configuration register
	GCK uint32
	// Generic Clock source frequency in Hz
	ParentClock uint32
	// Requested Generic Clock frequency in Hz
	TargetClock uint32
	// Controller pin configuration, skipped when nil
	ConfigurePins func() error
	// DMA memory used for ADMA2 descriptors and data. It defaults to
	// dma.Default() and must be controller-accessible, non-cacheable, and below
	// 4 GiB.
	DMA *dma.Region

	// control registers
	bsr    uint32
	bcr    uint32
	arg1r  uint32
	tmr    uint32
	cr     uint32
	rr     uint32
	psr    uint32
	hc1r   uint32
	pcr    uint32
	ccr    uint32
	tcr    uint32
	srr    uint32
	nistr  uint32
	eistr  uint32
	nister uint32
	eister uint32
	capr   uint32
	aesr   uint32
	asar0  uint32
	asar1  uint32
	mc1r   uint32

	// controller state
	controllerReady bool
	ready           bool

	// detected card properties
	card CardInfo

	maxBlocks int
}

func (hw *SDHCI) dumpRegisters() string {
	return fmt.Sprintf(
		"GCK=0x%08x PSR=0x%08x CCR=0x%04x SRR=0x%02x NISTR=0x%04x EISTR=0x%04x AESR=0x%02x TMR=0x%04x BCR=0x%04x",
		reg.Read(hw.GCK),
		reg.Read(hw.psr),
		reg.Read16(hw.ccr),
		reg.Read8(hw.srr),
		reg.Read16(hw.nistr),
		reg.Read16(hw.eistr),
		reg.Read8(hw.aesr),
		reg.Read16(hw.tmr),
		reg.Read16(hw.bcr),
	)
}

func genericClockPrescaler(parent uint32, target uint32) (uint32, error) {
	if parent == 0 || target == 0 {
		return 0, errors.New("sdhci: invalid clock frequency")
	}

	divider := (uint64(parent) + uint64(target) - 1) / uint64(target)

	if divider > GCK_PRESCALER_MASK+1 {
		return 0, errors.New("sdhci: clock prescaler out of range")
	}

	return uint32(divider - 1), nil
}

func gckConfigurationMatches(value uint32, prescaler uint32) bool {
	return bits.Get(&value, GCK_ENA) &&
		bits.GetN(&value, GCK_SRC_SEL, GCK_SRC_SEL_MASK) == 0 &&
		bits.GetN(&value, GCK_PRESCALER, GCK_PRESCALER_MASK) == prescaler
}

func (hw *SDHCI) enableGenericClock(prescaler uint32) error {
	value := reg.Read(hw.GCK)

	if bits.Get(&value, GCK_ENA) {
		// A previous firmware stage may leave SDCLK running. Stop it only
		// after the command and data paths become idle, while the inherited
		// functional clock is still available.
		var inhibitMask uint32
		bits.Set(&inhibitMask, PSR_CMDINHC)
		bits.Set(&inhibitMask, PSR_CMDINHD)

		if !reg.WaitFor(ControllerSetupTimeout, hw.psr, 0, int(inhibitMask), 0) {
			return fmt.Errorf("sdhci: inherited controller busy (%s)", hw.dumpRegisters())
		}

		// stop card clock
		reg.Clear16(hw.ccr, CCR_SDCLKEN)

		if gckConfigurationMatches(value, prescaler) {
			return nil
		}
	}

	// select source 0 and configure the functional clock
	reg.Clear(hw.GCK, GCK_ENA)
	reg.SetN(hw.GCK, GCK_SRC_SEL, GCK_SRC_SEL_MASK, 0)
	reg.SetN(hw.GCK, GCK_PRESCALER, GCK_PRESCALER_MASK, prescaler)
	reg.Set(hw.GCK, GCK_ENA)

	return nil
}

func (hw *SDHCI) setClockFrequency(frequencyHz uint32) error {
	if frequencyHz == 0 {
		return errors.New("sdhci: invalid SD clock frequency")
	}

	prescaler, err := genericClockPrescaler(hw.ParentClock, hw.TargetClock)
	if err != nil {
		return err
	}

	sourceHz := hw.ParentClock / (prescaler + 1)
	divisor := (uint64(sourceHz) + uint64(frequencyHz) - 1) / uint64(frequencyHz)

	if divisor > CCR_SDCLKFSEL_MASK+1 {
		return errors.New("sdhci: SD clock divider out of range")
	}

	divider := uint16(divisor - 1)

	var inhibitMask uint32
	bits.Set(&inhibitMask, PSR_CMDINHC)
	bits.Set(&inhibitMask, PSR_CMDINHD)

	if !reg.WaitFor(ClockSetupTimeout, hw.psr, 0, int(inhibitMask), 0) {
		return fmt.Errorf("sdhci: clock change timeout (%s)", hw.dumpRegisters())
	}

	// stop card clock
	reg.Clear16(hw.ccr, CCR_SDCLKEN)

	// configure and start the internal programmable clock
	var clock uint16
	bits.Set16(&clock, CCR_INTCLKEN)
	bits.Set16(&clock, CCR_CLKGSEL)
	bits.SetN16(&clock, CCR_SDCLKFSEL_UPPER, CCR_SDCLKFSEL_UPPER_MASK, divider>>8)
	bits.SetN16(&clock, CCR_SDCLKFSEL_LOWER, CCR_SDCLKFSEL_LOWER_MASK, divider&CCR_SDCLKFSEL_LOWER_MASK)
	reg.Write16(hw.ccr, clock)

	if !reg.WaitFor16(ClockSetupTimeout, hw.ccr, CCR_INTCLKS, 1, 1) {
		return fmt.Errorf("sdhci: internal clock did not stabilize (%s)", hw.dumpRegisters())
	}

	// start card clock
	reg.Set16(hw.ccr, CCR_SDCLKEN)

	return nil
}

func (hw *SDHCI) reset(mask uint8, timeout time.Duration) error {
	reg.Write8(hw.srr, mask)

	if !reg.WaitFor8(timeout, hw.srr, 0, int(mask), 0) {
		return fmt.Errorf("sdhci: controller reset 0x%02x timeout (%s)", mask, hw.dumpRegisters())
	}

	return nil
}

func (hw *SDHCI) initController(prescaler uint32) error {
	if hw.ConfigurePins != nil {
		if err := hw.ConfigurePins(); err != nil {
			return fmt.Errorf("sdhci: pin configuration failed: %w", err)
		}
	}

	if err := hw.enableGenericClock(prescaler); err != nil {
		return err
	}

	// reset all host circuits
	if err := hw.reset(1<<SRR_SWRSTALL, ControllerSetupTimeout); err != nil {
		return err
	}

	// use the maximum data timeout
	reg.Write8(hw.tcr, dataTimeoutCounter)

	// enable 3.3 V bus power
	var power uint16
	bits.SetN16(&power, PCR_SDBVSEL, PCR_SDBVSEL_MASK, PCR_SDBVSEL_3V3)
	bits.Set16(&power, PCR_SDBPWR)
	reg.Write8(hw.pcr, uint8(power))

	// force card insertion
	cardDetect := uint16(reg.Read8(hw.mc1r))
	bits.Set16(&cardDetect, MC1R_FCD)
	reg.Write8(hw.mc1r, uint8(cardDetect))

	// enable all status events
	reg.Write16(hw.nister, allInterrupts)
	reg.Write16(hw.eister, allInterrupts)

	return hw.setClockFrequency(mmcIdentificationClockHz)
}

func (hw *SDHCI) dmaBlockLimit() int {
	available := int(hw.DMA.Size()) - 2*(dmaAlignment-1)
	blocks := min(available/BlockSize, maxBlocksPerTransfer)

	for blocks > 0 {
		size := blocks * BlockSize

		if size+admaTableSize(size) <= available {
			return blocks
		}

		blocks--
	}

	return 0
}

// Init initializes the controller. Detect must be called afterward to
// initialize the eMMC card.
func (hw *SDHCI) Init() error {
	hw.Lock()
	defer hw.Unlock()

	hw.controllerReady = false
	hw.ready = false
	hw.card = CardInfo{}
	hw.maxBlocks = 0

	if hw.Base == 0 {
		return errors.New("sdhci: invalid controller base")
	}

	if hw.GCK == 0 {
		return errors.New("sdhci: invalid Generic Clock register")
	}

	if hw.DMA == nil {
		hw.DMA = dma.Default()
	}

	if hw.DMA == nil {
		return errors.New("sdhci: DMA memory is not configured")
	}

	if hw.DMA.End() > 1<<32 {
		return errors.New("sdhci: DMA memory exceeds ADMA2 address range")
	}

	hw.maxBlocks = hw.dmaBlockLimit()
	if hw.maxBlocks == 0 {
		return errors.New("sdhci: DMA memory is too small")
	}

	prescaler, err := genericClockPrescaler(hw.ParentClock, hw.TargetClock)
	if err != nil {
		return err
	}

	hw.bsr = hw.Base + SDMMC_BSR
	hw.bcr = hw.Base + SDMMC_BCR
	hw.arg1r = hw.Base + SDMMC_ARG1R
	hw.tmr = hw.Base + SDMMC_TMR
	hw.cr = hw.Base + SDMMC_CR
	hw.rr = hw.Base + SDMMC_RR0
	hw.psr = hw.Base + SDMMC_PSR
	hw.hc1r = hw.Base + SDMMC_HC1R
	hw.pcr = hw.Base + SDMMC_PCR
	hw.ccr = hw.Base + SDMMC_CCR
	hw.tcr = hw.Base + SDMMC_TCR
	hw.srr = hw.Base + SDMMC_SRR
	hw.nistr = hw.Base + SDMMC_NISTR
	hw.eistr = hw.Base + SDMMC_EISTR
	hw.nister = hw.Base + SDMMC_NISTER
	hw.eister = hw.Base + SDMMC_EISTER
	hw.capr = hw.Base + SDMMC_CAPR
	hw.aesr = hw.Base + SDMMC_AESR
	hw.asar0 = hw.Base + SDMMC_ASAR0
	hw.asar1 = hw.Base + SDMMC_ASAR1
	hw.mc1r = hw.Base + SDMMC_MC1R

	if err := hw.initController(prescaler); err != nil {
		return err
	}

	// LAN969x implements the SDHCI ADMA2 interface. Require the advertised
	// capability, then select its 32-bit descriptor format because DMA memory
	// is constrained below 4 GiB.
	if !reg.Get(hw.capr, CAPR_ADMA2) {
		return errors.New("sdhci: controller does not support ADMA2")
	}

	// select 32-bit ADMA2
	hostControl := uint16(reg.Read8(hw.hc1r))
	bits.SetN16(&hostControl, HC1R_DMASEL, HC1R_DMASEL_MASK, DMASEL_ADMA2_32)
	reg.Write8(hw.hc1r, uint8(hostControl))

	hw.controllerReady = true

	return nil
}

func (hw *SDHCI) validateTransfer(lba int, length int) error {
	if !hw.ready {
		return ErrNotInitialized
	}

	if length == 0 || length%BlockSize != 0 {
		return ErrAlignment
	}

	blocks := length / BlockSize

	if lba < 0 || lba > hw.card.Blocks-blocks {
		return ErrRange
	}

	return nil
}

func (hw *SDHCI) transferBlocks(index uint16, dtd uint32, lba int, buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if err = hw.validateTransfer(lba, len(buf)); err != nil {
		return
	}

	switch dtd {
	case WRITE:
		for len(buf) > 0 {
			if err = hw.writeBlock(index, uint32(lba), buf[:BlockSize]); err != nil {
				return
			}

			if err = hw.waitState(CURRENT_STATE_TRAN, WriteTimeout); err != nil {
				err = hw.invalidateTransfer(err)
				return
			}

			buf = buf[BlockSize:]
			lba++
		}
	case READ:
		for len(buf) > 0 {
			blocks := min(len(buf)/BlockSize, hw.maxBlocks)

			length := blocks * BlockSize

			if err = hw.readBlocks(index, uint32(lba), buf[:length], uint16(blocks)); err != nil {
				return
			}

			buf = buf[length:]
			lba += blocks
		}
	default:
		err = errors.New("sdhci: invalid transfer direction")
	}

	return
}

// WriteBlocks transfers full blocks of data to the card. Each block uses
// CMD24 so a failure cannot make the completion of later blocks ambiguous.
func (hw *SDHCI) WriteBlocks(lba int, buf []byte) (err error) {
	// CMD24 - WRITE_BLOCK - write one block at a time
	return hw.transferBlocks(24, WRITE, lba, buf)
}

// ReadBlocks transfers full blocks of data from the card.
func (hw *SDHCI) ReadBlocks(lba int, buf []byte) (err error) {
	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks (CMD17 for one)
	return hw.transferBlocks(18, READ, lba, buf)
}

// Read transfers data from the card.
func (hw *SDHCI) Read(offset int64, size int64) (buf []byte, err error) {
	if offset < 0 || size <= 0 {
		return nil, nil
	}

	startLBA := offset / BlockSize
	blockOffset := offset % BlockSize
	blocks := (blockOffset + size + BlockSize - 1) / BlockSize

	buf = make([]byte, int(blocks)*BlockSize)

	if err = hw.ReadBlocks(int(startLBA), buf); err != nil {
		return
	}

	start := int(blockOffset)
	buf = buf[start : start+int(size)]

	return
}

func (hw *SDHCI) readBlocks(index uint16, lba uint32, buf []byte, blocks uint16) error {
	return hw.transferDMA(index, READ, lba, buf, blocks)
}

func (hw *SDHCI) writeBlock(index uint16, lba uint32, buf []byte) error {
	return hw.transferDMA(index, WRITE, lba, buf, 1)
}

func copyAlignedDMABuffer(dst []byte, src []byte) {
	n := len(src)
	r := n % copyAlign
	n -= r

	copy(dst[:n], src[:n])

	for i := range r {
		dst[n+i] = src[n+i]
	}
}

func writeDMABuffer(dst []byte, src []byte) {
	staging := make([]byte, len(src))
	copy(staging, src)
	copyAlignedDMABuffer(dst, staging)
}

func readDMABuffer(dst []byte, src []byte) {
	staging := make([]byte, len(src))
	copyAlignedDMABuffer(staging, src)
	copy(dst, staging)
}

func (hw *SDHCI) transferDMA(index uint16, direction uint32, lba uint32, buf []byte, blocks uint16) (err error) {
	dmaAddress, dmaBuffer := hw.DMA.Reserve(len(buf), dmaAlignment)
	defer hw.DMA.Release(dmaAddress)

	if direction == WRITE {
		writeDMABuffer(dmaBuffer, buf)
	}

	descriptors := admaTable(dmaAddress, len(buf))
	descriptorAddress, descriptorBuffer := hw.DMA.Reserve(len(descriptors), dmaAlignment)
	defer hw.DMA.Release(descriptorAddress)
	writeDMABuffer(descriptorBuffer, descriptors)

	// program the ADMA table
	reg.Write(hw.asar0, uint32(descriptorAddress))
	reg.Write(hw.asar1, 0)

	// program block geometry
	reg.Write16(hw.bsr, BlockSize)
	reg.Write16(hw.bcr, blocks)

	multi := blocks > 1

	// enable ADMA and block-count termination
	var transferMode uint16
	bits.Set16(&transferMode, TMR_DMAEN)
	bits.Set16(&transferMode, TMR_BCEN)
	bits.SetTo16(&transferMode, TMR_DTDSEL, direction == READ)
	bits.SetTo16(&transferMode, TMR_MSBSEL, multi)
	reg.Write16(hw.tmr, transferMode)

	command := index
	if index == 18 && !multi {
		// CMD17 - READ_SINGLE_BLOCK - read one block
		command = 17
	}

	status, _, issued, commandErr := hw.runCommand(command, lba, 0)
	if command == 18 && issued {
		defer func() {
			err = hw.stopTransmission(err)
		}()
	}

	if commandErr != nil {
		err = fmt.Errorf("CMD%d transfer failed: %w", command, commandErr)

		if command != 18 || !issued {
			err = hw.invalidateTransfer(err)
		}

		return
	}

	if responseErr := checkR1(status); responseErr != nil {
		err = fmt.Errorf("CMD%d transfer: %w", command, responseErr)

		if command != 18 {
			err = hw.invalidateTransfer(err)
		}

		return
	}

	transferTimeout := CommandTimeout
	if direction == WRITE {
		transferTimeout = WriteTimeout
	}

	if _, statusErr := hw.pollStatus(1<<NISTR_TRFC, transferTimeout); statusErr != nil {
		err = fmt.Errorf("CMD%d transfer: %w", command, statusErr)

		if command != 18 {
			err = hw.invalidateTransfer(err)
		}

		return
	}

	if direction == READ {
		readDMABuffer(buf, dmaBuffer)
	}

	return
}
