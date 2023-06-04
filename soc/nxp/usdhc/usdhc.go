// NXP Ultra Secured Digital Host Controller (uSDHC) driver
// https://github.com/usbarmory/tamago
//
// IP: https://www.mobiveil.com/esdhc/
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usdhc implements a driver for the Freescale Enhanced Secure Digital
// Host Controller (eSDHC) interface, also known as NXP Ultra Secured Digital
// Host Controller (uSDHC).
//
// The following specifications are adopted:
//   - IMX6ULLRM  - i.MX 6ULL Applications Processor Reference Manual                - Rev 1      2017/11
//   - IMX6FG     - i.MX 6 Series Firmware Guide                                     - Rev 0      2012/11
//   - SD-PL-7.10 - SD Specifications Part 1 Physical Layer Simplified Specification - 7.10       2020/03/25
//   - JESD84-B51 - Embedded Multi-Media Card (e•MMC) Electrical Standard (5.1)      - JESD84-B51 2015/02
//
// The driver currently supports interfacing with SD/MMC cards up to High Speed
// mode and Dual Data Rate.
//
// Higher speed modes for eMMC cards are HS200 (controller supported and driver
// supported) and HS400 mode (unsupported at controller level) [p35, Table 4,
// JESD84-B51].
//
// Higher speed modes for SD cards are SDR50/SDR104 (controller and driver
// supported), DDR50 (controller supported, unimplemented in this driver) and
// UHS-II modes (unsupported at controller level) [p37-38, Figure 3-14 and
// 3-15, SD-PL-7.10].
//
// The highest speed supported by the driver, card and controller is
// automatically selected by Detect().
//
// For eMMC cards, speed mode HS200 requires the target board to have eMMC I/O
// signaling to 1.8V, this must be advertised by the board package by defining
// LowVoltage() on the relevant USDHC instance.
//
// For SD cards, speed modes SDR50/SDR104 require the target board to switch SD
// I/O signaling to 1.8V, the switching procedure must be implemented by the
// board package by defining LowVoltage() on the relevant USDHC instance.
//
// Note that due to NXP errata ERR010450 the following maximum theoretical
// limits apply:
//   - eMMC  HS200: 150MB/s - 150MHz (instead of 200MB/s - 200MHz), supported
//   - eMMC  DDR52:  90MB/s -  45MHz (instead of 104MB/s -  52MHz), supported
//   - SD   SDR104:  75MB/s - 150MHz (instead of 104MB/s - 208MHz), supported
//   - SD    DDR50:  45MB/s -  45MHz (instead of  50MB/s -  50MHz), unsupported
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package usdhc

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// USDHC registers (p4012, 58.8 uSDHC Memory Map/Register Definition, IMX6ULLRM).
const (
	USDHCx_BLK_ATT  = 0x04
	BLK_ATT_BLKCNT  = 16
	BLK_ATT_BLKSIZE = 0

	USDHCx_CMD_ARG = 0x08

	USDHCx_CMD_XFR_TYP = 0x0c
	CMD_XFR_TYP_CMDINX = 24
	CMD_XFR_TYP_CMDTYP = 22
	CMD_XFR_TYP_DPSEL  = 21
	CMD_XFR_TYP_CICEN  = 20
	CMD_XFR_TYP_CCCEN  = 19
	CMD_XFR_TYP_RSPTYP = 16

	USDHCx_CMD_RSP0 = 0x10
	USDHCx_CMD_RSP1 = 0x14
	USDHCx_CMD_RSP2 = 0x18
	USDHCx_CMD_RSP3 = 0x1c

	USDHCx_PRES_STATE = 0x24
	PRES_STATE_DLSL   = 24
	PRES_STATE_WPSPL  = 19
	PRES_STATE_BREN   = 11
	PRES_STATE_SDSTB  = 3
	PRES_STATE_CDIHB  = 1
	PRES_STATE_CIHB   = 0

	USDHCx_PROT_CTRL = 0x28
	PROT_CTRL_DMASEL = 8
	PROT_CTRL_EMODE  = 4
	PROT_CTRL_DTW    = 1

	USDHCx_SYS_CTRL  = 0x2c
	SYS_CTRL_INITA   = 27
	SYS_CTRL_RSTD    = 26
	SYS_CTRL_RSTC    = 25
	SYS_CTRL_RSTA    = 24
	SYS_CTRL_DTOCV   = 16
	SYS_CTRL_SDCLKFS = 8
	SYS_CTRL_DVS     = 4

	USDHCx_INT_STATUS = 0x30
	INT_STATUS_DMAE   = 28
	INT_STATUS_TNE    = 26
	INT_STATUS_AC12E  = 24
	INT_STATUS_CIE    = 19
	INT_STATUS_CEBE   = 18
	INT_STATUS_CCE    = 17
	INT_STATUS_CTOE   = 16
	INT_STATUS_CRM    = 7
	INT_STATUS_BRR    = 5
	INT_STATUS_TC     = 1
	INT_STATUS_CC     = 0

	USDHCx_INT_STATUS_EN  = 0x34
	INT_STATUS_EN_DTOESEN = 20
	INT_STATUS_EN_BWRSEN  = 4

	USDHCx_INT_SIGNAL_EN = 0x38
	INT_SIGNAL_EN_BWRIEN = 4

	USDHCx_AUTOCMD12_ERR_STATUS      = 0x3c
	AUTOCMD12_ERR_STATUS_SMP_CLK_SEL = 23
	AUTOCMD12_ERR_STATUS_EXE_TUNE    = 22

	USDHCx_WTMK_LVL = 0x44
	WTMK_LVL_WR_WML = 16
	WTMK_LVL_RD_WML = 0

	USDHCx_MIX_CTRL       = 0x48
	MIX_CTRL_FBCLK_SEL    = 25
	MIX_CTRL_AUTO_TUNE_EN = 24
	MIX_CTRL_SMP_CLK_SEL  = 23
	MIX_CTRL_EXE_TUNE     = 22
	MIX_CTRL_MSBSEL       = 5
	MIX_CTRL_DTDSEL       = 4
	MIX_CTRL_DDR_EN       = 3
	MIX_CTRL_AC12EN       = 2
	MIX_CTRL_BCEN         = 1
	MIX_CTRL_DMAEN        = 0

	USDHCx_ADMA_ERR_STATUS = 0x54
	USDHCx_ADMA_SYS_ADDR   = 0x58

	USDHCx_VEND_SPEC       = 0xc0
	VEND_SPEC_FRC_SDCLK_ON = 8
	VEND_SPEC_VSELECT      = 1

	USDHCx_VEND_SPEC2         = 0xc8
	VEND_SPEC2_TUNING_1bit_EN = 5
	VEND_SPEC2_TUNING_8bit_EN = 4

	USDHCx_TUNING_CTRL           = 0xcc
	TUNING_CTRL_STD_TUNING_EN    = 24
	TUNING_CTRL_TUNING_STEP      = 16
	TUNING_CTRL_TUNING_START_TAP = 0
)

// Configuration constants (p348, 35.4.2 Frequency divider configuration,
// IMX6FG) to support the following frequencies:
//   - Identification frequency ≤ 400 KHz
//   - Operating frequency ≤ 25 MHz
//   - High frequency ≤ 50 MHz
const (
	// p346, 35.2 Clocks, IMX6FG.
	//
	// The root clock is derived by default from PLL2 PFD2 (396MHz) with divide
	// by 2, therefore 198MHz.

	// Data Timeout Counter Value: SDCLK x 2** 29
	DTOCV = 0xf

	// Divide-by-8
	DVS_ID = 7
	// Root clock divided by 64
	SDCLKFS_ID = 0x20
	// Identification frequency: 198 / (8 * 64) == ~400 KHz

	// Divide-by-2
	DVS_OP = 1
	// Root clock divided by 4
	SDCLKFS_OP = 0x02
	// Operating frequency: 198 / (2 * 4) == 24.75 MHz

	// PLL2 PFD2 clock divided by 2
	ROOTCLK_HS_SDR = 1
	// Root clock frequency: 396 MHz / (1 + 1) = 198 MHz

	// Divide-by-1
	DVS_HS = 0
	// Root clock divided by 4 (Single Data Rate mode)
	SDCLKFS_HS_SDR = 0x02
	// Root clock divided by 4 (Dual Data Rate mode)
	SDCLKFS_HS_DDR = 0x01
	// High Speed frequency: 198 / (1 * 4) == 49.5 MHz
)

// CardInfo holds detected card information.
type CardInfo struct {
	// eMMC card
	MMC bool
	// SD card
	SD bool
	// High Capacity
	HC bool
	// High Speed
	HS bool
	// Dual Data Rate
	DDR bool
	// Maximum throughput (on this controller)
	Rate int

	// Block Size
	BlockSize int
	// Capacity
	Blocks int

	// device identification number
	CID [16]byte
}

// USDHC represents an SD/MMC controller instance.
type USDHC struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Clock setup function
	SetClock func(index int, podf uint32, clksel uint32) error

	// LowVoltage is the board specific function responsible for voltage
	// switching (SD) or low voltage indication (eMMC).
	//
	// The return value reflects whether the voltage switch (SD) or
	// low voltage indication (MMC) is successful.
	LowVoltage func(enable bool) bool

	// bus width
	width int
	// Relative Card Address
	rca uint32

	// control registers
	blk_att         uint32
	wtmk_lvl        uint32
	cmd_arg         uint32
	cmd_xfr         uint32
	cmd_rsp         uint32
	prot_ctrl       uint32
	sys_ctrl        uint32
	mix_ctrl        uint32
	pres_state      uint32
	int_status      uint32
	int_status_en   uint32
	int_signal_en   uint32
	adma_sys_addr   uint32
	adma_err_status uint32
	ac12_err_status uint32
	vend_spec       uint32
	vend_spec2      uint32
	tuning_ctrl     uint32

	// detected card properties
	card CardInfo

	// eMMC Replay Protected Memory Block (RPMB) operation
	rpmb bool

	readTimeout  time.Duration
	writeTimeout time.Duration
}

// setFreq controls the clock of USDHCx_CLK line by setting
// the SDCLKFS and DVS fields of USDHCx_SYS_CTRL register
// p4035, 58.8.12 System Control (uSDHCx_SYS_CTRL), IMX6ULLRM.
func (hw *USDHC) setFreq(dvs int, sdclkfs int) {
	// Prevent possible glitch on the card clock as noted in
	// p4011, 58.7.7 Change Clock Frequency, IMX6ULLRM.
	reg.Clear(hw.vend_spec, VEND_SPEC_FRC_SDCLK_ON)

	if dvs < 0 && sdclkfs < 0 {
		return
	}

	// Wait for stable clock as noted in
	// p4038, DVS[3:0], IMX6ULLRM.
	reg.Wait(hw.pres_state, PRES_STATE_SDSTB, 1, 1)

	sys := reg.Read(hw.sys_ctrl)

	// p348, 35.4.2 Frequency divider configuration, IMX6FG
	bits.SetN(&sys, SYS_CTRL_DVS, 0xf, uint32(dvs))
	bits.SetN(&sys, SYS_CTRL_SDCLKFS, 0xff, uint32(sdclkfs))

	reg.Write(hw.sys_ctrl, sys)
	reg.Wait(hw.pres_state, PRES_STATE_SDSTB, 1, 1)

	reg.SetTo(hw.vend_spec, VEND_SPEC_FRC_SDCLK_ON, hw.card.SD)
}

// executeTuning performs the bus tuning, `cmd` should be set to the relevant
// send tuning block command index, `blocks` represents the number of tuning
// blocks.
func (hw *USDHC) executeTuning(index uint32, blocks uint32) error {
	reg.SetN(hw.tuning_ctrl, TUNING_CTRL_TUNING_STEP, 0b111, TUNING_STEP)
	reg.SetN(hw.tuning_ctrl, TUNING_CTRL_TUNING_START_TAP, 0xff, TUNING_START_TAP)
	reg.Set(hw.tuning_ctrl, TUNING_CTRL_STD_TUNING_EN)

	reg.Clear(hw.ac12_err_status, AUTOCMD12_ERR_STATUS_SMP_CLK_SEL)
	reg.Set(hw.ac12_err_status, AUTOCMD12_ERR_STATUS_EXE_TUNE)

	reg.Set(hw.mix_ctrl, MIX_CTRL_FBCLK_SEL)
	reg.Set(hw.mix_ctrl, MIX_CTRL_AUTO_TUNE_EN)

	// temporarily disable all interrupts other than Buffer Read Ready
	defer reg.Write(hw.int_signal_en, reg.Read(hw.int_signal_en))
	defer reg.Write(hw.int_status_en, reg.Read(hw.int_status_en))
	reg.Write(hw.int_signal_en, INT_SIGNAL_EN_BWRIEN)
	reg.Write(hw.int_status_en, INT_STATUS_EN_BWRSEN)

	// temporarily lower read timeout for faster tuning
	defer func(d time.Duration) {
		hw.readTimeout = d
	}(hw.readTimeout)
	hw.readTimeout = 1 * time.Millisecond

	tuning_block := make([]byte, blocks)

	for i := 0; i < TUNING_MAX_LOOP_COUNT; i++ {
		// send tuning block command, ignore responses
		hw.transfer(index, READ, 0, 1, blocks, tuning_block)

		ac12_err_status := reg.Read(hw.ac12_err_status)

		if bits.Get(&ac12_err_status, AUTOCMD12_ERR_STATUS_EXE_TUNE, 1) == 0 &&
			bits.Get(&ac12_err_status, AUTOCMD12_ERR_STATUS_SMP_CLK_SEL, 1) == 1 {
			return nil
		}
	}

	return errors.New("tuning failed")
}

// Info returns detected card information.
func (hw *USDHC) Info() CardInfo {
	return hw.card
}

// Init initializes the uSDHC controller instance.
func (hw *USDHC) Init(width int) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Index == 0 || hw.Base == 0 || hw.SetClock == nil || hw.CCGR == 0 {
		panic("invalid uSDHC controller instance")
	}

	hw.width = width
	hw.blk_att = hw.Base + USDHCx_BLK_ATT
	hw.wtmk_lvl = hw.Base + USDHCx_WTMK_LVL
	hw.cmd_arg = hw.Base + USDHCx_CMD_ARG
	hw.cmd_xfr = hw.Base + USDHCx_CMD_XFR_TYP
	hw.cmd_rsp = hw.Base + USDHCx_CMD_RSP0
	hw.prot_ctrl = hw.Base + USDHCx_PROT_CTRL
	hw.sys_ctrl = hw.Base + USDHCx_SYS_CTRL
	hw.mix_ctrl = hw.Base + USDHCx_MIX_CTRL
	hw.pres_state = hw.Base + USDHCx_PRES_STATE
	hw.int_status = hw.Base + USDHCx_INT_STATUS
	hw.int_status_en = hw.Base + USDHCx_INT_STATUS_EN
	hw.int_signal_en = hw.Base + USDHCx_INT_SIGNAL_EN
	hw.adma_sys_addr = hw.Base + USDHCx_ADMA_SYS_ADDR
	hw.adma_err_status = hw.Base + USDHCx_ADMA_ERR_STATUS
	hw.ac12_err_status = hw.Base + USDHCx_AUTOCMD12_ERR_STATUS
	hw.vend_spec = hw.Base + USDHCx_VEND_SPEC
	hw.vend_spec2 = hw.Base + USDHCx_VEND_SPEC2
	hw.tuning_ctrl = hw.Base + USDHCx_TUNING_CTRL

	// Generic SD specs read/write timeout rules (applied also to MMC by
	// this driver).
	//
	// p106, 4.6.2.1 Read, SD-PL-7.10
	hw.readTimeout = 100 * time.Millisecond
	// p106, 4.6.2.2 Write, SD-PL-7.10
	hw.writeTimeout = 500 * time.Millisecond

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
}

// Detect initializes an SD/MMC card. The highest speed supported by the
// driver, card and controller is automatically selected. Speed modes that
// require voltage switching require definition of function VoltageSelect on
// the USDHC instance, which is up to board packages.
func (hw *USDHC) Detect() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.sys_ctrl == 0 {
		return errors.New("controller is not initialized")
	}

	// check if a card has already been detected and not removed since
	if reg.Get(hw.int_status, INT_STATUS_CRM, 1) == 0 && (hw.card.MMC || hw.card.SD) {
		return
	}

	// clear card information
	hw.card = CardInfo{}

	// soft reset uSDHC
	reg.Set(hw.sys_ctrl, SYS_CTRL_RSTA)
	reg.Wait(hw.sys_ctrl, SYS_CTRL_RSTA, 1, 0)

	// A soft reset fails to clear MIX_CTRL register, clear it all except
	// tuning bits.
	mix := reg.Read(hw.mix_ctrl)
	bits.Clear(&mix, MIX_CTRL_FBCLK_SEL)
	bits.Clear(&mix, MIX_CTRL_AUTO_TUNE_EN)
	bits.Clear(&mix, MIX_CTRL_SMP_CLK_SEL)
	bits.Clear(&mix, MIX_CTRL_EXE_TUNE)
	reg.Write(hw.mix_ctrl, mix)

	// data transfer width, default to 1-bit mode
	dtw := 0b00

	switch hw.width {
	case 1:
		dtw = 0b00
	case 4:
		dtw = 0b01
	case 8:
		dtw = 0b10
	default:
		return errors.New("unsupported controller data transfer width")
	}

	// set data transfer width
	reg.SetN(hw.prot_ctrl, PROT_CTRL_DTW, 0b11, uint32(dtw))
	// set little endian mode
	reg.SetN(hw.prot_ctrl, PROT_CTRL_EMODE, 0b11, 0b10)

	// clear clock
	hw.setFreq(-1, -1)
	// set identification frequency
	hw.setFreq(DVS_ID, SDCLKFS_ID)

	// set data timeout counter to SDCLK x 2^28
	reg.Clear(hw.int_status_en, INT_STATUS_EN_DTOESEN)
	reg.SetN(hw.sys_ctrl, SYS_CTRL_DTOCV, 0xf, DTOCV)
	reg.Set(hw.int_status_en, INT_STATUS_EN_DTOESEN)

	// initialize
	reg.Set(hw.sys_ctrl, SYS_CTRL_INITA)
	reg.Wait(hw.sys_ctrl, SYS_CTRL_INITA, 1, 0)

	// CMD0 - GO_IDLE_STATE - reset card
	if err = hw.cmd(0, GO_IDLE_STATE, 0, 0); err != nil {
		return
	}

	if hw.voltageValidationSD() {
		err = hw.initSD()
	} else if hw.voltageValidationMMC() {
		err = hw.initMMC()
	} else {
		return fmt.Errorf("no card detected on uSDHC%d", hw.Index)
	}

	if err != nil {
		return
	}

	if !hw.card.DDR {
		// CMD16 - SET_BLOCKLEN - define the block length,
		// only legal In single data rate mode.
		err = hw.cmd(16, uint32(hw.card.BlockSize), 0, 0)
	}

	return
}

// transfer data from/to the card as specified in:
//
//	p347, 35.5.1 Reading data from the card, IMX6FG,
//	p354, 35.5.2 Writing data to the card, IMX6FG.
func (hw *USDHC) transfer(index uint32, dtd uint32, arg uint64, blocks uint32, blockSize uint32, buf []byte) (err error) {
	var timeout time.Duration

	if hw.blk_att == 0 {
		return errors.New("controller is not initialized")
	}

	if blocks == 0 || blockSize == 0 {
		return
	}

	if (blocks & 0xffff) > 0xffff {
		return errors.New("transfer size cannot exceed 65535 blocks")
	}

	// State polling cannot be issued while tuning (CMD19 and CMD21).
	if !(index == 19 || index == 21) {
		if err = hw.waitState(CURRENT_STATE_TRAN, 1*time.Millisecond); err != nil {
			return
		}
	}

	// set block size
	reg.SetN(hw.blk_att, BLK_ATT_BLKSIZE, 0x1fff, blockSize)
	// set block count
	reg.SetN(hw.blk_att, BLK_ATT_BLKCNT, 0xffff, blocks)

	bufAddress := dma.Alloc(buf, 32)
	defer dma.Free(bufAddress)

	// ADMA2 descriptor
	bd := &ADMABufferDescriptor{}
	bd.Init(bufAddress, len(buf))

	bdAddress := dma.Alloc(bd.Bytes(), 4)
	defer dma.Free(bdAddress)

	reg.Write(hw.adma_sys_addr, uint32(bdAddress))

	if hw.card.HC && (index == 18 || index == 25) {
		// p102, 4.3.14 Command Functional Difference in Card Capacity Types, SD-PL-7.10
		arg = arg / uint64(blockSize)
	}

	if hw.rpmb {
		err = hw.partitionAccessMMC(PARTITION_ACCESS_RPMB)

		if err != nil {
			return
		}

		// CMD23 - SET_BLOCK_COUNT - define read/write block count
		if err = hw.cmd(23, blocks, 0, 0); err != nil {
			return
		}

		defer hw.partitionAccessMMC(PARTITION_ACCESS_NONE)
	}

	switch dtd {
	case WRITE:
		timeout = hw.writeTimeout * time.Duration(blocks)
		// set write watermark level
		reg.SetN(hw.wtmk_lvl, WTMK_LVL_WR_WML, 0xff, blockSize/4)
	case READ:
		timeout = hw.readTimeout * time.Duration(blocks)
		// set read watermark level
		reg.SetN(hw.wtmk_lvl, WTMK_LVL_RD_WML, 0xff, blockSize/4)
	}

	err = hw.cmd(index, uint32(arg), blocks, timeout)
	adma_err := reg.Read(hw.adma_err_status)

	if err != nil {
		return fmt.Errorf("len:%d arg:%#x timeout:%v ADMA:%#x, %v", len(buf), arg, timeout, adma_err, err)
	}

	if adma_err > 0 {
		return fmt.Errorf("len:%d arg:%#x timeout:%v ADMA:%#x", len(buf), arg, timeout, adma_err)
	}

	if dtd == READ {
		dma.Read(bufAddress, 0, buf)
	}

	return
}

func (hw *USDHC) transferBlocks(index uint32, dtd uint32, lba int, buf []byte) (err error) {
	blockSize := hw.card.BlockSize
	offset := uint64(lba) * uint64(blockSize)
	size := len(buf)

	if size == 0 || blockSize == 0 {
		return
	}

	if size%blockSize != 0 {
		return fmt.Errorf("write size must be %d bytes aligned", blockSize)
	}

	blocks := size / blockSize

	hw.Lock()
	defer hw.Unlock()

	return hw.transfer(index, dtd, offset, uint32(blocks), uint32(blockSize), buf)
}

// WriteBlocks transfers full blocks of data to the card.
func (hw *USDHC) WriteBlocks(lba int, buf []byte) (err error) {
	// CMD25 - WRITE_MULTIPLE_BLOCK - write consecutive blocks
	return hw.transferBlocks(25, WRITE, lba, buf)
}

// ReadBlocks transfers full blocks of data from the card.
func (hw *USDHC) ReadBlocks(lba int, buf []byte) (err error) {
	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
	return hw.transferBlocks(18, READ, lba, buf)
}

// Read transfers data from the card.
func (hw *USDHC) Read(offset int64, size int64) (buf []byte, err error) {
	blockSize := int64(hw.card.BlockSize)

	if size == 0 || blockSize == 0 {
		return
	}

	blockOffset := offset % blockSize
	blocks := (blockOffset + size) / blockSize

	if blocks == 0 {
		blocks = 1
	} else if (offset+size)%blockSize != 0 {
		blocks += 1
	}

	bufSize := int(blocks * blockSize)

	// data buffer
	buf = make([]byte, bufSize)

	hw.Lock()
	defer hw.Unlock()

	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
	err = hw.transfer(18, READ, uint64(offset), uint32(blocks), uint32(blockSize), buf)

	if err != nil {
		return
	}

	trim := size % blockSize

	if hw.card.HC {
		if blockOffset != 0 || trim > 0 {
			buf = buf[blockOffset : blockOffset+size]
		}
	} else if trim > 0 {
		buf = buf[:offset+size]
	}

	return
}
