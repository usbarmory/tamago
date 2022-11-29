// NXP i.MX6UL clock control
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6ul

import (
	"errors"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// Clock registers
const (
	CCM_CACRR      = 0x020c4010
	CACRR_ARM_PODF = 0

	CCM_CBCDR      = 0x020c4014
	CBCDR_IPG_PODF = 8

	CCM_CSCDR1           = 0x020c4024
	CSCDR1_USDHC2_PODF   = 16
	CSCDR1_USDHC1_PODF   = 11
	CSCDR1_UART_CLK_SEL  = 6
	CSCDR1_UART_CLK_PODF = 0

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

	CCM_ANALOG_PLL_USB1 = CCM_ANALOG_PLL_ARM + 0x10
	CCM_ANALOG_PLL_USB2 = CCM_ANALOG_PLL_ARM + 0x20
	PLL_EN_USB_CLKS     = 6

	CCM_ANALOG_PLL_ENET  = CCM_ANALOG_PLL_ARM + 0xe0
	PLL_ENET2_125M_EN    = 20
	PLL_ENET1_125M_EN    = 13
	PLL_ENET1_DIV_SELECT = 2
	PLL_ENET0_DIV_SELECT = 0

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

	CCM_CCGR0 = 0x020c4068
	CCM_CCGR1 = 0x020c406c
	CCM_CCGR2 = 0x020c4070
	CCM_CCGR3 = 0x020c4074
	CCM_CCGR5 = 0x020c407c
	CCM_CCGR6 = 0x020c4080

	CCGRx_CG15 = 30
	CCGRx_CG14 = 28
	CCGRx_CG13 = 26
	CCGRx_CG12 = 24
	CCGRx_CG11 = 22
	CCGRx_CG10 = 20
	CCGRx_CG9  = 18
	CCGRx_CG8  = 16
	CCGRx_CG7  = 14
	CCGRx_CG6  = 12
	CCGRx_CG5  = 10
	CCGRx_CG4  = 8
	CCGRx_CG3  = 6
	CCGRx_CG2  = 4
	CCGRx_CG1  = 2
	CCGRx_CG0  = 0
)

const (
	IOMUXC_GPR_GPR1  = 0x020e4004
	ENET2_TX_CLK_DIR = 18
	ENET1_TX_CLK_DIR = 17
	ENET2_CLK_SEL    = 14
	ENET1_CLK_SEL    = 13
)

// Oscillator frequencies
const (
	OSC_FREQ  = 24000000
	PLL2_FREQ = 528000000
	PLL3_FREQ = 480000000
)

// Operating ARM core frequencies in MHz (care must be taken as not all P/Ns
// support the entire range)
// (p24, Table 10. Operating Ranges, IMX6ULLCEC).
const (
	FreqMax = Freq900
	Freq900 = 900
	Freq792 = 792
	Freq528 = 528
	Freq396 = 396
	Freq198 = 198
	FreqLow = Freq198
)

// Clocks at boot time
// (p261, Table 8-4. Normal frequency clocks configuration, IMX6ULLRM)
const (
	IPG_FREQ = 66000000
	AHB_FREQ = 132000000
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

func setOperatingPoint(uV uint32) {
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

// SetARMFreq changes the ARM core frequency, see `Freq*` constants for
// supported values. This function allows overclocking as it does not verify
// P/N compatibility with the desired frequency.
func SetARMFreq(mhz uint32) (err error) {
	var div_select uint32
	var arm_podf uint32
	var uV uint32

	curMHz := ARMFreq() / 1000000

	if mhz == curMHz {
		return
	}

	// p24, Table 10. Operating Ranges, IMX6ULLCEC
	switch mhz {
	case Freq900:
		div_select = 75
		arm_podf = 0
		uV = 1275000
	case Freq792:
		div_select = 66
		arm_podf = 0
		uV = 1225000
	case Freq528:
		div_select = 88
		arm_podf = 1
		uV = 1175000
	case Freq396:
		div_select = 66
		arm_podf = 1
		uV = 1025000
	case Freq198:
		div_select = 66
		arm_podf = 3
		uV = 950000
	default:
		return errors.New("unsupported")
	}

	if mhz > curMHz {
		setOperatingPoint(uV)
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

	if mhz < curMHz {
		setOperatingPoint(uV)
	}

	return
}

// GetPeripheralClock returns the IPG_CLK_ROOT frequency,
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM).
func GetPeripheralClock() uint32 {
	// IPG_CLK_ROOT derived from AHB_CLK_ROOT which is 132 MHz
	ipg_podf := reg.Get(CCM_CBCDR, CBCDR_IPG_PODF, 0b11)
	return AHB_FREQ / (ipg_podf + 1)
}

// GetHighFrequencyClock returns the PERCLK_CLK_ROOT frequency,
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM).
func GetHighFrequencyClock() uint32 {
	var freq uint32

	if reg.Get(CCM_CSCMR1, CSCMR1_PERCLK_SEL, 1) == 1 {
		freq = OSC_FREQ
	} else {
		freq = GetPeripheralClock()
	}

	podf := reg.Get(CCM_CSCMR1, CSCMR1_PERCLK_PODF, 0x3f)

	return freq / (podf + 1)
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

	if reg.Get(register, gate_pos, 1) == 1 {
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

// GetUARTClock returns the UART_CLK_ROOT frequency,
// (p630, Figure 18-3. Clock Tree - Part 2, IMX6ULLRM).
func GetUARTClock() uint32 {
	var freq uint32

	if reg.Get(CCM_CSCDR1, CSCDR1_UART_CLK_SEL, 1) == 1 {
		freq = OSC_FREQ
	} else {
		// match /6 static divider (p630, Figure 18-3. Clock Tree - Part 2, IMX6ULLRM)
		freq = PLL3_FREQ / 6
	}

	podf := reg.Get(CCM_CSCDR1, CSCDR1_UART_CLK_PODF, 0b111111)

	return freq / (podf + 1)
}

// GetUSDHCClock returns the USDHCx_CLK_ROOT clock by reading CSCMR1[USDHCx_CLK_SEL]
// and CSCDR1[USDHCx_PODF]
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM)
func GetUSDHCClock(index int) (podf uint32, clksel uint32, clock uint32) {
	var podf_pos int
	var clksel_pos int
	var freq uint32

	switch index {
	case 1:
		podf_pos = CSCDR1_USDHC1_PODF
		clksel_pos = CSCMR1_USDHC1_CLK_SEL
	case 2:
		podf_pos = CSCDR1_USDHC2_PODF
		clksel_pos = CSCMR1_USDHC2_CLK_SEL
	default:
		return
	}

	podf = reg.Get(CCM_CSCDR1, podf_pos, 0b111)
	clksel = reg.Get(CCM_CSCMR1, clksel_pos, 1)

	if clksel == 1 {
		_, freq = GetPFD(2, 0)
	} else {
		_, freq = GetPFD(2, 2)
	}

	clock = freq / (podf + 1)

	return
}

// SetUSDHCClock controls the USDHCx_CLK_ROOT clock by setting CSCMR1[USDHCx_CLK_SEL]
// and CSCDR1[USDHCx_PODF]
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM).
func SetUSDHCClock(index int, podf uint32, clksel uint32) (err error) {
	var podf_pos int
	var clksel_pos int

	if podf < 0 || podf > 7 {
		return errors.New("podf value out of range")
	}

	if clksel < 0 || clksel > 1 {
		return errors.New("selector value out of range")
	}

	switch index {
	case 1:
		podf_pos = CSCDR1_USDHC1_PODF
		clksel_pos = CSCMR1_USDHC1_CLK_SEL
	case 2:
		podf_pos = CSCDR1_USDHC2_PODF
		clksel_pos = CSCMR1_USDHC2_CLK_SEL
	default:
		return errors.New("invalid interface index")
	}

	reg.SetN(CCM_CSCDR1, podf_pos, 0b111, podf)
	reg.SetN(CCM_CSCMR1, clksel_pos, 1, clksel)

	return
}

// EnableUSBPLL enables the USBPHY0 480MHz PLL.
func EnableUSBPLL(index int) (err error) {
	var pll uint32

	switch index {
	case 1:
		pll = CCM_ANALOG_PLL_USB1
	case 2:
		pll = CCM_ANALOG_PLL_USB2
	default:
		return errors.New("invalid interface index")
	}

	// power up PLL
	reg.Set(pll, PLL_POWER)
	reg.Set(pll, PLL_EN_USB_CLKS)

	// wait for lock
	reg.Wait(pll, PLL_LOCK, 1, 1)

	// remove bypass
	reg.Clear(pll, PLL_BYPASS)

	// enable PLL
	reg.Set(pll, PLL_ENABLE)

	return
}

// EnableENETPLL enables the Ethernet MAC 50MHz PLL.
func EnableENETPLL(index int) (err error) {
	var sel int
	var dir int
	var enable int
	var div_select int
	var pll uint32 = CCM_ANALOG_PLL_ENET

	switch index {
	case 1:
		sel = ENET1_CLK_SEL
		dir = ENET1_TX_CLK_DIR
		enable = PLL_ENET1_125M_EN
		div_select = PLL_ENET0_DIV_SELECT
	case 2:
		sel = ENET2_CLK_SEL
		dir = ENET2_TX_CLK_DIR
		enable = PLL_ENET2_125M_EN
		div_select = PLL_ENET1_DIV_SELECT
	default:
		return errors.New("invalid interface index")
	}

	// set reference clock
	reg.Clear(IOMUXC_GPR_GPR1, sel)
	reg.Set(IOMUXC_GPR_GPR1, dir)

	// set frequency to 50MHz
	reg.SetN(pll, div_select, 0b11, 1)

	// power up PLL
	reg.Clear(pll, PLL_POWER)

	// wait for lock
	reg.Wait(pll, PLL_LOCK, 1, 1)

	// enable PLL
	reg.Set(pll, enable)

	// remove bypass
	reg.Clear(pll, PLL_BYPASS)

	return
}
