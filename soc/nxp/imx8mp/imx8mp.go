// NXP i.MX8MP configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package imx8mp provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the NXP i.MX8MP family of System-on-Chip (SoC)
// application processors.
//
// The package implements initialization and drivers for NXP i.MX8MP SoCs,
// adopting the following reference specifications:
//   - IMX8MPRM - i.MX 8M Plus Applications Processor Reference Manual - Rev 1 2021/06
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package imx8mp

import (
	"encoding/binary"

	"github.com/usbarmory/tamago/internal/reg"

	"github.com/usbarmory/tamago/arm64"

	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/ocotp"
	"github.com/usbarmory/tamago/soc/nxp/uart"
	"github.com/usbarmory/tamago/soc/nxp/wdog"
)

// Peripheral registers
const (
	// DDR base address
	DDR_BASE = 0x40000000

	// Ethernet MAC
	ENET1_BASE = 0x30be0000

	// On-Chip OTP Controller
	OCOTP_BASE      = 0x021bc000
	OCOTP_BANK_BASE = 0x021bc400

	// Serial ports
	UART1_BASE = 0x30860000
	UART2_BASE = 0x30890000
	UART3_BASE = 0x30880000
	UART4_BASE = 0x30a60000

	// Watchdog Timers
	WDOG1_BASE = 0x30280000
	WDOG2_BASE = 0x30290000
	WDOG3_BASE = 0x302a0000
)

// Peripheral instances
var (
	// ARM64 core
	ARM64 = &arm64.CPU{
		// required before Init()
		TimerOffset: 1,
	}

	// Ethernet MAC 1
	ENET1 = &enet.ENET{
		Index:     1,
		Base:      ENET1_BASE,
		CCGR:      CCM_CCGR10,
		Clock:     GetPeripheralClock,
		EnablePLL: func(_ int) (err error) { return },
	}

	// On-Chip OTP Controller
	OCOTP = &ocotp.OCOTP{
		Base:     OCOTP_BASE,
		BankBase: OCOTP_BANK_BASE,
		CCGR:     CCM_CCGR34,
	}

	// Serial port 1
	UART1 = &uart.UART{
		Index: 1,
		Base:  UART1_BASE,
		CCGR:  CCM_CCGR73,
		Clock: GetUARTClock,
	}

	// Serial port 2
	UART2 = &uart.UART{
		Index: 2,
		Base:  UART2_BASE,
		CCGR:  CCM_CCGR74,
		Clock: GetUARTClock,
	}

	// Watchdog Timer 1
	WDOG1 = &wdog.WDOG{
		Index: 1,
		Base:  WDOG1_BASE,
		CCGR:  CCM_CCGR83,
	}

	// Watchdog Timer 2
	WDOG2 = &wdog.WDOG{
		Index: 2,
		Base:  WDOG2_BASE,
		CCGR:  CCM_CCGR84,
	}

	// Watchdog Timer 3
	WDOG3 = &wdog.WDOG{
		Index: 3,
		Base:  WDOG3_BASE,
		CCGR:  CCM_CCGR85,
	}
)

// SiliconVersion returns the SoC silicon version information
// (p566, 5.1.8.39 DIGPROG Register (CCM_ANALOG_DIGPROG), IMX8MPRM).
func SiliconVersion() (sv, family, revMajor, revMinor uint32) {
	sv = reg.Read(CCM_ANALOG_DIGPROG)

	family = (sv >> 16) & 0xff
	revMajor = (sv >> 8) & 0xff
	revMinor = sv & 0xff

	return
}

// UniqueID returns the NXP SoC Device Unique 128-bit ID.
func UniqueID() (uid [16]byte) {
	otp0, _ := OCOTP.Read(0, 2)
	otp1, _ := OCOTP.Read(0, 3)
	otp2, _ := OCOTP.Read(40, 0)
	otp3, _ := OCOTP.Read(40, 1)

	binary.LittleEndian.PutUint32(uid[0:4], otp0)
	binary.LittleEndian.PutUint32(uid[4:8], otp1)
	binary.LittleEndian.PutUint32(uid[8:12], otp2)
	binary.LittleEndian.PutUint32(uid[8:16], otp3)

	return
}

// Model returns the SoC model name.
func Model() (model string) {
	switch Family {
	case IMX8MP:
		model = "i.MX8MP"
	default:
		model = "unknown"
	}

	return
}
