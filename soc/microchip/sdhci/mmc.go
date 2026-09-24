// Microchip Secure Digital Host Controller Interface support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sdhci

import (
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// MMC registers
const (
	MMC_OCR_BUSY        = 31
	MMC_OCR_ACCESS_MODE = 29
	MMC_OCR_VDD_HV_MIN  = 15

	ACCESS_MODE_SECTOR = 0b10

	MMC_SWITCH_ACCESS = 24
	MMC_SWITCH_INDEX  = 16
	MMC_SWITCH_VALUE  = 8

	ACCESS_WRITE_BYTE = 0b11

	STATUS_CURRENT_STATE      = 9
	STATUS_CURRENT_STATE_MASK = 0xf
	STATUS_READY_FOR_DATA     = 8
	CURRENT_STATE_TRAN        = 4

	EXT_CSD_SEC_COUNT   = 212
	EXT_CSD_DEVICE_TYPE = 196
	EXT_CSD_REV         = 192
	EXT_CSD_HS_TIMING   = 185
	EXT_CSD_BUS_WIDTH   = 183
	EXT_CSD_CACHE_CTRL  = 33
	EXT_CSD_FLUSH_CACHE = 32

	CACHE_ENABLED        = 0
	BUS_WIDTH_8          = 2
	BUS_WIDTH_8_DDR      = 6
	DEVICE_TYPE_HS52     = 1
	DEVICE_TYPE_HS_DDR52 = 2
	HS_TIMING_HS         = 1
)

const MMC_DEFAULT_BLOCK_SIZE = 512

var (
	// MMCDetectTimeout controls how long card detection waits for power-up.
	MMCDetectTimeout = 1 * time.Second

	// MMCPowerUpDelay controls the initial delay before card detection.
	MMCPowerUpDelay = 10 * time.Millisecond
)

func (hw *SDHCI) voltageValidationMMC() (ready bool) {
	var arg uint32

	// sector mode supported
	bits.SetN(&arg, MMC_OCR_ACCESS_MODE, 0b11, ACCESS_MODE_SECTOR)
	// set HV range
	bits.SetN(&arg, MMC_OCR_VDD_HV_MIN, 0x1ff, 0x1ff)

	time.Sleep(MMCPowerUpDelay)
	start := time.Now()

	for time.Since(start) <= MMCDetectTimeout {
		// CMD1 - SEND_OP_COND - send operating conditions
		response, err := hw.cmd(1, arg)

		if err == nil && bits.GetN(&response, MMC_OCR_BUSY, 1) == 1 {
			hw.card.OCR = response
			return bits.GetN(&response, MMC_OCR_ACCESS_MODE, 0b11) == ACCESS_MODE_SECTOR
		}

		runtime.Gosched()
	}

	return
}

func (hw *SDHCI) writeCardRegisterMMC(register uint32, value uint32, timeout time.Duration) (err error) {
	var arg uint32

	// write MMC_SWITCH_VALUE in register pointed in MMC_SWITCH_INDEX
	bits.SetN(&arg, MMC_SWITCH_ACCESS, 0b11, ACCESS_WRITE_BYTE)
	// set MMC_SWITCH_INDEX to desired register
	bits.SetN(&arg, MMC_SWITCH_INDEX, 0xff, register)
	// set register value
	bits.SetN(&arg, MMC_SWITCH_VALUE, 0xff, value)

	// CMD6 - SWITCH - switch mode of operation
	status, err := hw.cmd(6, arg)

	if err != nil {
		return
	}

	if err = checkR1(status); err != nil {
		return
	}

	return hw.waitState(CURRENT_STATE_TRAN, timeout)
}

func (hw *SDHCI) detectCapabilitiesMMC() (err error) {
	extCSD := make([]byte, MMC_DEFAULT_BLOCK_SIZE)

	// CMD8 - SEND_EXT_CSD - read extended device data
	if err = hw.readExtCSD(extCSD); err != nil {
		return
	}

	hw.card.BlockSize = MMC_DEFAULT_BLOCK_SIZE
	hw.card.Blocks = int(binary.LittleEndian.Uint32(extCSD[EXT_CSD_SEC_COUNT:]))

	if hw.card.Blocks == 0 {
		return errors.New("CMD8 SEND_EXT_CSD: card reports zero sectors")
	}

	hw.card.ExtCSDRev = extCSD[EXT_CSD_REV]
	hw.card.DeviceType = extCSD[EXT_CSD_DEVICE_TYPE]
	cacheControl := uint32(extCSD[EXT_CSD_CACHE_CTRL])
	hw.card.CacheEnabled = bits.Get(&cacheControl, CACHE_ENABLED)

	return
}

func (hw *SDHCI) enableHighSpeedMMC() (err error) {
	deviceType := uint32(hw.card.DeviceType)

	// p220, Table 137, Device types, JESD84-B51
	if !bits.Get(&deviceType, DEVICE_TYPE_HS52) || !reg.Get(hw.capr, CAPR_HSSUP) {
		return
	}

	// p222, 7.4.65 HS_TIMING [185], JESD84-B51
	if err = hw.writeCardRegisterMMC(EXT_CSD_HS_TIMING, HS_TIMING_HS, ControllerSetupTimeout); err != nil {
		return fmt.Errorf("CMD6 SWITCH high-speed timing: %w", err)
	}

	// stop the card clock while moving output to the rising edge
	reg.Clear16(hw.ccr, CCR_SDCLKEN)

	hostControl := uint16(reg.Read8(hw.hc1r))
	bits.Set16(&hostControl, HC1R_HSEN)
	reg.Write8(hw.hc1r, uint8(hostControl))

	return hw.setClockFrequency(mmcHighSpeedClockHz)
}

func (hw *SDHCI) enableDualDataRateMMC() (err error) {
	deviceType := uint32(hw.card.DeviceType)
	hostControl := uint32(reg.Read8(hw.hc1r))

	// p220, Table 137, Device types, JESD84-B51
	if !hw.DualDataRate || !bits.Get(&hostControl, HC1R_HSEN) ||
		!bits.Get(&deviceType, DEVICE_TYPE_HS_DDR52) || !reg.Get(hw.ca1r, CA1R_DDR50SUP) {
		return
	}

	// p223, 7.4.67 BUS_WIDTH [183], JESD84-B51
	if err = hw.writeCardRegisterMMC(EXT_CSD_BUS_WIDTH, BUS_WIDTH_8_DDR, ControllerSetupTimeout); err != nil {
		return fmt.Errorf("CMD6 SWITCH dual data rate bus width: %w", err)
	}

	// stop the card clock while selecting dual data rate sampling
	reg.Clear16(hw.ccr, CCR_SDCLKEN)

	mode := uint16(reg.Read8(hw.mc1r))
	bits.Set16(&mode, MC1R_DDR)
	reg.Write8(hw.mc1r, uint8(mode))
	reg.SetN16(hw.hc2r, HC2R_UHSMS, HC2R_UHSMS_MASK, UHSMS_DDR50)

	reg.Set16(hw.ccr, CCR_SDCLKEN)

	extCSD := make([]byte, MMC_DEFAULT_BLOCK_SIZE)

	if err = hw.readExtCSD(extCSD); err != nil {
		return fmt.Errorf("dual data rate EXT_CSD read: %w", err)
	}

	if extCSD[EXT_CSD_BUS_WIDTH] != BUS_WIDTH_8_DDR ||
		int(binary.LittleEndian.Uint32(extCSD[EXT_CSD_SEC_COUNT:])) != hw.card.Blocks {
		return errors.New("dual data rate EXT_CSD mismatch")
	}

	return
}

func (hw *SDHCI) initMMC() (err error) {
	// CMD2 - ALL_SEND_CID - get unique card identification
	if _, err = hw.cmd(2, 0); err != nil {
		return fmt.Errorf("CMD2 ALL_SEND_CID failed: %w", err)
	}

	hw.card.CID = hw.response136()

	// CMD3 - SET_RELATIVE_ADDR - set relative card address (RCA)
	hw.card.RCA = 1
	status, err := hw.cmd(3, uint32(hw.card.RCA)<<16)

	if err != nil {
		return fmt.Errorf("CMD3 SET_RELATIVE_ADDR failed: %w", err)
	}

	if err = checkR1(status); err != nil {
		return fmt.Errorf("CMD3 SET_RELATIVE_ADDR: %w", err)
	}

	// CMD7 - SELECT/DESELECT CARD - enter transfer state
	status, err = hw.cmd(7, uint32(hw.card.RCA)<<16)

	if err != nil {
		return fmt.Errorf("CMD7 SELECT_CARD failed: %w", err)
	}

	if err = checkR1(status); err != nil {
		return fmt.Errorf("CMD7 SELECT_CARD: %w", err)
	}

	// p223, 7.4.67 BUS_WIDTH [183], JESD84-B51
	if err = hw.writeCardRegisterMMC(EXT_CSD_BUS_WIDTH, BUS_WIDTH_8, ControllerSetupTimeout); err != nil {
		return fmt.Errorf("CMD6 SWITCH bus width: %w", err)
	}

	// select 8-bit bus width
	hostControl := uint16(reg.Read8(hw.hc1r))
	bits.Clear16(&hostControl, HC1R_DW_4BIT)
	bits.Set16(&hostControl, HC1R_EXTDW)
	reg.Write8(hw.hc1r, uint8(hostControl))

	if err = hw.setClockFrequency(mmcLegacyClockHz); err != nil {
		return
	}

	if err = hw.detectCapabilitiesMMC(); err != nil {
		return
	}

	if err = hw.enableHighSpeedMMC(); err != nil {
		return
	}

	return hw.enableDualDataRateMMC()
}

// Detect initializes the eMMC card attached to an initialized controller.
func (hw *SDHCI) Detect() (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.ready = false
	hw.card = CardInfo{}

	if !hw.controllerReady {
		return errors.New("controller is not initialized")
	}

	defer func() {
		if err != nil {
			hw.controllerReady = false
			hw.card = CardInfo{}
		} else {
			hw.ready = true
		}
	}()

	// CMD0 - GO_IDLE_STATE - reset card
	if _, err = hw.cmd(0, 0); err != nil {
		return fmt.Errorf("CMD0 GO_IDLE_STATE failed: %w", err)
	}

	if !hw.voltageValidationMMC() {
		if hw.card.OCR != 0 {
			return fmt.Errorf("CMD1 SEND_OP_COND: card does not use sector addressing (OCR=0x%08x)", hw.card.OCR)
		}

		return errors.New("CMD1 SEND_OP_COND: card did not power up")
	}

	return hw.initMMC()
}

// Ready reports whether Detect completed and no transfer failure has forced
// controller reinitialization.
func (hw *SDHCI) Ready() bool {
	return hw.ready
}

// Info returns detected card metadata.
func (hw *SDHCI) Info() CardInfo {
	return hw.card
}

// ExtCSD reads the 512-byte Extended CSD register.
func (hw *SDHCI) ExtCSD(buf []byte) error {
	hw.Lock()
	defer hw.Unlock()

	if !hw.ready {
		return ErrNotInitialized
	}

	return hw.readExtCSD(buf)
}

func (hw *SDHCI) readExtCSD(buf []byte) error {
	if len(buf) != BlockSize {
		return errors.New("Extended CSD buffer must be 512 bytes")
	}

	// CMD8 - SEND_EXT_CSD - read extended device data
	return hw.readBlocks(8, 0, buf, 1)
}

// Sync flushes an enabled eMMC write cache to non-volatile storage.
func (hw *SDHCI) Sync() error {
	hw.Lock()
	defer hw.Unlock()

	if !hw.ready {
		return ErrNotInitialized
	}

	if !hw.card.CacheEnabled {
		if err := hw.waitState(CURRENT_STATE_TRAN, WriteTimeout); err != nil {
			return hw.invalidateTransfer(err)
		}

		return nil
	}

	if err := hw.writeCardRegisterMMC(EXT_CSD_FLUSH_CACHE, 1, WriteTimeout); err != nil {
		return hw.invalidateTransfer(fmt.Errorf("CMD6 FLUSH_CACHE: %w", err))
	}

	return nil
}
