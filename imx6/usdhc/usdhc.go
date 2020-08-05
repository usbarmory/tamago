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
	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/internal/reg"
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

	USDHCx_INT_SIGNAL_EN        = 0x38
	USDHCx_AUTOCMD12_ERR_STATUS = 0x3c

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
)

// Configuration constants (p348, 35.4.2 Frequency divider configuration,
// IMX6FG) to support the following frequencies:
//   * Identification frequency ≤ 400 KHz
//   * Operating frequency ≤ 25 MHz
//   * High frequency ≤ 50 MHz
const (
	// p346, 35.2 Clocks, IMX6FG.
	//
	// The base clock is derived by default from PDF2 (396MHz) with divide
	// by 2, therefore 198MHz.

	// Data Timeout Counter Value: SDCLK x 2** 29
	DTOCV = 0xf

	// Divide-by-8
	DVS_ID = 8
	// Base clock divided by 64
	SDCLKFS_ID = 0x20
	// Identification frequency: 200 / (8 * 64) == ~400 KHz

	// Divide-by-2
	DVS_OP = 2
	// Base clock divided by 4
	SDCLKFS_OP = 0x02
	// Operating frequency: 200 / (2 * 4) == 25 MHz

	// Divide-by-1
	DVS_HS = 0
	// Base clock divided by 4 (Single Data Rate mode)
	SDCLKFS_HS_SDR = 0x02
	// Base clock divided by 4 (Dual Data Rate mode)
	SDCLKFS_HS_DDR = 0x01
	// High Speed frequency: 200 / (1 * 4) == 50 MHz

	// p35, Table 4, JESD84-B51
	//
	// Higher speed modes for eMMC cards are HS200 (controller supported,
	// currently unimplemented by this driver) and HS400 mode (unsupported
	// at controller level).
	//
	// p37-38, Figure 3-14 and 3-15, SD-PL-7.10
	//
	// Higher speed modes for SD cards are SDR50/SDR104/DDR50 (controller
	// supported, currently unimplemented by this driver) and UHS-II modes
	// (unsupported at controller level).
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

// USHDC represents a controller instance.
type USDHC struct {
	sync.Mutex

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

	// detected card properties
	card CardInfo

	readTimeout  time.Duration
	writeTimeout time.Duration
}

// USDHC1 instance
var USDHC1 = &USDHC{n: 1}

// USDHC2 instance
var USDHC2 = &USDHC{n: 2}

// p348, 35.4.2 Frequency divider configuration, IMX6FG
func (hw *USDHC) setClock(dvs int, sdclkfs int) {
	// wait for stable clock to comply with p4038, IMX6ULLRM DVS note
	reg.Wait(hw.pres_state, PRES_STATE_SDSTB, 1, 1)

	sys := reg.Read(hw.sys_ctrl)

	bits.SetN(&sys, SYS_CTRL_DVS, 0xf, uint32(dvs))
	bits.SetN(&sys, SYS_CTRL_SDCLKFS, 0xff, uint32(sdclkfs))

	reg.Write(hw.sys_ctrl, sys)
	reg.Wait(hw.pres_state, PRES_STATE_SDSTB, 1, 1)
}

// Detect performs voltage validation to detect an SD or MMC card.
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
		hw.cg = imx6.CCGR_CG1
	case 2:
		base = USDHC2_BASE
		hw.cg = imx6.CCGR_CG2
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

	// Generic SD specs read/write timeout rules (applied also to MMC by
	// this driver).
	//
	// p106, 4.6.2.1 Read, SD-PL-7.10
	hw.readTimeout = 100 * time.Millisecond
	// p106, 4.6.2.2 Write, SD-PL-7.10
	hw.writeTimeout = 500 * time.Millisecond

	hw.Unlock()
}

// Detect initializes an SD/MMC card as specified in
// p347, 35.4.1 Initializing the SD/MMC card, IMX6FG.
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
	hw.setClock(0, 0)
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

	err = hw.waitState(CURRENT_STATE_TRAN, 1*time.Millisecond)

	if err != nil {
		return
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

	bdAddress := dma.Alloc(bd.Bytes(), 0)
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

// ReadBlocks transfers full blocks of data from the card.
func (hw *USDHC) ReadBlocks(lba int, blocks int, buf []byte) (err error) {
	blockSize := hw.card.BlockSize
	bufSize := blocks * blockSize
	offset := uint64(lba) * uint64(blockSize)

	if bufSize == 0 {
		return
	}

	if len(buf) < bufSize {
		return errors.New("invalid buffer size")
	}

	hw.Lock()
	defer hw.Unlock()

	// CMD18 - READ_MULTIPLE_BLOCK - read consecutive blocks
	err = hw.transfer(18, READ, offset, uint32(blocks), uint32(blockSize), buf)

	return
}

// Read transfers data from the card.
func (hw *USDHC) Read(offset int64, size int64) (buf []byte, err error) {
	blockSize := int64(hw.card.BlockSize)

	if size == 0 {
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

// WriteBlocks transfers full blocks of data to the card.
func (hw *USDHC) WriteBlocks(lba int, buf []byte) (err error) {
	blockSize := hw.card.BlockSize
	offset := int64(lba) * int64(blockSize)

	return hw.Write(offset, buf)
}

// Write transfers data to the card.
func (hw *USDHC) Write(offset int64, buf []byte) (err error) {
	blockSize := uint32(hw.card.BlockSize)
	size := len(buf)

	if size == 0 {
		return
	}

	// TODO: support arbitrary write

	if uint32(offset)%blockSize != 0 {
		return fmt.Errorf("write offset must be %d bytes aligned", blockSize)
	}

	if uint32(size)%blockSize != 0 {
		return fmt.Errorf("write size must be %d bytes aligned", blockSize)
	}

	blocks := uint32(size) / blockSize

	hw.Lock()
	defer hw.Unlock()

	// CMD25 - WRITE_MULTIPLE_BLOCK - write consecutive blocks
	return hw.transfer(25, WRITE, uint64(offset), blocks, blockSize, buf)
}
