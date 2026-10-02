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
// The driver supports sector-addressed 8-bit eMMC devices up to HS_DDR.
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
	"github.com/usbarmory/tamago/soc/microchip/gck"
)

// SDMMC registers
const (
	SDMMC_SSAR = 0x00

	SDMMC_BSR = 0x04
	SDMMC_BCR = 0x06

	SDMMC_ARG1R = 0x08

	SDMMC_TMR        = 0x0c
	TMR_DMAEN        = 0
	TMR_BCEN         = 1
	TMR_ACMDEN       = 2
	TMR_ACMDEN_MASK  = 0x3
	TMR_ACMDEN_CMD23 = 0x2
	TMR_DTDSEL       = 4
	TMR_MSBSEL       = 5

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
	HC1R_HSEN        = 2
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
	EISTR_ACMD                       = 8
	EISTR_ADMA                       = 9
	EISTR_DAT_LINE_ERROR_MASK uint16 = 0x0070

	SDMMC_NISTER = 0x34
	SDMMC_EISTER = 0x36
	SDMMC_NISIER = 0x38
	SDMMC_EISIER = 0x3a

	SDMMC_ACESR = 0x3c

	SDMMC_CAPR = 0x40
	CAPR_ADMA2 = 19
	CAPR_HSSUP = 21

	SDMMC_CA1R    = 0x44
	CA1R_DDR50SUP = 2

	SDMMC_AESR  = 0x54
	SDMMC_ASAR0 = 0x58
	SDMMC_ASAR1 = 0x5c

	SDMMC_MC1R       = 0x204
	MC1R_CMDTYP      = 0
	MC1R_CMDTYP_MASK = 0x3
	MC1R_DDR         = 3
	MC1R_OPD         = 4
	MC1R_FCD         = 7
)

// SDHCI constants
const (
	// data transfer direction
	WRITE = 0
	READ  = 1

	// Maximum SDHCI data timeout exponent; 0xf is reserved.
	dataTimeoutCounter = 0x0e
	allInterrupts      = 0xffff

	// Waits spin this long before sleeping on the controller interrupt.
	interruptWaitThreshold = 100 * time.Microsecond

	mmcIdentificationClockHz = 400_000
	mmcLegacyClockHz         = 25_000_000
	mmcHighSpeedClockHz      = 50_000_000

	r1ErrorMask uint32 = 0xfff9a080

	// arm64 SIMD copies require 16-byte aligned Device memory.
	dmaAlignment = max(admaAlignment, 0x10)

	// BlockSize is the size of a sector-addressed eMMC block.
	BlockSize            = MMC_DEFAULT_BLOCK_SIZE
	maxBlocksPerTransfer = 0xffff

	// write command boundary in blocks
	writeAlignment = 0x4000 / BlockSize
)

var (
	// ControllerSetupTimeout controls controller reset and setup waits.
	ControllerSetupTimeout = 500 * time.Millisecond
	// ClockSetupTimeout controls clock idle and stabilization waits.
	ClockSetupTimeout = 100 * time.Millisecond
	// WriteTimeout controls card write and cache synchronization waits.
	WriteTimeout = 30 * time.Second
	// ReadBlockTimeout is added to CommandTimeout for each block of a read
	// transfer.
	ReadBlockTimeout = 1 * time.Millisecond
	// InterruptPollInterval bounds how long a transfer or busy wait sleeps
	// between status checks after [SDHCI.EnableInterrupt], when no
	// controller interrupt wakes it.
	InterruptPollInterval = 1 * time.Millisecond

	// ErrNotInitialized indicates that Detect has not completed successfully.
	ErrNotInitialized = errors.New("eMMC card is not initialized")
	// ErrRange indicates that a transfer exceeds the detected card or LBA range.
	ErrRange = errors.New("LBA range overflows")
	// ErrAlignment indicates that a transfer is empty or not block-aligned.
	ErrAlignment = errors.New("buffer size must be a non-zero multiple of 512")
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
	// Volatile write cache size in bytes, zero without a cache
	CacheSize int
	// High Speed
	HS bool
	// Dual Data Rate
	DDR bool

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
	// Interrupt ID (placeholder for caller use)
	IRQ int
	// Generic Clock Configuration register
	GCK uint32
	// Generic Clock source frequency in Hz
	ParentClock uint32
	// Requested Generic Clock frequency in Hz
	TargetClock uint32
	// Region represents the memory used for ADMA2 descriptors and data. It
	// defaults to dma.Default() and must be controller-accessible,
	// non-cacheable, and below 4 GiB.
	Region *dma.Region

	// control registers
	ssar   uint32
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
	nisier uint32
	eisier uint32
	acesr  uint32
	capr   uint32
	ca1r   uint32
	aesr   uint32
	asar0  uint32
	asar1  uint32
	mc1r   uint32

	// controller state
	controllerReady bool
	ready           bool
	event           chan struct{}

	// detected card properties
	card CardInfo

	maxBlocks int
}

func gckConfigurationMatches(value uint32, prescaler uint32) bool {
	return bits.Get(&value, gck.GCK_ENA) &&
		bits.GetN(&value, gck.GCK_SRC_SEL, gck.GCK_SRC_SEL_MASK) == 0 &&
		bits.GetN(&value, gck.GCK_PRESCALER, gck.GCK_PRESCALER_MASK) == prescaler
}

func (hw *SDHCI) enableGenericClock(prescaler uint32) error {
	value := reg.Read(hw.GCK)

	if bits.Get(&value, gck.GCK_ENA) {
		// A previous firmware stage may leave SDCLK running. Stop it only
		// after the command and data paths become idle, while the inherited
		// functional clock is still available.
		var inhibitMask uint32
		bits.Set(&inhibitMask, PSR_CMDINHC)
		bits.Set(&inhibitMask, PSR_CMDINHD)

		if !reg.WaitFor(ControllerSetupTimeout, hw.psr, 0, int(inhibitMask), 0) {
			return errors.New("inherited controller busy")
		}

		// stop card clock
		reg.Clear16(hw.ccr, CCR_SDCLKEN)

		if gckConfigurationMatches(value, prescaler) {
			return nil
		}
	}

	// select source 0 and configure the functional clock
	gck.Enable(hw.GCK, prescaler)

	return nil
}

func (hw *SDHCI) setClockFrequency(frequencyHz uint32) (err error) {
	if frequencyHz == 0 {
		return errors.New("invalid SD clock frequency")
	}

	prescaler, err := gck.Prescaler(hw.ParentClock, hw.TargetClock)

	if err != nil {
		return
	}

	sourceHz := hw.ParentClock / (prescaler + 1)
	divisor := (uint64(sourceHz) + uint64(frequencyHz) - 1) / uint64(frequencyHz)

	if divisor > CCR_SDCLKFSEL_MASK+1 {
		return errors.New("SD clock divider out of range")
	}

	divider := uint16(divisor - 1)

	var inhibitMask uint32
	bits.Set(&inhibitMask, PSR_CMDINHC)
	bits.Set(&inhibitMask, PSR_CMDINHD)

	if !reg.WaitFor(ClockSetupTimeout, hw.psr, 0, int(inhibitMask), 0) {
		return errors.New("clock change timeout")
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
		return errors.New("internal clock did not stabilize")
	}

	// start card clock
	reg.Set16(hw.ccr, CCR_SDCLKEN)

	return nil
}

func (hw *SDHCI) reset(mask uint8, timeout time.Duration) error {
	reg.Write8(hw.srr, mask)

	if !reg.WaitFor8(timeout, hw.srr, 0, int(mask), 0) {
		return fmt.Errorf("controller reset 0x%02x timeout", mask)
	}

	return nil
}

func (hw *SDHCI) initController(prescaler uint32) (err error) {
	if err = hw.enableGenericClock(prescaler); err != nil {
		return
	}

	// reset all host circuits
	if err = hw.reset(1<<SRR_SWRSTALL, ControllerSetupTimeout); err != nil {
		return
	}

	// use the maximum data timeout
	reg.Write8(hw.tcr, dataTimeoutCounter)

	// enable 3.3 V bus power
	var power uint16
	bits.SetN16(&power, PCR_SDBVSEL, PCR_SDBVSEL_MASK, PCR_SDBVSEL_3V3)
	bits.Set16(&power, PCR_SDBPWR)
	reg.Write8(hw.pcr, uint8(power))

	// force card insertion with single data rate sampling
	cardDetect := uint16(reg.Read8(hw.mc1r))
	bits.Set16(&cardDetect, MC1R_FCD)
	bits.Clear16(&cardDetect, MC1R_DDR)
	reg.Write8(hw.mc1r, uint8(cardDetect))

	// enable all status events
	reg.Write16(hw.nister, allInterrupts)
	reg.Write16(hw.eister, allInterrupts)

	return hw.setClockFrequency(mmcIdentificationClockHz)
}

func (hw *SDHCI) dmaBlockLimit() int {
	available := int(hw.Region.Size()) - 2*(dmaAlignment-1)
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
func (hw *SDHCI) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.controllerReady = false
	hw.ready = false
	hw.card = CardInfo{}

	if hw.Base == 0 {
		return errors.New("invalid controller base")
	}

	if hw.GCK == 0 {
		return errors.New("invalid Generic Clock register")
	}

	if hw.Region == nil {
		hw.Region = dma.Default()
	}

	if hw.Region == nil {
		return errors.New("invalid DMA region")
	}

	if hw.Region.End() > 1<<32 {
		return errors.New("DMA memory exceeds ADMA2 address range")
	}

	hw.maxBlocks = hw.dmaBlockLimit()
	if hw.maxBlocks == 0 {
		return errors.New("DMA memory is too small")
	}

	prescaler, err := gck.Prescaler(hw.ParentClock, hw.TargetClock)

	if err != nil {
		return
	}

	hw.ssar = hw.Base + SDMMC_SSAR
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
	hw.nisier = hw.Base + SDMMC_NISIER
	hw.eisier = hw.Base + SDMMC_EISIER
	hw.acesr = hw.Base + SDMMC_ACESR
	hw.capr = hw.Base + SDMMC_CAPR
	hw.ca1r = hw.Base + SDMMC_CA1R
	hw.aesr = hw.Base + SDMMC_AESR
	hw.asar0 = hw.Base + SDMMC_ASAR0
	hw.asar1 = hw.Base + SDMMC_ASAR1
	hw.mc1r = hw.Base + SDMMC_MC1R

	if err = hw.initController(prescaler); err != nil {
		return err
	}

	if !reg.Get(hw.capr, CAPR_ADMA2) {
		return errors.New("controller does not support ADMA2")
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

	if dtd != WRITE && dtd != READ {
		return errors.New("invalid transfer direction")
	}

	for len(buf) > 0 {
		blocks := min(len(buf)/BlockSize, hw.maxBlocks)

		// end each write command on a 16 KiB boundary
		if dtd == WRITE && writeAlignment <= hw.maxBlocks {
			limit := hw.maxBlocks - hw.maxBlocks%writeAlignment
			blocks = min(len(buf)/BlockSize, limit-lba%writeAlignment)
		}

		length := blocks * BlockSize

		if err = hw.transferDMA(index, dtd, uint32(lba), buf[:length], uint16(blocks)); err != nil {
			return
		}

		// CMD13 reports programming errors once the card leaves the
		// programming state
		if dtd == WRITE {
			if err = hw.waitState(CURRENT_STATE_TRAN, WriteTimeout); err != nil {
				return hw.invalidateTransfer(err)
			}
		}

		buf = buf[length:]
		lba += blocks
	}

	return nil
}

// WriteBlocks transfers full blocks of data to the card. On error the contents
// of the target blocks are undefined.
func (hw *SDHCI) WriteBlocks(lba int, buf []byte) error {
	// CMD25 - WRITE_MULTIPLE_BLOCK - write consecutive blocks (CMD24 for one)
	return hw.transferBlocks(25, WRITE, lba, buf)
}

// ReadBlocks transfers full blocks of data from the card.
func (hw *SDHCI) ReadBlocks(lba int, buf []byte) error {
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

func copyDMABuffer(dst []byte, src []byte) {
	const size = 0x40

	// Bounded copies avoid arm64 memmove realigning Device accesses.
	for len(src) >= size {
		copy(dst[:size], src[:size])
		dst, src = dst[size:], src[size:]
	}

	// Payloads and descriptor tables are multiples of eight bytes.
	copy(dst, src)
}

func (hw *SDHCI) transferDMA(index uint16, direction uint32, lba uint32, buf []byte, blocks uint16) (err error) {
	dmaAddress, dmaBuffer := hw.Region.Reserve(len(buf), dmaAlignment)
	defer hw.Region.Release(dmaAddress)

	if direction == WRITE {
		copyDMABuffer(dmaBuffer, buf)
	}

	descriptors := admaTable(dmaAddress, len(buf))
	descriptorAddress, descriptorBuffer := hw.Region.Reserve(len(descriptors), dmaAlignment)
	defer hw.Region.Release(descriptorAddress)
	copyDMABuffer(descriptorBuffer, descriptors)

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

	if multi {
		// predefine the block count so that the card ends the transfer itself
		reg.Write(hw.ssar, uint32(blocks))
		bits.SetN16(&transferMode, TMR_ACMDEN, TMR_ACMDEN_MASK, TMR_ACMDEN_CMD23)
	}

	reg.Write16(hw.tmr, transferMode)

	command := index

	switch {
	case index == 18 && !multi:
		// CMD17 - READ_SINGLE_BLOCK - read one block
		command = 17
	case index == 25 && !multi:
		// CMD24 - WRITE_BLOCK - write one block
		command = 24
	}

	// multiple-block transfers need CMD12 only to abort after an error
	stop := command == 18 || command == 25

	transferTimeout := CommandTimeout + ReadBlockTimeout*time.Duration(blocks)
	stopTimeout := ControllerSetupTimeout

	if direction == WRITE {
		transferTimeout = WriteTimeout
		stopTimeout = WriteTimeout
	}

	status, _, issued, commandErr := hw.runCommand(command, lba, 0)

	if stop && issued {
		defer func() {
			if err != nil {
				err = hw.stopTransmission(err, stopTimeout)
			}
		}()
	}

	if commandErr != nil {
		err = fmt.Errorf("CMD%d transfer failed, %w", command, commandErr)

		if !stop || !issued {
			err = hw.invalidateTransfer(err)
		}

		return
	}

	if responseErr := checkR1(status); responseErr != nil {
		err = fmt.Errorf("CMD%d transfer, %w", command, responseErr)

		if !stop {
			err = hw.invalidateTransfer(err)
		}

		return
	}

	if _, statusErr := hw.pollStatus(1<<NISTR_TRFC, transferTimeout); statusErr != nil {
		err = fmt.Errorf("CMD%d transfer, %w", command, statusErr)

		if !stop {
			err = hw.invalidateTransfer(err)
		}

		return
	}

	if direction == READ {
		copyDMABuffer(buf, dmaBuffer)
	}

	return
}
