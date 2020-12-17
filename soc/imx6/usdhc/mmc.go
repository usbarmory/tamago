// NXP Ultra Secured Digital Host Controller (uSDHC) driver
// https://github.com/f-secure-foundry/tamago
//
// IP: https://www.mobiveil.com/esdhc/
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usdhc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/f-secure-foundry/tamago/bits"
)

// MMC registers
const (
	// p181, 7.1 OCR register, JESD84-B51
	MMC_OCR_BUSY        = 31
	MMC_OCR_ACCESS_MODE = 29
	MMC_OCR_VDD_HV_MAX  = 23
	MMC_OCR_VDD_HV_MIN  = 15
	MMC_OCR_VDD_LV      = 7

	ACCESS_MODE_BYTE   = 0b00
	ACCESS_MODE_SECTOR = 0b10

	// p62, 6.6.1 Command sets and extended settings, JESD84-B51
	MMC_SWITCH_ACCESS  = 24
	MMC_SWITCH_INDEX   = 16
	MMC_SWITCH_VALUE   = 8
	MMC_SWITCH_CMD_SET = 0

	ACCESS_WRITE_BYTE = 0b11

	// p184 7.3 CSD register, JESD84-B51
	MMC_CSD_SPEC_VERS   = 122 + CSD_RSP_OFF
	MMC_CSD_TRAN_SPEED  = 96 + CSD_RSP_OFF
	MMC_CSD_READ_BL_LEN = 80 + CSD_RSP_OFF
	MMC_CSD_C_SIZE      = 62 + CSD_RSP_OFF
	MMC_CSD_C_SIZE_MULT = 47 + CSD_RSP_OFF

	// p186 TRAN_SPEED [103:96], JESD84-B51
	TRAN_SPEED_26MHZ = 0x32

	// p193, 7.4 Extended CSD register, JESD84-B51
	EXT_CSD_SEC_COUNT        = 212
	EXT_CSD_DEVICE_TYPE      = 196
	EXT_CSD_HS_TIMING        = 185
	EXT_CSD_BUS_WIDTH        = 183
	EXT_CSD_PARTITION_CONFIG = 179

	// p224, PARTITION_CONFIG, JESD84-B51
	PARTITION_ACCESS_NONE = 0x0
	PARTITION_ACCESS_RPMB = 0x3

	// p222, 7.4.65 HS_TIMING [185], JESD84-B51
	HS_TIMING_HS    = 0x1
	HS_TIMING_HS200 = 0x2

	// p35, 5.3.2 Bus Speed Modes, JESD84-B51
	HSSDR_MBPS = 52
	HSDDR_MBPS = 104
	HS200_MBPS = 150 // instead of 200 due to NXP ERR010450

	// Base clock divided by 1 (Single Data Rate mode)
	SDCLKFS_HS200 = 0x00
	// SDR104 frequency: 200 / (1 * 1) == 200 MHz

)

// MMC constants
const (
	MMC_DETECT_TIMEOUT     = 1 * time.Second
	MMC_DEFAULT_BLOCK_SIZE = 512
)

// p352, 35.4.6 MMC voltage validation flow chart, IMX6FG
func (hw *USDHC) voltageValidationMMC() (mmc bool, hc bool) {
	var arg uint32

	// CMD1 - SEND_OP_COND
	// p57, 6.4.2 Access mode validation, JESD84-B51

	// sector mode supported
	bits.SetN(&arg, MMC_OCR_ACCESS_MODE, 0b11, ACCESS_MODE_SECTOR)
	// set HV range
	bits.SetN(&arg, MMC_OCR_VDD_HV_MIN, 0x1ff, 0x1ff)

	// p46, 6.3.1 Device reset to Pre-idle state, JESD84-B51
	time.Sleep(1 * time.Millisecond)

	start := time.Now()

	for time.Since(start) <= MMC_DETECT_TIMEOUT {
		// CMD1 - SEND_OP_COND - send operating conditions
		if err := hw.cmd(1, arg, 0, 0); err != nil {
			return false, false
		}

		rsp := hw.rsp(0)

		if bits.Get(&rsp, MMC_OCR_BUSY, 1) == 0 {
			continue
		}

		if bits.Get(&rsp, MMC_OCR_ACCESS_MODE, 0b11) == ACCESS_MODE_SECTOR {
			hc = true
		}

		return true, hc
	}

	return false, false
}

func (hw *USDHC) writeCardRegisterMMC(reg uint32, val uint32) (err error) {
	var arg uint32

	// write MMC_SWITCH_VALUE in register pointed in MMC_SWITCH_INDEX
	bits.SetN(&arg, MMC_SWITCH_ACCESS, 0b11, ACCESS_WRITE_BYTE)
	// set MMC_SWITCH_INDEX to desired register
	bits.SetN(&arg, MMC_SWITCH_INDEX, 0xff, reg)
	// set register value
	bits.SetN(&arg, MMC_SWITCH_VALUE, 0xff, val)

	// CMD6 - SWITCH - switch mode of operation
	err = hw.cmd(6, arg, 0, 0)

	if err != nil {
		return
	}

	// We could use EXT_CSD[GENERIC_CMD6_TIME] for a better tran state
	// timeout, we rather choose to apply a generic timeout for now (as
	// most drivers do).
	err = hw.waitState(CURRENT_STATE_TRAN, 500*time.Millisecond)

	if err != nil {
		return
	}

	if (hw.rsp(0)>>STATUS_SWITCH_ERROR)&1 != 0 {
		err = errors.New("switch error")
	}

	return
}

func (hw *USDHC) detectCapabilitiesMMC(c_size_mult uint32, c_size uint32, read_bl_len uint32) (err error) {
	extCSD := make([]byte, MMC_DEFAULT_BLOCK_SIZE)

	// CMD8 - SEND_EXT_CSD - read extended device data
	if err = hw.transfer(8, READ, 0, 1, MMC_DEFAULT_BLOCK_SIZE, extCSD); err != nil {
		return
	}

	// p128, Table 39 — e•MMC internal sizes and related Units / Granularities, JESD84-B51

	// density greater than 2GB (emulation mode is always assumed)
	if c_size > 0xff {
		hw.card.BlockSize = MMC_DEFAULT_BLOCK_SIZE
		hw.card.Blocks = int(binary.LittleEndian.Uint32(extCSD[EXT_CSD_SEC_COUNT:]))
	} else {
		// p188, 7.3.12 C_SIZE [73:62], JESD84-B51
		hw.card.BlockSize = 2 << (read_bl_len - 1)
		hw.card.Blocks = int((c_size + 1) * (2 << (c_size_mult + 2)))
	}

	// p220, Table 137 — Device types, JESD84-B51
	deviceType := extCSD[EXT_CSD_DEVICE_TYPE]

	if (deviceType>>4)&0b11 > 0 && hw.LowVoltage != nil && hw.LowVoltage() {
		hw.card.Rate = HS200_MBPS
	} else if (deviceType>>2)&0b11 > 0 {
		hw.card.Rate = HSDDR_MBPS
	} else if deviceType&0b11 > 0 {
		hw.card.Rate = HSSDR_MBPS
	}

	return
}

// p352, 35.4.7 MMC card initialization flow chart, IMX6FG
// p58, 6.4.4 Device identification process, JESD84-B51
func (hw *USDHC) initMMC() (err error) {
	var arg uint32
	var bus_width uint32
	var timing uint32
	var clk int
	var ddr bool

	// CMD2 - ALL_SEND_CID - get unique card identification
	if err = hw.cmd(2, arg, 0, 0); err != nil {
		return
	}

	// Send CMD3 with a chosen RCA, with value greater than 1,
	// p301, A.6.1 Bus initialization , JESD84-B51.
	hw.rca = (uint32(hw.n) + 1) << RCA_ADDR

	// CMD3 - SET_RELATIVE_ADDR - set relative card address (RCA),
	if err = hw.cmd(3, hw.rca, 0, 0); err != nil {
		return
	}

	if state := (hw.rsp(0) >> STATUS_CURRENT_STATE) & 0b1111; state != CURRENT_STATE_IDENT {
		return fmt.Errorf("card not in ident state (%d)", state)
	}

	// CMD9 - SEND_CSD - read device data
	if err = hw.cmd(9, hw.rca, 0, 0); err != nil {
		return
	}

	// block count multiplier
	c_size_mult := hw.rspVal(MMC_CSD_C_SIZE_MULT, 0b111)
	// block count
	c_size := hw.rspVal(MMC_CSD_C_SIZE, 0xfff)
	// block size
	read_bl_len := hw.rspVal(MMC_CSD_READ_BL_LEN, 0xf)
	// operating frequency
	mhz := hw.rspVal(MMC_CSD_TRAN_SPEED, 0xff)
	// e•MMC specification version
	ver := hw.rspVal(MMC_CSD_SPEC_VERS, 0xf)

	if mhz == TRAN_SPEED_26MHZ {
		// clear clock
		hw.setClock(-1, -1)
		// set operating frequency
		hw.setClock(DVS_OP, SDCLKFS_OP)
	} else {
		return fmt.Errorf("unexpected TRAN_SPEED %#x", mhz)
	}

	// CMD7 - SELECT/DESELECT CARD - enter transfer state
	if err = hw.cmd(7, hw.rca, 0, 0); err != nil {
		return
	}

	err = hw.waitState(CURRENT_STATE_TRAN, 1*time.Millisecond)

	if err != nil {
		return
	}

	// p223, 7.4.67 BUS_WIDTH [183], JESD84-B51
	switch hw.width {
	case 4:
		bus_width = 1
	case 8:
		bus_width = 2
	default:
		return errors.New("unsupported MMC bus width")
	}

	err = hw.writeCardRegisterMMC(EXT_CSD_BUS_WIDTH, bus_width)

	if err != nil {
		return
	}

	err = hw.detectCapabilitiesMMC(c_size_mult, c_size, read_bl_len)

	if err != nil {
		return
	}

	// Enable High Speed DDR mode only on Version 4.1 or above eMMC cards
	// with supported rate.
	if ver < 4 || hw.card.Rate <= HSSDR_MBPS {
		return
	}

	switch hw.card.Rate {
	case HSDDR_MBPS:
		timing = HS_TIMING_HS
		clk = SDCLKFS_HS_DDR
		ddr = true

		// p223, 7.4.67 BUS_WIDTH [183], JESD84-B51
		switch hw.width {
		case 4:
			bus_width = 5
		case 8:
			bus_width = 6
		}
	case HS200_MBPS:
		timing = HS_TIMING_HS200
		clk = SDCLKFS_HS200
	default:
		return
	}

	// p112, Dual Data Rate mode operation, JESD84-B51
	err = hw.writeCardRegisterMMC(EXT_CSD_HS_TIMING, timing)

	if err != nil {
		return
	}

	err = hw.writeCardRegisterMMC(EXT_CSD_BUS_WIDTH, bus_width)

	if err != nil {
		return
	}

	// clear clock
	hw.setClock(-1, -1)
	// set high speed frequency
	hw.setClock(DVS_HS, clk)

	hw.card.DDR = ddr
	hw.card.HS = true

	return
}

// p224, 7.4.69 PARTITION_CONFIG [179], JESD84-B51
func (hw *USDHC) partitionAccessMMC(access uint32) (err error) {
	return hw.writeCardRegisterMMC(EXT_CSD_PARTITION_CONFIG, access)
}

// p106, 6.6.22.4.3 Authenticated Data Write, JESD84-B51
// p108, 6.6.22.4.4 Authenticated Data Read,  JESD84-B51
func (hw *USDHC) transferRPMB(dtd int, buf []byte) (err error) {
	if !hw.card.MMC {
		return fmt.Errorf("no MMC card detected on uSDHC%d", hw.n)
	}

	if len(buf) != 512 {
		return errors.New("transfer size must be 512")
	}

	hw.Lock()
	hw.rpmb = true

	defer func() {
		hw.rpmb = false
		hw.Unlock()
	}()

	if dtd == WRITE {
		// CMD25 - WRITE_MULTIPLE_BLOCK - write consecutive blocks
		err = hw.transfer(25, WRITE, 0, 1, 512, buf)
	} else {
		// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
		err = hw.transfer(18, READ, 0, 1, 512, buf)
	}

	return
}

// WriteRPMB transfers a single Replay Protected Memory Block (RPMB) data
// frame to the card.
func (hw *USDHC) WriteRPMB(buf []byte) (err error) {
	return hw.transferRPMB(WRITE, buf)
}

// ReadRPMB transfers a single Replay Protected Memory Block (RPMB) data
// frame from the card.
func (hw *USDHC) ReadRPMB(buf []byte) (err error) {
	return hw.transferRPMB(READ, buf)
}
