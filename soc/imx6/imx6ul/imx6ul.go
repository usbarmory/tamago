// NXP i.MX6UL initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package imx6ul provides hardware initialization, automatically on import,
// for the i.MX6UL family of System-on-Chip components.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package imx6ul

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/csu"
	"github.com/usbarmory/tamago/soc/imx6/dcp"
	"github.com/usbarmory/tamago/soc/imx6/i2c"
	"github.com/usbarmory/tamago/soc/imx6/ocotp"
	"github.com/usbarmory/tamago/soc/imx6/rngb"
	"github.com/usbarmory/tamago/soc/imx6/tzasc"
	"github.com/usbarmory/tamago/soc/imx6/uart"
	"github.com/usbarmory/tamago/soc/imx6/usb"
	"github.com/usbarmory/tamago/soc/imx6/usdhc"
)

// Peripheral registers
const (
	CSU_BASE = 0x021c0000

	DCP_BASE = 0x02280000

	I2C1_BASE = 0x021a0000
	I2C2_BASE = 0x021a4000

	OCOTP_BASE      = 0x021bc000
	OCOTP_BANK_BASE = 0x021bc400

	RNGB_BASE = 0x02284000

	TZASC_BASE   = 0x021d0000
	TZASC_BYPASS = 0x020e4024

	UART1_BASE = 0x02020000
	UART2_BASE = 0x021e8000
	UART3_BASE = 0x021ec000
	UART4_BASE = 0x021f0000

	USB_ANALOG1_BASE = 0x020c81a0
	USB_ANALOG2_BASE = 0x020c8200
	USBPHY1_BASE     = 0x020c9000
	USBPHY2_BASE     = 0x020ca000
	USB1_BASE        = 0x02184000
	USB2_BASE        = 0x02184200

	USDHC1_BASE = 0x02190000
	USDHC2_BASE = 0x02194000
)

// Peripheral instances
var (
	// Central Security Unit
	CSU = &csu.CSU{
		Base: CSU_BASE,
		CCGR: imx6.CCM_CCGR1,
		CG:   imx6.CCGRx_CG14,
	}

	// Data Co-Processor
	DCP = &dcp.DCP{
		Base:            DCP_BASE,
		DeriveKeyMemory: dma.Default(),
	}

	// I2C controller 1
	I2C1 = &i2c.I2C{
		Index: 1,
		Base:  I2C1_BASE,
		CCGR:  imx6.CCM_CCGR2,
		CG:    imx6.CCGRx_CG3,
	}

	// I2C controller 2
	I2C2 = &i2c.I2C{
		Index: 2,
		Base:  I2C2_BASE,
		CCGR:  imx6.CCM_CCGR2,
		CG:    imx6.CCGRx_CG5,
	}

	// On-Chip OTP Controller
	OCOTP = &ocotp.OCOTP{
		Base:     OCOTP_BASE,
		BankBase: OCOTP_BANK_BASE,
		CCGR:     imx6.CCM_CCGR2,
		CG:       imx6.CCGRx_CG6,
	}

	// Random Number Generator
	RNGB = &rngb.RNGB{
		Base: RNGB_BASE,
	}

	// TrustZone Address Space Controller
	TZASC = &tzasc.TZASC{
		Base:              TZASC_BASE,
		Bypass:            TZASC_BYPASS,
		SecureBootLockReg: imx6.IOMUXC_GPR_GPR1,
		SecureBootLockPos: imx6.GPR1_TZASC1_BOOT_LOCK,
	}

	// Serial port 1
	UART1 = &uart.UART{
		Index: 1,
		Base:  UART1_BASE,
	}

	// Serial port 2
	UART2 = &uart.UART{
		Index: 2,
		Base:  UART2_BASE,
	}

	// USB controller 1
	USB1 = &usb.USB{
		Index:  1,
		Base:   USB1_BASE,
		CCGR:   imx6.CCM_CCGR6,
		CG:     imx6.CCGRx_CG0,
		Analog: USB_ANALOG1_BASE,
		PHY:    USBPHY1_BASE,
		PLL:    imx6.CCM_ANALOG_PLL_USB1,
	}

	// USB controller 2
	USB2 = &usb.USB{
		Index:  2,
		Base:   USB2_BASE,
		CCGR:   imx6.CCM_CCGR6,
		CG:     imx6.CCGRx_CG0,
		Analog: USB_ANALOG2_BASE,
		PHY:    USBPHY2_BASE,
		PLL:    imx6.CCM_ANALOG_PLL_USB2,
	}

	// SD/MMC controller 1
	USDHC1 = &usdhc.USDHC{
		Index: 1,
		Base:  USDHC1_BASE,
		CCGR:  imx6.CCM_CCGR6,
		CG:    imx6.CCGRx_CG1,
	}

	// SD/MMC controller 2
	USDHC2 = &usdhc.USDHC{
		Index: 2,
		Base:  USDHC2_BASE,
		CCGR:  imx6.CCM_CCGR6,
		CG:    imx6.CCGRx_CG2,
	}
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100

// Flag for SDP detection
var sdp bool

func init() {
	if !imx6.Native || imx6.ARM.NonSecure() {
		return
	}

	// On the i.MX6UL family the only way to detect if we are booting
	// through Serial Download Mode over USB is to check whether the USB
	// OTG1 controller was running in device mode prior to our own
	// initialization.
	if reg.Get(USB1_BASE+usb.USB_UOGx_USBMODE, usb.USBMODE_CM, 0b11) == usb.USBMODE_CM_DEVICE &&
		reg.Get(USB1_BASE+usb.USB_UOGx_USBCMD, usb.USBCMD_RS, 1) != 0 {
		sdp = true
	}

	DCP.DeriveKeyMemory.Init()
}

// SDP returns whether Serial Download Protocol over USB has been used to boot
// this runtime. The function always return false on non-secure (e.g. TrustZone
// Normal World) processor modes.
func SDP() bool {
	return sdp
}
