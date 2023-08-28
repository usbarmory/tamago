// NXP USBOH3USBO2 / USBPHY driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usb implements a driver for the USB PHY designated as NXP
// USBOH3USBO2, included in several i.MX SoCs, adopting the following
// specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//   - USB2.0    - USB Specification Revision 2.0
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package usb

import (
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// USB registers
const (
	USB_ANALOG_USBx_CHRG_DETECT = 0x10
	CHRG_DETECT_EN_B            = 20
	CHRG_DETECT_CHK_CHRG_B      = 19

	USBPHYx_PWD = 0x00

	USBPHYx_CTRL            = 0x30
	CTRL_SFTRST             = 31
	CTRL_CLKGATE            = 30
	CTRL_ENUTMILEVEL3       = 15
	CTRL_ENUTMILEVEL2       = 14
	CTRL_ENHOSTDISCONDETECT = 1

	// p3823, 56.6 USB Core Memory Map/Register Definition, IMX6ULLRM

	USB_UOGx_USBCMD = 0x140
	USBCMD_RST      = 1
	USBCMD_RS       = 0

	USB_UOGx_USBSTS = 0x144
	USBSTS_URI      = 6
	USBSTS_UI       = 0

	USB_UOGx_USBINTR = 0x148

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

// USB interrupt events
const (
	// p3848, 56.6.19 Interrupt Status Register (USB_nUSBSTS),  IMX6ULLRM
	// p3852, 56.6.20 Interrupt Enable Register (USB_nUSBINTR), IMX6ULLRM

	IRQ_TI1   = 25
	IRQ_TI0   = 24
	IRQ_NAKI  = 16
	IRQ_AS    = 15
	IRQ_PS    = 14
	IRQ_RCP   = 13
	IRQ_HCH   = 12
	IRQ_ULPII = 10
	IRQ_SLI   = 8
	IRQ_SRI   = 7
	IRQ_URI   = 6
	IRQ_AAI   = 5
	IRQ_SEI   = 4
	IRQ_FRI   = 3
	IRQ_PCI   = 2
	IRQ_UEI   = 1
	IRQ_UI    = 0
)

// USB represents a USB controller instance.
type USB struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Analog base register
	Analog uint32
	// PHY base register
	PHY uint32
	// Interrupt ID
	IRQ int
	// PLL enable function
	EnablePLL func(index int) error

	// USB device configuration
	Device *Device

	// EP1-N transfer completion rendezvous point
	event *sync.Cond
	// EP1-N cancellation signal
	exit chan struct{}
	// EP-1-N completion synchronization
	wg sync.WaitGroup

	// control registers
	ctrl     uint32
	pwd      uint32
	chrg     uint32
	mode     uint32
	otg      uint32
	cmd      uint32
	addr     uint32
	sts      uint32
	intr     uint32
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

// Init initializes the USB controller.
func (hw *USB) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 || hw.Analog == 0 || hw.PHY == 0 || hw.EnablePLL == nil {
		panic("invalid USB controller instance")
	}

	hw.ctrl = hw.PHY + USBPHYx_CTRL
	hw.pwd = hw.PHY + USBPHYx_PWD
	hw.chrg = hw.Analog + USB_ANALOG_USBx_CHRG_DETECT
	hw.mode = hw.Base + USB_UOGx_USBMODE
	hw.otg = hw.Base + USB_UOGx_OTGSC
	hw.cmd = hw.Base + USB_UOGx_USBCMD
	hw.addr = hw.Base + USB_UOGx_DEVICEADDR
	hw.sts = hw.Base + USB_UOGx_USBSTS
	hw.intr = hw.Base + USB_UOGx_USBINTR
	hw.sc = hw.Base + USB_UOGx_PORTSC1
	hw.eplist = hw.Base + USB_UOGx_ENDPTLISTADDR
	hw.setup = hw.Base + USB_UOGx_ENDPTSETUPSTAT
	hw.flush = hw.Base + USB_UOGx_ENDPTFLUSH
	hw.prime = hw.Base + USB_UOGx_ENDPTPRIME
	hw.stat = hw.Base + USB_UOGx_ENDPTSTAT
	hw.complete = hw.Base + USB_UOGx_ENDPTCOMPLETE
	hw.epctrl = hw.Base + USB_UOGx_ENDPTCTRL

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
	hw.EnablePLL(hw.Index)

	// soft reset USB PHY
	reg.Set(hw.ctrl, CTRL_SFTRST)
	reg.Clear(hw.ctrl, CTRL_SFTRST)

	// disable clock gate
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

// EnableInterrupt enables interrupt generation for a specific event.
func (hw *USB) EnableInterrupt(event int) {
	reg.Set(hw.intr, event)
}

// ClearInterrupt clears the interrupt corresponding to a specific event.
func (hw *USB) ClearInterrupt(event int) {
	reg.Set(hw.sts, event)
}
