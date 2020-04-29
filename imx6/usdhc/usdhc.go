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
//
// +build tamago,arm

package usdhc

import (
	"errors"
	"sync"

	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/internal/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	// p4012, 58.8 uSDHC Memory Map/Register Definition, IMX6ULLRM
	HW_USDHC1_BASE uint32 = 0x02190000
	HW_USDHC2_BASE uint32 = 0x02194000

	HW_USDHCx_CMD_ARG = 0x08

	HW_USDHCx_CMD_XFR_TYP = 0x0c
	CMD_XFR_TYP_CMDINX    = 24
	CMD_XFR_TYP_DPSEL     = 21
	CMD_XFR_TYP_CICEN     = 20
	CMD_XFR_TYP_CCCEN     = 19
	CMD_XFR_TYP_RSPTYP    = 16

	HW_USDHCx_CMD_RSP0 = 0x10
	HW_USDHCx_CMD_RSP1 = 0x14
	HW_USDHCx_CMD_RSP2 = 0x18
	HW_USDHCx_CMD_RSP3 = 0x1c

	HW_USDHCx_PRES_STATE = 0x24
	PRES_STATE_SDSTB     = 3
	PRES_STATE_CDIHB     = 1
	PRES_STATE_CIHB      = 0

	HW_USDHCx_PROT_CTRL = 0x28
	PROT_CTRL_DMASEL    = 8
	PROT_CTRL_EMODE     = 4
	PROT_CTRL_DTW       = 1

	HW_USDHCx_SYS_CTRL = 0x2c
	SYS_CTRL_INITA     = 27
	SYS_CTRL_RSTD      = 26
	SYS_CTRL_RSTC      = 25
	SYS_CTRL_RSTA      = 24
	SYS_CTRL_DTOCV     = 16
	SYS_CTRL_SDCLKFS   = 8
	SYS_CTRL_DVS       = 4

	HW_USDHCx_INT_STATUS = 0x30
	INT_STATUS_DMAE      = 28
	INT_STATUS_TNE       = 26
	INT_STATUS_CIE       = 19
	INT_STATUS_CEBE      = 18
	INT_STATUS_CCE       = 17
	INT_STATUS_CTOE      = 16
	INT_STATUS_CC        = 0

	HW_USDHCx_INT_STATUS_EN = 0x34
	INT_STATUS_EN_DTOESEN   = 20

	HW_USDHCx_INT_SIGNAL_EN = 0x38

	HW_USDHCx_MIX_CTRL = 0x48
	MIX_CTRL_MSBSEL    = 5
	MIX_CTRL_DTDSEL    = 4
	MIX_CTRL_DDR_EN    = 3
	MIX_CTRL_AC12EN    = 2
	MIX_CTRL_BCEN      = 1
	MIX_CTRL_DMAEN     = 0
)

// p348, 35.4.2 Frequency divider configuration, IMX6FG
//   Identification frequency ≤ 400 KHz
//   Operating frequency ≤ 25 MHz
//   High frequency ≤ 50 MHz
const (
	// 200 MHz
	BASE_CLOCK = 200

	// Dual Data Rate
	DDR_ID = 0
	// Divide-by-8
	DVS_ID = 8
	// Base clock divided by 64
	SDCLKFS_ID = 0x20
	// identification frequency: 200 / (8 * 64) == ~400 KHz

	// Dual Data Rate
	DDR_OP = 0
	// Divide-by-2
	DVS_OP = 2
	// Base clock divided by 4
	SDCLKFS_OP = 0x02
	// operating frequency: 200 / (2 * 4) == 25 MHz

	// Dual Data Rate
	DDR_HS = 1
	// Divide-by-1
	DVS_HS = 1
	// Base clock divided by 4
	SDCLKFS_HS = 0x01
	// high speed frequency: 200 / (1 * 4) == 50 MHz

	// Data Timeout Counter Value: SDCLK x 2** 28
	DTOCV = 0b1110
)

type usdhc struct {
	sync.Mutex

	n             int
	width         int
	ddr           bool
	cg            int
	cmd_arg       uint32
	cmd_xfr       uint32
	cmd_rsp       uint32
	prot_ctrl     uint32
	sys_ctrl      uint32
	mix_ctrl      uint32
	pres_state    uint32
	int_status    uint32
	int_status_en uint32
	int_signal_en uint32

	mmc bool
	sd  bool
	hc  bool
}

var USDHC1 = &usdhc{n: 1}
var USDHC2 = &usdhc{n: 2}

// p348, 35.4.2 Frequency divider configuration, IMX6FG
func (hw *usdhc) setClock(dvs int, sdclkfs int) {
	// wait for stable clock to comply with p4038, IMX6ULLRM DVS note
	reg.Wait(hw.pres_state, PRES_STATE_SDSTB, 1, 1)

	ctrl := reg.Read(hw.sys_ctrl)

	bits.SetN(&ctrl, SYS_CTRL_DVS, 0xf, uint32(dvs))
	bits.SetN(&ctrl, SYS_CTRL_SDCLKFS, 0xff, uint32(sdclkfs))

	reg.Write(hw.sys_ctrl, ctrl)
	reg.Wait(hw.pres_state, PRES_STATE_SDSTB, 1, 1)
}

// Detect performs voltage validation to detect an SD or MMC card.
func (hw *usdhc) detect() (sd bool, mmc bool, hc bool, err error) {
	sd, hc = hw.voltageValidationSD()

	if sd {
		return
	}

	mmc, hc = hw.voltageValidationMMC()

	return
}

// Init initializes the uSDHC controller instance.
func (hw *usdhc) Init(width int, ddr bool) {
	var base uint32

	hw.Lock()

	switch hw.n {
	case 1:
		base = HW_USDHC1_BASE
		hw.cg = imx6.CCM_CCGR6_CG1
	case 2:
		base = HW_USDHC2_BASE
		hw.cg = imx6.CCM_CCGR6_CG2
	}

	hw.width = width
	hw.ddr = ddr
	hw.cmd_arg = base + HW_USDHCx_CMD_ARG
	hw.cmd_xfr = base + HW_USDHCx_CMD_XFR_TYP
	hw.cmd_rsp = base + HW_USDHCx_CMD_RSP0
	hw.prot_ctrl = base + HW_USDHCx_PROT_CTRL
	hw.sys_ctrl = base + HW_USDHCx_SYS_CTRL
	hw.mix_ctrl = base + HW_USDHCx_MIX_CTRL
	hw.pres_state = base + HW_USDHCx_PRES_STATE
	hw.int_status = base + HW_USDHCx_INT_STATUS
	hw.int_status_en = base + HW_USDHCx_INT_STATUS_EN
	hw.int_signal_en = base + HW_USDHCx_INT_SIGNAL_EN

	hw.Unlock()
}

// Detect initializes an SD/MMC card as specified in
// p347, 35.4.1 Initializing the SD/MMC card, IMX6FG.
func (hw *usdhc) Detect() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.cg == 0 {
		return errors.New("controller is uninitialized")
	}

	// enable clock
	reg.SetN(imx6.CCM_CCGR6, hw.cg, 0b11, 0b11)

	// soft reset uSDHC
	reg.Set(hw.sys_ctrl, SYS_CTRL_RSTA)
	reg.Wait(hw.sys_ctrl, SYS_CTRL_RSTA, 1, 0)

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
	// set endianness (little), (TODO: expose via API?)
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
	err = hw.cmd(0, READ, GO_IDLE_STATE, RSP_NONE, false, false)

	if err != nil {
		return
	}

	hw.sd, hw.mmc, hw.hc, err = hw.detect()

	if err != nil {
		return
	}

	if hw.sd {
		err = hw.initSD()
	} else if hw.mmc {
		err = hw.initMMC()
	} else {
		err = errors.New("no SD/MMC card detected")
	}

	return
}
