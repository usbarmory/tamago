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

// Package usdhc implements a driver for Freescale Enhanced Secure Digital
// Host Controller (eSDHC) interface, also known as NXP Ultra Secured Digital
// Host Controller (uSDHC).
//
// It currently supports interfacing with SD/MMC cards up to High Speed mode
// and Dual Data Rate.
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
// Note that due to NXP errata ERR010450 the following maximum values apply:
//  * eMMC  HS200: 150MB/s - 150MHz (instead of 200MB/s - 200MHz), unimplemented
//  * eMMC  DDR52:  90MB/s -  45MHz (instead of 104MB/s -  52MHz), supported
//  *   SD SDR104:  75MB/s - 150MHz (instead of 104MB/s - 208MHz), supported
//  *   SD  DDR50:  45MB/s -  45MHz (instead of  50MB/s -  50MHz), unsupported
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package usdhc

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/internal/reg"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

// USDHC registers (p4012, 58.8 uSDHC Memory Map/Register Definition, IMX6ULLRM).
const (
	USDHC1_BASE = 0x02190000
	USDHC2_BASE = 0x02194000

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
//   * Identification frequency ≤ 400 KHz
//   * Operating frequency ≤ 25 MHz
//   * High frequency ≤ 50 MHz
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
}

// USDHC represents a controller instance.
type USDHC struct {
	sync.Mutex

	// LowVoltage is the board specific function responsible for low
	// voltage switching (SD) or indication (eMMC). The return value
	// reflects whether LV I/O signaling is present.
	LowVoltage func() bool

	// controller index
	n int
	// bus width
	width int
	// clock gate
	cg int
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

	readTimeout  time.Duration
	writeTimeout time.Duration
}

// USDHC1 instance
var USDHC1 = &USDHC{n: 1}

// USDHC2 instance
var USDHC2 = &USDHC{n: 2}

// getRootClock returns the USDHCx_CLK_ROOT clock by reading CSCMR1[USDHCx_CLK_SEL]
// and CSCDR1[USDHCx_PODF]
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM)
func (hw *USDHC) getRootClock() (podf uint32, sel uint32, clock uint32) {
	var podf_pos int
	var clksel_pos int
	var freq uint32

	switch hw.n {
	case 1:
		podf_pos = imx6.CSCDR1_USDHC1_CLK_PODF
		clksel_pos = imx6.CSCMR1_USDHC1_CLK_SEL
	case 2:
		podf_pos = imx6.CSCDR1_USDHC2_CLK_PODF
		clksel_pos = imx6.CSCMR1_USDHC2_CLK_SEL
	default:
		return
	}

	podf = reg.Get(imx6.CCM_CSCDR1, podf_pos, 0b111)
	sel = reg.Get(imx6.CCM_CSCMR1, clksel_pos, 0b1)

	if sel == 1 {
		_, freq = imx6.GetPFD(2, 0)
	} else {
		_, freq = imx6.GetPFD(2, 2)
	}

	clock = freq / (podf + 1)

	return
}

// setRootClock controls the USDHCx_CLK_ROOT clock by setting CSCMR1[USDHCx_CLK_SEL]
// and CSCDR1[USDHCx_PODF]
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM).
func (hw *USDHC) setRootClock(podf uint32, sel uint32) (err error) {
	var podf_pos int
	var clksel_pos int

	if podf < 0 || podf > 7 {
		return errors.New("podf value out of range")
	}

	if sel < 0 || sel > 1 {
		return errors.New("selector value out of range")
	}

	switch hw.n {
	case 1:
		podf_pos = imx6.CSCDR1_USDHC1_CLK_PODF
		clksel_pos = imx6.CSCMR1_USDHC1_CLK_SEL
	case 2:
		podf_pos = imx6.CSCDR1_USDHC2_CLK_PODF
		clksel_pos = imx6.CSCMR1_USDHC2_CLK_SEL
	default:
		return errors.New("invalid interface index")
	}

	reg.SetN(imx6.CCM_CSCDR1, podf_pos, 0b111, podf)
	reg.SetN(imx6.CCM_CSCMR1, clksel_pos, 0b1, sel)

	return
}

// setClock controls the clock of USDHCx_CLK line by setting
// the SDCLKFS and DVS fields of USDHCx_SYS_CTRL register
// p4035, 58.8.12 System Control (uSDHCx_SYS_CTRL), IMX6ULLRM.
func (hw *USDHC) setClock(dvs int, sdclkfs int) {
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

	if hw.card.SD {
		reg.Set(hw.vend_spec, VEND_SPEC_FRC_SDCLK_ON)
	}
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

	// Temporarly disable interrupts other than Buffer Read Ready
	defer reg.Write(hw.int_signal_en, reg.Read(hw.int_signal_en))
	defer reg.Write(hw.int_status_en, reg.Read(hw.int_status_en))
	reg.Write(hw.int_signal_en, INT_SIGNAL_EN_BWRIEN)
	reg.Write(hw.int_status_en, INT_STATUS_EN_BWRSEN)

	tuning_block := make([]byte, blocks)

	for i := 0; i < TUNING_MAX_LOOP_COUNT; i++ {
		// send tuning block command, ignore responses
		hw.transfer(index, READ, 0, 1, blocks, tuning_block)

		ac12_err_status := reg.Read(hw.ac12_err_status)

		if bits.Get(&ac12_err_status, AUTOCMD12_ERR_STATUS_EXE_TUNE, 0b1) == 0 &&
			bits.Get(&ac12_err_status, AUTOCMD12_ERR_STATUS_SMP_CLK_SEL, 0b1) == 1 {
			return nil
		}
	}

	return errors.New("tuning failed")
}

func (hw *USDHC) detect() (sd bool, mmc bool, hc bool, err error) {
	sd, hc = hw.voltageValidationSD()

	if sd {
		return
	}

	mmc, hc = hw.voltageValidationMMC()

	return
}

// Info returns detected card information.
func (hw *USDHC) Info() CardInfo {
	return hw.card
}

// Init initializes the uSDHC controller instance.
func (hw *USDHC) Init(width int) {
	var base uint32

	hw.Lock()

	switch hw.n {
	case 1:
		base = USDHC1_BASE
		hw.cg = imx6.CCGR6_CG1
	case 2:
		base = USDHC2_BASE
		hw.cg = imx6.CCGR6_CG2
	default:
		panic("invalid uSDHC controller instance")
	}

	hw.width = width
	hw.blk_att = base + USDHCx_BLK_ATT
	hw.wtmk_lvl = base + USDHCx_WTMK_LVL
	hw.cmd_arg = base + USDHCx_CMD_ARG
	hw.cmd_xfr = base + USDHCx_CMD_XFR_TYP
	hw.cmd_rsp = base + USDHCx_CMD_RSP0
	hw.prot_ctrl = base + USDHCx_PROT_CTRL
	hw.sys_ctrl = base + USDHCx_SYS_CTRL
	hw.mix_ctrl = base + USDHCx_MIX_CTRL
	hw.pres_state = base + USDHCx_PRES_STATE
	hw.int_status = base + USDHCx_INT_STATUS
	hw.int_status_en = base + USDHCx_INT_STATUS_EN
	hw.int_signal_en = base + USDHCx_INT_SIGNAL_EN
	hw.adma_sys_addr = base + USDHCx_ADMA_SYS_ADDR
	hw.adma_err_status = base + USDHCx_ADMA_ERR_STATUS
	hw.ac12_err_status = base + USDHCx_AUTOCMD12_ERR_STATUS
	hw.vend_spec = base + USDHCx_VEND_SPEC
	hw.vend_spec2 = base + USDHCx_VEND_SPEC2
	hw.tuning_ctrl = base + USDHCx_TUNING_CTRL

	// Generic SD specs read/write timeout rules (applied also to MMC by
	// this driver).
	//
	// p106, 4.6.2.1 Read, SD-PL-7.10
	hw.readTimeout = 100 * time.Millisecond
	// p106, 4.6.2.2 Write, SD-PL-7.10
	hw.writeTimeout = 500 * time.Millisecond

	hw.Unlock()
}

// Detect initializes an SD/MMC card. The highest speed supported by the
// driver, card and controller is automatically selected. Speed modes that
// require voltage switching require definition of function VoltageSelect on
// the USDHC instance, which is up to board packages.
func (hw *USDHC) Detect() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.cg == 0 {
		return errors.New("controller is not initialized")
	}

	// clear card information
	hw.card = CardInfo{}

	// enable clock
	reg.SetN(imx6.CCM_CCGR6, hw.cg, 0b11, 0b11)

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
	hw.setClock(-1, -1)
	// set identification frequency
	hw.setClock(DVS_ID, SDCLKFS_ID)

	// set data timeout counter to SDCLK x 2^28
	reg.Clear(hw.int_status_en, INT_STATUS_EN_DTOESEN)
	reg.SetN(hw.sys_ctrl, SYS_CTRL_DTOCV, 0xf, DTOCV)
	reg.Set(hw.int_status_en, INT_STATUS_EN_DTOESEN)

	// initialize
	reg.Set(hw.sys_ctrl, SYS_CTRL_INITA)
	reg.Wait(hw.sys_ctrl, SYS_CTRL_INITA, 1, 0)

	// CMD0 - GO_IDLE_STATE - reset card
	if err = hw.cmd(0, READ, GO_IDLE_STATE, RSP_NONE, false, false, false, 0); err != nil {
		return
	}

	hw.card.SD, hw.card.MMC, hw.card.HC, err = hw.detect()

	if err != nil {
		return
	}

	if hw.card.SD {
		err = hw.initSD()
	} else if hw.card.MMC {
		err = hw.initMMC()
	} else {
		err = fmt.Errorf("no card detected on uSDHC%d", hw.n)
	}

	if err != nil {
		return
	}

	if !hw.card.DDR {
		// CMD16 - SET_BLOCKLEN - define the block length,
		// only legal In single data rate mode.
		err = hw.cmd(16, READ, uint32(hw.card.BlockSize), RSP_48, true, true, false, 0)
	}

	return
}

// Transfer data from/to the card as specified in:
//   p347, 35.5.1 Reading data from the card, IMX6FG,
//   p354, 35.5.2 Writing data to the card, IMX6FG.
func (hw *USDHC) transfer(index uint32, dtd uint32, offset uint64, blocks uint32, blockSize uint32, buf []byte) (err error) {
	var timeout time.Duration

	if hw.cg == 0 {
		return errors.New("controller is not initialized")
	}

	if blocks == 0 || blockSize == 0 {
		return
	}

	if blocks > 0xffff {
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

	reg.Write(hw.adma_sys_addr, bdAddress)

	if hw.card.HC && index != 6 {
		// p102, 4.3.14 Command Functional Difference in Card Capacity Types, SD-PL-7.10
		offset = offset / uint64(blockSize)
	}

	if dtd == WRITE {
		timeout = hw.writeTimeout * time.Duration(blocks)
		// set write watermark level
		reg.SetN(hw.wtmk_lvl, WTMK_LVL_WR_WML, 0xff, blockSize/4)
	} else {
		timeout = hw.readTimeout * time.Duration(blocks)
		// set read watermark level
		reg.SetN(hw.wtmk_lvl, WTMK_LVL_RD_WML, 0xff, blockSize/4)
	}

	err = hw.cmd(index, dtd, uint32(offset), RSP_48, true, true, true, timeout)
	adma_err := reg.Read(hw.adma_err_status)

	if err != nil {
		return fmt.Errorf("len:%d offset:%#x timeout:%v ADMA:%#x, %v", len(buf), offset, timeout, adma_err, err)
	}

	if adma_err > 0 {
		return fmt.Errorf("len:%d offset:%#x timeout:%v ADMA:%#x", len(buf), offset, timeout, adma_err)
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
