// NXP i.MX6UL ARM clock control
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"errors"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// Clock registers
const (
	CCM_CACRR      = 0x020c4010
	CACRR_ARM_PODF = 0

	CCM_CBCDR      = 0x020c4014
	CBCDR_IPG_PODF = 8

	CCM_CSCDR1             = 0x020c4024
	CSCDR1_USDHC2_CLK_PODF = 16
	CSCDR1_USDHC1_CLK_PODF = 11
	CSCDR1_UART_CLK_SEL    = 6
	CSCDR1_UART_CLK_PODF   = 0

	CCM_CSCMR1            = 0x020c401c
	CSCMR1_USDHC2_CLK_SEL = 17
	CSCMR1_USDHC1_CLK_SEL = 16
	CSCMR1_PERCLK_SEL     = 6
	CSCMR1_PERCLK_PODF    = 0

	CCM_ANALOG_PLL_ARM = 0x020c8000
	PLL_LOCK           = 31
	PLL_BYPASS         = 16
	PLL_BYPASS_CLK_SRC = 14
	PLL_ENABLE         = 13
	PLL_POWER          = 12
	PLL_DIV_SELECT     = 0

	CCM_ANALOG_PFD_480  = 0x020c80f0
	CCM_ANALOG_PFD_528  = 0x020c8100
	ANALOG_PFD3_CLKGATE = 31
	ANALOG_PFD3_FRAC    = 24
	ANALOG_PFD2_CLKGATE = 23
	ANALOG_PFD2_FRAC    = 16
	ANALOG_PFD1_CLKGATE = 15
	ANALOG_PFD1_FRAC    = 8
	ANALOG_PFD0_CLKGATE = 7
	ANALOG_PFD0_FRAC    = 0

	PMU_REG_CORE   = 0x020c8140
	CORE_REG2_TARG = 18
	CORE_REG0_TARG = 0

	CCM_CCGR2 = 0x020c4070
	CCGR2_CG5 = 10
	CCGR2_CG4 = 8
	CCGR2_CG3 = 6

	CCM_CCGR6  = 0x020c4080
	CCGR6_CG12 = 24
	CCGR6_CG2  = 4
	CCGR6_CG1  = 2
	CCGR6_CG0  = 0
)

// Oscillator frequencies
const (
	OSC_FREQ  = 24000000
	PLL2_FREQ = 528000000
	PLL3_FREQ = 480000000
)

// ARMCoreDiv returns the ARM core divider value
// (p665, 18.6.5 CCM Arm Clock Root Register, IMX6ULLRM).
func ARMCoreDiv() (div float32) {
	return float32(reg.Get(CCM_CACRR, CACRR_ARM_PODF, 0b111) + 1)
}

// ARMPLLDiv returns the ARM PLL divider value
// (p714, 18.7.1 Analog ARM PLL control Register, IMX6ULLRM).
func ARMPLLDiv() (div float32) {
	return float32(reg.Get(CCM_ANALOG_PLL_ARM, PLL_DIV_SELECT, 0b1111111)) / 2
}

// ARMFreq returns the ARM core frequency.
func ARMFreq() (hz uint32) {
	// (OSC_FREQ * (DIV_SELECT / 2)) / (ARM_PODF + 1)
	return uint32((OSC_FREQ * ARMPLLDiv()) / ARMCoreDiv())
}

func setOperatingPointIMX6ULL(uV uint32) {
	var reg0Targ uint32
	var reg2Targ uint32

	curTarg := reg.Get(PMU_REG_CORE, CORE_REG0_TARG, 0b11111)

	// p2456, 39.6.4 Digital Regulator Core Register, IMX6ULLRM
	if uV < 725000 {
		reg0Targ = 0b00000
	} else if uV > 1450000 {
		reg0Targ = 0b11111
	} else {
		reg0Targ = (uV - 700000) / 25000
	}

	if reg0Targ == curTarg {
		return
	}

	// VDD_SOC_CAP Min is 1150000 (targ == 18)
	if reg0Targ < 18 {
		reg2Targ = 18
	} else {
		reg2Targ = reg0Targ
	}

	r := reg.Read(PMU_REG_CORE)

	// set ARM core target voltage
	bits.SetN(&r, CORE_REG0_TARG, 0b11111, reg0Targ)
	// set SOC target voltage
	bits.SetN(&r, CORE_REG2_TARG, 0b11111, reg2Targ)

	reg.Write(PMU_REG_CORE, r)
	arm.Busyloop(10000)
}

func setARMFreqIMX6ULL(hz uint32) (err error) {
	var div_select uint32
	var arm_podf uint32
	var uV uint32

	curHz := ARMFreq()

	if hz == curHz {
		return
	}

	// p24, Table 10. Operating Ranges, IMX6ULLCEC
	switch hz {
	case 900000000:
		div_select = 75
		arm_podf = 0
		uV = 1275000
	case 792000000:
		div_select = 66
		arm_podf = 0
		uV = 1225000
	case 528000000:
		div_select = 88
		arm_podf = 1
		uV = 1175000
	case 396000000:
		div_select = 66
		arm_podf = 1
		uV = 1025000
	case 198000000:
		div_select = 66
		arm_podf = 3
		uV = 950000
	default:
		return errors.New("unsupported")
	}

	if hz > curHz {
		setOperatingPointIMX6ULL(uV)
	}

	// set bypass source to main oscillator
	reg.SetN(CCM_ANALOG_PLL_ARM, PLL_BYPASS_CLK_SRC, 0b11, 0)

	// bypass
	reg.Set(CCM_ANALOG_PLL_ARM, PLL_BYPASS)

	// set PLL divisor
	reg.SetN(CCM_ANALOG_PLL_ARM, PLL_DIV_SELECT, 0b1111111, div_select)

	// wait for lock
	reg.Wait(CCM_ANALOG_PLL_ARM, PLL_LOCK, 1, 1)

	// remove bypass
	reg.Clear(CCM_ANALOG_PLL_ARM, PLL_BYPASS)

	// set core divisor
	reg.SetN(CCM_CACRR, CACRR_ARM_PODF, 0b111, arm_podf)

	if hz < curHz {
		setOperatingPointIMX6ULL(uV)
	}

	return
}

// SetARMFreq changes the ARM core frequency to the desired setting (in MHz).
func SetARMFreq(mhz uint32) (err error) {
	switch Family {
	case IMX6ULL:
		err = setARMFreqIMX6ULL(mhz * 1000000)
	default:
		err = errors.New("unsupported")
	}

	return
}

// GetPFD returns the fractional divider and frequency in Hz of a PLL PFD
// (p734, 18.7.15 480MHz Clock (PLL3) Phase Fractional Divider Control Register, IMX6ULLRM)
// (p736, 18.7.16 480MHz Clock (PLL2) Phase Fractional Divider Control Register, IMX6ULLRM).
func GetPFD(pll int, pfd int) (div uint32, hz uint32) {
	var register uint32
	var div_pos, gate_pos int
	var freq float64

	switch pll {
	case 2:
		register = CCM_ANALOG_PFD_528
		freq = PLL2_FREQ
	case 3:
		register = CCM_ANALOG_PFD_480
		freq = PLL3_FREQ
	default:
		// Only PLL2 and PLL3 have PFD's.
		return
	}

	switch pfd {
	case 0:
		gate_pos = ANALOG_PFD0_CLKGATE
		div_pos = ANALOG_PFD0_FRAC
	case 1:
		gate_pos = ANALOG_PFD1_CLKGATE
		div_pos = ANALOG_PFD1_FRAC
	case 2:
		gate_pos = ANALOG_PFD2_CLKGATE
		div_pos = ANALOG_PFD2_FRAC
	case 3:
		gate_pos = ANALOG_PFD3_CLKGATE
		div_pos = ANALOG_PFD3_FRAC
	default:
		return
	}

	if reg.Get(register, gate_pos, 0b1) == 1 {
		return
	}

	// Output frequency has a static multiplicator of 18
	// p646, 18.5.1.4 Phase Fractional Dividers (PFD)
	div = reg.Get(register, div_pos, 0b111111)
	hz = uint32((freq * 18) / float64(div))

	return
}

// SetPFD sets the fractional divider of a PPL PFD
// (p734, 18.7.15 480MHz Clock (PLL3) Phase Fractional Divider Control Register, IMX6ULLRM)
// (p736, 18.7.16 480MHz Clock (PLL2) Phase Fractional Divider Control Register, IMX6ULLRM).
func SetPFD(pll uint32, pfd uint32, div uint32) error {
	var register uint32
	var div_pos int

	switch pll {
	case 2:
		register = CCM_ANALOG_PFD_528
	case 3:
		register = CCM_ANALOG_PFD_480
	default:
		return errors.New("invalid pll index")
	}

	// Divider can range from 12 to 35
	// p646, 18.5.1.4 Phase Fractional Dividers (PFD), IMX6ULLRM.
	if div < 12 || div > 35 {
		return errors.New("invalid div value")
	}

	switch pfd {
	case 0:
		div_pos = ANALOG_PFD0_FRAC
	case 1:
		div_pos = ANALOG_PFD1_FRAC
	case 2:
		div_pos = ANALOG_PFD2_FRAC
	case 3:
		div_pos = ANALOG_PFD3_FRAC
	default:
		return errors.New("invalid pfd index")
	}

	reg.SetN(register, div_pos, 0b111111, div)

	return nil
}
