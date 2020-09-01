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
	"log"
	"sync"

	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// USB registers
const (
	CCM_ANALOG_PLL_USB1 = 0x020c8010
	PLL_EN_USB_CLKS     = 6

	USB_ANALOG_USB1_CHRG_DETECT = 0x020c81b0
	CHRG_DETECT_EN_B            = 20
	CHRG_DETECT_CHK_CHRG_B      = 19

	USBPHY1_PWD = 0x020c9000

	USBPHY1_CTRL            = 0x020c9030
	CTRL_SFTRST             = 31
	CTRL_CLKGATE            = 30
	CTRL_ENUTMILEVEL3       = 15
	CTRL_ENUTMILEVEL2       = 15
	CTRL_ENHOSTDISCONDETECT = 1

	// p3823, 56.6 USB Core Memory Map/Register Definition, IMX6ULLRM

	USB_UOG1_USBCMD = 0x02184140
	USBCMD_SUTW     = 13
	USBCMD_ATDTW    = 12
	USBCMD_RST      = 1
	USBCMD_RS       = 0

	USB_UOG1_USBSTS = 0x02184144
	USBSTS_URI      = 6
	USBSTS_UI       = 0

	USB_UOG1_DEVICEADDR = 0x02184154
	DEVICEADDR_USBADR   = 25
	DEVICEADDR_USBADRA  = 24

	USB_UOG1_ENDPTLISTADDR = 0x02184158
	ENDPTLISTADDR_EPBASE   = 11

	USB_UOG1_PORTSC1 = 0x02184184
	PORTSC_PTS_1     = 30
	PORTSC_PSPD      = 26
	PORTSC_PR        = 8

	USB_UOG1_OTGSC = 0x021841a4
	OTGSC_OT       = 3

	USB_UOG1_USBMODE  = 0x021841a8
	USBMODE_SDIS      = 4
	USBMODE_SLOM      = 3
	USBMODE_CM        = 0
	USBMODE_CM_DEVICE = 0b10
	USBMODE_CM_HOST   = 0b11

	USB_UOG1_ENDPTSETUPSTAT = 0x021841ac

	USB_UOG1_ENDPTPRIME = 0x021841b0
	ENDPTPRIME_PETB     = 16
	ENDPTPRIME_PERB     = 0

	USB_UOG1_ENDPTFLUSH = 0x021841b4
	ENDPTFLUSH_FETB     = 16
	ENDPTFLUSH_FERB     = 0

	USB_UOG1_ENDPTSTAT = 0x021841b8

	USB_UOG1_ENDPTCOMPLETE = 0x021841bc
	ENDPTCOMPLETE_ETBR     = 16
	ENDPTCOMPLETE_ERBR     = 0

	USB_UOG1_ENDPTCTRL = 0x021841c0
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
}

// USB1 instance
var USB1 = &USB{
	pll:      CCM_ANALOG_PLL_USB1,
	ctrl:     USBPHY1_CTRL,
	pwd:      USBPHY1_PWD,
	chrg:     USB_ANALOG_USB1_CHRG_DETECT,
	mode:     USB_UOG1_USBMODE,
	otg:      USB_UOG1_OTGSC,
	cmd:      USB_UOG1_USBCMD,
	addr:     USB_UOG1_DEVICEADDR,
	sts:      USB_UOG1_USBSTS,
	sc:       USB_UOG1_PORTSC1,
	eplist:   USB_UOG1_ENDPTLISTADDR,
	setup:    USB_UOG1_ENDPTSETUPSTAT,
	flush:    USB_UOG1_ENDPTFLUSH,
	prime:    USB_UOG1_ENDPTPRIME,
	stat:     USB_UOG1_ENDPTSTAT,
	complete: USB_UOG1_ENDPTCOMPLETE,
	epctrl:   USB_UOG1_ENDPTCTRL,
}

// Init initializes the USB controller.
func (hw *USB) Init() {
	hw.Lock()
	defer hw.Unlock()

	// enable clock
	reg.SetN(imx6.CCM_CCGR6, imx6.CCGR_CG0, 0b11, 0b11)

	// power up PLL
	reg.Set(hw.pll, imx6.PLL_POWER)
	reg.Set(hw.pll, PLL_EN_USB_CLKS)

	// wait for lock
	log.Printf("imx6_usb: waiting for PLL lock")
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

	log.Printf("imx6_usb: waiting for bus reset")
	reg.Wait(hw.sts, USBSTS_URI, 1, 1)

	// p3792, 56.4.6.2.1 Bus Reset, IMX6ULLRM

	// read and write back to clear setup token semaphores
	reg.WriteBack(hw.setup)
	// read and write back to clear setup status
	reg.WriteBack(hw.complete)
	// flush endpoint buffers
	reg.Write(hw.flush, 0xffffffff)

	log.Printf("imx6_usb: waiting for port reset")
	reg.Wait(hw.sc, PORTSC_PR, 1, 0)

	// clear reset
	reg.Or(hw.sts, (1<<USBSTS_URI | 1<<USBSTS_UI))
}
