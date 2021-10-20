// NXP USBOH3USBO2 / USBPHY driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usb implements a driver for the USB PHY designated as NXP
// USBOH3USBO2, included in i.MX6 SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package usb

import (
	"sync"

	"github.com/f-secure-foundry/tamago/internal/reg"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

// USB registers
const (
	USB_ANALOG1_BASE = 0x020c81a0
	USB_ANALOG2_BASE = 0x020c8200

	USB_ANALOG_USBx_CHRG_DETECT = 0x10
	CHRG_DETECT_EN_B            = 20
	CHRG_DETECT_CHK_CHRG_B      = 19

	USBPHY1_BASE = 0x020c9000
	USBPHY2_BASE = 0x020ca000

	USBPHYx_PWD = 0x00

	USBPHYx_CTRL            = 0x30
	CTRL_SFTRST             = 31
	CTRL_CLKGATE            = 30
	CTRL_ENUTMILEVEL3       = 15
	CTRL_ENUTMILEVEL2       = 14
	CTRL_ENHOSTDISCONDETECT = 1

	USB1_BASE = 0x02184000
	USB2_BASE = 0x02184200

	// p3823, 56.6 USB Core Memory Map/Register Definition, IMX6ULLRM

	USB_UOGx_USBCMD = 0x140
	USBCMD_RST      = 1
	USBCMD_RS       = 0

	USB_UOGx_USBSTS = 0x144
	USBSTS_URI      = 6
	USBSTS_UI       = 0

	USB_UOGx_DEVICEADDR = 0x154
	DEVICEADDR_USBADR   = 25
	DEVICEADDR_USBADRA  = 24

	USB_UOGx_ENDPTLISTADDR = 0x158
	ENDPTLISTADDR_EPBASE   = 11

	USB_UOGx_PORTSC1 = 0x184
	PORTSC_PTS_1     = 30
	PORTSC_PSPD      = 26
	PORTSC_PR        = 8

	USB_UOGx_OTGSC = 0x1a4
	OTGSC_OT       = 3

	USB_UOGx_USBMODE  = 0x1a8
	USBMODE_SDIS      = 4
	USBMODE_SLOM      = 3
	USBMODE_CM        = 0
	USBMODE_CM_DEVICE = 0b10
	USBMODE_CM_HOST   = 0b11

	USB_UOGx_ENDPTSETUPSTAT = 0x1ac

	USB_UOGx_ENDPTPRIME = 0x1b0
	ENDPTPRIME_PETB     = 16
	ENDPTPRIME_PERB     = 0

	USB_UOGx_ENDPTFLUSH = 0x1b4
	ENDPTFLUSH_FETB     = 16
	ENDPTFLUSH_FERB     = 0

	USB_UOGx_ENDPTSTAT = 0x1b8

	USB_UOGx_ENDPTCOMPLETE = 0x1bc
	ENDPTCOMPLETE_ETBR     = 16
	ENDPTCOMPLETE_ERBR     = 0

	USB_UOGx_ENDPTCTRL = 0x1c0
	ENDPTCTRL_TXE      = 23
	ENDPTCTRL_TXR      = 22
	ENDPTCTRL_TXI      = 21
	ENDPTCTRL_TXT      = 18
	ENDPTCTRL_TXS      = 16
	ENDPTCTRL_RXE      = 7
	ENDPTCTRL_RXR      = 6
	ENDPTCTRL_RXI      = 5
	ENDPTCTRL_RXT      = 2
	ENDPTCTRL_RXS      = 0
)

// USB represents a controller instance.
type USB struct {
	sync.Mutex

	// signal for EP1-N cancellation
	done chan bool
	// controller index
	n int

	// control registers
	pll      uint32
	ctrl     uint32
	pwd      uint32
	chrg     uint32
	mode     uint32
	otg      uint32
	cmd      uint32
	addr     uint32
	sts      uint32
	sc       uint32
	eplist   uint32
	setup    uint32
	flush    uint32
	prime    uint32
	stat     uint32
	complete uint32
	epctrl   uint32

	// cache for endpoint list pointer
	epListAddr uint32
	// cache for endpoint queue heads pointers
	dQH [MAX_ENDPOINTS][2]uint32
}

// USB1 instance
var USB1 = &USB{n: 1}

// USB2 instance
var USB2 = &USB{n: 2}

// Init initializes the USB controller.
func (hw *USB) Init() {
	var base uint32
	var analogBase uint32
	var phyBase uint32

	hw.Lock()
	defer hw.Unlock()

	switch hw.n {
	case 1:
		base = USB1_BASE
		analogBase = USB_ANALOG1_BASE
		phyBase = USBPHY1_BASE
		hw.pll = imx6.CCM_ANALOG_PLL_USB1
	case 2:
		base = USB2_BASE
		analogBase = USB_ANALOG2_BASE
		phyBase = USBPHY2_BASE
		hw.pll = imx6.CCM_ANALOG_PLL_USB2
	default:
		panic("invalid USB controller instance")
	}

	hw.ctrl = phyBase + USBPHYx_CTRL
	hw.pwd = phyBase + USBPHYx_PWD
	hw.chrg = analogBase + USB_ANALOG_USBx_CHRG_DETECT
	hw.mode = base + USB_UOGx_USBMODE
	hw.otg = base + USB_UOGx_OTGSC
	hw.cmd = base + USB_UOGx_USBCMD
	hw.addr = base + USB_UOGx_DEVICEADDR
	hw.sts = base + USB_UOGx_USBSTS
	hw.sc = base + USB_UOGx_PORTSC1
	hw.eplist = base + USB_UOGx_ENDPTLISTADDR
	hw.setup = base + USB_UOGx_ENDPTSETUPSTAT
	hw.flush = base + USB_UOGx_ENDPTFLUSH
	hw.prime = base + USB_UOGx_ENDPTPRIME
	hw.stat = base + USB_UOGx_ENDPTSTAT
	hw.complete = base + USB_UOGx_ENDPTCOMPLETE
	hw.epctrl = base + USB_UOGx_ENDPTCTRL

	// enable clock
	reg.SetN(imx6.CCM_CCGR6, imx6.CCGRx_CG0, 0b11, 0b11)

	// power up PLL
	reg.Set(hw.pll, imx6.PLL_POWER)
	reg.Set(hw.pll, imx6.PLL_EN_USB_CLKS)

	// wait for lock
	reg.Wait(hw.pll, imx6.PLL_LOCK, 1, 1)

	// remove bypass
	reg.Clear(hw.pll, imx6.PLL_BYPASS)

	// enable PLL
	reg.Set(hw.pll, imx6.PLL_ENABLE)

	// soft reset USB PHY
	reg.Set(hw.ctrl, CTRL_SFTRST)
	reg.Clear(hw.ctrl, CTRL_SFTRST)

	// enable clocks
	reg.Clear(hw.ctrl, CTRL_CLKGATE)

	// clear power down
	reg.Write(hw.pwd, 0)

	// enable UTMI+
	reg.Set(hw.ctrl, CTRL_ENUTMILEVEL3)
	reg.Set(hw.ctrl, CTRL_ENUTMILEVEL2)
	// enable disconnection detect
	reg.Set(hw.ctrl, CTRL_ENHOSTDISCONDETECT)

	// disable charger detector
	reg.Set(hw.chrg, CHRG_DETECT_EN_B)
	reg.Set(hw.chrg, CHRG_DETECT_CHK_CHRG_B)
}

// Speed returns the port speed.
func (hw *USB) Speed() (speed string) {
	hw.Lock()
	defer hw.Unlock()

	switch reg.Get(hw.sc, PORTSC_PSPD, 0b11) {
	case 0b00:
		speed = "full"
	case 0b01:
		speed = "low"
	case 0b10:
		speed = "high"
	case 0b11:
		panic("invalid port speed")
	}

	return
}

// PowerDown shuts down the USB PHY.
func (hw *USB) PowerDown() {
	reg.Write(hw.pwd, 0xffffffff)
}

// Run sets the controller in run mode.
func (hw *USB) Run() {
	reg.Set(hw.cmd, USBCMD_RS)
}

// Stop sets the controller in stop mode.
func (hw *USB) Stop() {
	reg.Clear(hw.cmd, USBCMD_RS)
}

// Reset waits for and handles a bus reset.
func (hw *USB) Reset() {
	hw.Lock()
	defer hw.Unlock()

	reg.Wait(hw.sts, USBSTS_URI, 1, 1)

	// p3792, 56.4.6.2.1 Bus Reset, IMX6ULLRM

	// read and write back to clear setup token semaphores
	reg.WriteBack(hw.setup)
	// read and write back to clear setup status
	reg.WriteBack(hw.complete)
	// flush endpoint buffers
	reg.Write(hw.flush, 0xffffffff)

	reg.Wait(hw.sc, PORTSC_PR, 1, 0)

	// clear reset
	reg.Or(hw.sts, (1<<USBSTS_URI | 1<<USBSTS_UI))
}
