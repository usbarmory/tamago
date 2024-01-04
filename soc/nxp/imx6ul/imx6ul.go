// NXP i.MX6UL configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package imx6ul provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the NXP i.MX6UL family of System-on-Chip (SoC)
// application processors.
//
// The package implements initialization and drivers for NXP
// i.MX6UL/i.MX6ULL/i.MX6ULZ SoCs, adopting the following reference
// specifications:
//   - IMX6ULCEC  - i.MX6UL  Data Sheet                               - Rev 2.2 2015/05
//   - IMX6ULLCEC - i.MX6ULL Data Sheet                               - Rev 1.2 2017/11
//   - IMX6ULZCEC - i.MX6ULZ Data Sheet                               - Rev 0   2018/09
//   - IMX6ULRM   - i.MX 6UL  Applications Processor Reference Manual - Rev 1   2016/04
//   - IMX6ULLRM  - i.MX 6ULL Applications Processor Reference Manual - Rev 1   2017/11
//   - IMX6ULZRM  - i.MX 6ULZ Applications Processor Reference Manual - Rev 0   2018/10
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package imx6ul

import (
	"encoding/binary"

	"github.com/usbarmory/tamago/internal/reg"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/arm/gic"
	"github.com/usbarmory/tamago/arm/tzc380"

	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/caam"
	"github.com/usbarmory/tamago/soc/nxp/csu"
	"github.com/usbarmory/tamago/soc/nxp/dcp"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/gpio"
	"github.com/usbarmory/tamago/soc/nxp/i2c"
	"github.com/usbarmory/tamago/soc/nxp/ocotp"
	"github.com/usbarmory/tamago/soc/nxp/rngb"
	"github.com/usbarmory/tamago/soc/nxp/snvs"
	"github.com/usbarmory/tamago/soc/nxp/tempmon"
	"github.com/usbarmory/tamago/soc/nxp/uart"
	"github.com/usbarmory/tamago/soc/nxp/usb"
	"github.com/usbarmory/tamago/soc/nxp/usdhc"
	"github.com/usbarmory/tamago/soc/nxp/wdog"
)

// Peripheral registers
const (
	// Bus Encryption Engine (UL only)
	BEE_BASE = 0x02044000

	// Cryptographic Acceleration and Assurance Module (UL only)
	CAAM_BASE = 0x02140000

	// Central Security Unit
	CSU_BASE = 0x021c0000

	// Data Co-Processor (ULL/ULZ only)
	DCP_BASE = 0x02280000

	// General Interrupt Controller
	GIC_BASE = 0x00a00000

	// General Purpose I/O
	GPIO1_BASE = 0x0209c000
	GPIO2_BASE = 0x020a0000
	GPIO3_BASE = 0x020a4000
	GPIO4_BASE = 0x020a8000
	GPIO5_BASE = 0x020ac000

	// Ethernet MAC (UL/ULL only)
	ENET1_BASE = 0x02188000
	ENET2_BASE = 0x020b4000

	// Ethernet MAC interrupts
	ENET1_IRQ  = 32 + 118
	ENET2_IRQ  = 32 + 120

	// I2C
	I2C1_BASE = 0x021a0000
	I2C2_BASE = 0x021a4000

	// Multi Mode DDR Controller
	MMDC_BASE = 0x80000000

	// On-Chip OTP Controller
	OCOTP_BASE      = 0x021bc000
	OCOTP_BANK_BASE = 0x021bc400

	// On-Chip Random-Access Memory
	OCRAM_START = 0x00900000
	OCRAM_SIZE  = 0x20000

	// True Random Number Generator (ULL/ULZ only)
	RNGB_BASE = 0x02284000

	// Secure Non-Volatile Storage
	SNVS_HP_BASE = 0x020cc000
	SNVS_LP_BASE = 0x020b0000

	// Temperature Monitor
	TEMPMON_BASE = 0x020c8180

	// TrustZone Address Space Controller
	TZASC_BASE            = 0x021d0000
	TZASC_BYPASS          = 0x020e4024
	GPR1_TZASC1_BOOT_LOCK = 23

	// Serial ports
	UART1_BASE = 0x02020000
	UART2_BASE = 0x021e8000
	UART3_BASE = 0x021ec000
	UART4_BASE = 0x021f0000

	// USB 2.0 controller
	USB_ANALOG1_BASE   = 0x020c81a0
	USB_ANALOG2_BASE   = 0x020c8200
	USB_ANALOG_DIGPROG = 0x020c8260
	USBPHY1_BASE       = 0x020c9000
	USBPHY2_BASE       = 0x020ca000
	USB1_BASE          = 0x02184000
	USB2_BASE          = 0x02184200

	// USB 2.0 controller interrupts
	USB1_IRQ           = 32 + 43
	USB2_IRQ           = 32 + 42

	// SD/MMC
	USDHC1_BASE = 0x02190000
	USDHC2_BASE = 0x02194000

	// Watchdog Timers
	WDOG1_BASE = 0x020bc000
	WDOG2_BASE = 0x020c0000
	WDOG3_BASE = 0x021e4000

	// Watchdog Timer interrupts
	WDOG1_IRQ  = 32 + 80
	WDOG2_IRQ  = 32 + 81
	WDOG3_IRQ  = 32 + 11
)

// Peripheral instances
var (
	// ARM core
	ARM = &arm.CPU{}

	// Bus Encryption Engine (UL only)
	BEE *bee.BEE

	// Cryptographic Acceleration and Assurance Module (UL only)
	CAAM *caam.CAAM

	// Central Security Unit
	CSU = &csu.CSU{
		Base: CSU_BASE,
		CCGR: CCM_CCGR1,
		CG:   CCGRx_CG14,
	}

	// Data Co-Processor (ULL/ULZ only)
	DCP *dcp.DCP

	// Generic Interrupt Controller
	GIC = &gic.GIC{
		Base: GIC_BASE,
	}

	// GPIO controller 1
	GPIO1 = &gpio.GPIO{
		Index: 1,
		Base:  GPIO1_BASE,
		CCGR:  CCM_CCGR1,
		CG:    CCGRx_CG13,
	}

	// GPIO controller 2
	GPIO2 = &gpio.GPIO{
		Index: 2,
		Base:  GPIO2_BASE,
		CCGR:  CCM_CCGR0,
		CG:    CCGRx_CG15,
	}

	// GPIO controller 3
	GPIO3 = &gpio.GPIO{
		Index: 3,
		Base:  GPIO3_BASE,
		CCGR:  CCM_CCGR2,
		CG:    CCGRx_CG13,
	}

	// GPIO controller 4
	GPIO4 = &gpio.GPIO{
		Index: 4,
		Base:  GPIO4_BASE,
		CCGR:  CCM_CCGR3,
		CG:    CCGRx_CG6,
	}

	// GPIO controller 5
	GPIO5 = &gpio.GPIO{
		Index: 5,
		Base:  GPIO5_BASE,
		CCGR:  CCM_CCGR1,
		CG:    CCGRx_CG15,
	}

	// Ethernet MAC 1 (UL/ULL only)
	ENET1 *enet.ENET
	ENET2 *enet.ENET

	// I2C controller 1
	I2C1 = &i2c.I2C{
		Index: 1,
		Base:  I2C1_BASE,
		CCGR:  CCM_CCGR2,
		CG:    CCGRx_CG3,
	}

	// I2C controller 2
	I2C2 = &i2c.I2C{
		Index: 2,
		Base:  I2C2_BASE,
		CCGR:  CCM_CCGR2,
		CG:    CCGRx_CG5,
	}

	// On-Chip OTP Controller
	OCOTP = &ocotp.OCOTP{
		Base:     OCOTP_BASE,
		BankBase: OCOTP_BANK_BASE,
		CCGR:     CCM_CCGR2,
		CG:       CCGRx_CG6,
	}

	// True Random Number Generator (ULL/ULZ only)
	RNGB *rngb.RNGB

	// Secure Non-Volatile Storage
	SNVS = &snvs.SNVS{
		Base: SNVS_HP_BASE,
		CCGR: CCM_CCGR5,
		CG:   CCGRx_CG9,
	}

	// Temperature Monitor
	TEMPMON = &tempmon.TEMPMON{
		Base: TEMPMON_BASE,
	}

	// TrustZone Address Space Controller
	TZASC = &tzc380.TZASC{
		Base:              TZASC_BASE,
		Bypass:            TZASC_BYPASS,
		SecureBootLockReg: IOMUXC_GPR_GPR1,
		SecureBootLockPos: GPR1_TZASC1_BOOT_LOCK,
	}

	// Serial port 1
	UART1 = &uart.UART{
		Index: 1,
		Base:  UART1_BASE,
		CCGR:  CCM_CCGR5,
		CG:    CCGRx_CG12,
		Clock: GetUARTClock,
	}

	// Serial port 2
	UART2 = &uart.UART{
		Index: 2,
		Base:  UART2_BASE,
		CCGR:  CCM_CCGR0,
		CG:    CCGRx_CG14,
		Clock: GetUARTClock,
	}

	// USB controller 1
	USB1 = &usb.USB{
		Index:     1,
		Base:      USB1_BASE,
		CCGR:      CCM_CCGR6,
		CG:        CCGRx_CG0,
		Analog:    USB_ANALOG1_BASE,
		PHY:       USBPHY1_BASE,
		IRQ:       USB1_IRQ,
		EnablePLL: EnableUSBPLL,
	}

	// USB controller 2
	USB2 = &usb.USB{
		Index:     2,
		Base:      USB2_BASE,
		CCGR:      CCM_CCGR6,
		CG:        CCGRx_CG0,
		Analog:    USB_ANALOG2_BASE,
		PHY:       USBPHY2_BASE,
		IRQ:       USB2_IRQ,
		EnablePLL: EnableUSBPLL,
	}

	// SD/MMC controller 1
	USDHC1 = &usdhc.USDHC{
		Index:    1,
		Base:     USDHC1_BASE,
		CCGR:     CCM_CCGR6,
		CG:       CCGRx_CG1,
		SetClock: SetUSDHCClock,
	}

	// SD/MMC controller 2
	USDHC2 = &usdhc.USDHC{
		Index:    2,
		Base:     USDHC2_BASE,
		CCGR:     CCM_CCGR6,
		CG:       CCGRx_CG2,
		SetClock: SetUSDHCClock,
	}

	// Watchdog Timer 1
	WDOG1 = &wdog.WDOG{
		Index: 1,
		Base:  WDOG1_BASE,
		CCGR:  CCM_CCGR3,
		CG:    CCGRx_CG8,
		IRQ:   WDOG1_IRQ,
	}

	// Watchdog Timer 2
	WDOG2 = &wdog.WDOG{
		Index: 2,
		Base:  WDOG2_BASE,
		CCGR:  CCM_CCGR5,
		CG:    CCGRx_CG5,
		IRQ:   WDOG2_IRQ,
	}

	// TrustZone Watchdog
	TZ_WDOG = WDOG2

	// Watchdog Timer 3
	WDOG3 = &wdog.WDOG{
		Index: 3,
		Base:  WDOG3_BASE,
		CCGR:  CCM_CCGR6,
		CG:    CCGRx_CG10,
		IRQ:   WDOG3_IRQ,
	}
)

// SiliconVersion returns the SoC silicon version information
// (p3945, 57.4.11 Chip Silicon Version (USB_ANALOG_DIGPROG), IMX6ULLRM).
func SiliconVersion() (sv, family, revMajor, revMinor uint32) {
	sv = reg.Read(USB_ANALOG_DIGPROG)

	family = (sv >> 16) & 0xff
	revMajor = (sv >> 8) & 0xff
	revMinor = sv & 0xff

	return
}

// UniqueID returns the NXP SoC Device Unique 64-bit ID.
func UniqueID() (uid [8]byte) {
	cfg0, _ := OCOTP.Read(0, 1)
	cfg1, _ := OCOTP.Read(0, 2)

	binary.LittleEndian.PutUint32(uid[0:4], cfg0)
	binary.LittleEndian.PutUint32(uid[4:8], cfg1)

	return
}

// Model returns the SoC model name.
func Model() (model string) {
	switch Family {
	case IMX6UL:
		model = "i.MX6UL"
	case IMX6ULL:
		cfg5, _ := OCOTP.Read(0, 6)

		if (cfg5>>6)&1 == 1 {
			model = "i.MX6ULZ"
		} else {
			model = "i.MX6ULL"
		}
	default:
		model = "unknown"
	}

	return
}
