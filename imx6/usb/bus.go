// NXP USBOH3USBO2 / USBPHY driver
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package usb

import (
	"log"
	"sync"
	"unsafe"

	"github.com/inversepath/tamago/imx6/internal/reg"
)

const (
	CCM_CCGR6     uint32 = 0x20c4080
	CCM_CCGR6_CG0        = 0

	CCM_ANALOG_PLL_USB1             uint32 = 0x020c8010
	CCM_ANALOG_PLL_USB1_LOCK               = 31
	CCM_ANALOG_PLL_USB1_BYPASS             = 16
	CCM_ANALOG_PLL_USB1_ENABLE             = 13
	CCM_ANALOG_PLL_USB1_POWER              = 12
	CCM_ANALOG_PLL_USB1_EN_USB_CLKS        = 6

	USB_ANALOG_USB1_CHRG_DETECT            uint32 = 0x020c81b0
	USB_ANALOG_USB1_CHRG_DETECT_EN_B              = 20
	USB_ANALOG_USB1_CHRG_DETECT_CHK_CHRG_B        = 19

	USBPHY1_PWD uint32 = 0x020c9000

	USBPHY1_CTRL                    uint32 = 0x020c9030
	USBPHY1_CTRL_SFTRST                    = 31
	USBPHY1_CTRL_CLKGATE                   = 30
	USBPHY1_CTRL_ENUTMILEVEL3              = 15
	USBPHY1_CTRL_ENUTMILEVEL2              = 15
	USBPHY1_CTRL_ENHOSTDISCONDETECT        = 1

	// p3823, 56.6 USB Core Memory Map/Register Definition, IMX6ULLRM

	USB_UOG1_USBCMD uint32 = 0x02184140
	USBCMD_SUTW            = 13
	USBCMD_ATDTW           = 12
	USBCMD_RST             = 1
	USBCMD_RS              = 0

	USB_UOG1_USBSTS uint32 = 0x02184144
	USBSTS_URI             = 6
	USBSTS_UI              = 0

	USB_UOG1_DEVICEADDR uint32 = 0x02184154
	DEVICEADDR_USBADR          = 25
	DEVICEADDR_USBADRA         = 24

	USB_UOG1_ENDPTLISTADDR uint32 = 0x02184158
	ENDPTLISTADDR_EPBASE          = 11

	USB_UOG1_PORTSC1 uint32 = 0x02184184
	PORTSC_PTS_1            = 30
	PORTSC_PR               = 8

	USB_UOG1_OTGSC uint32 = 0x021841a4
	OTGSC_OT              = 3

	USB_UOG1_USBMODE  uint32 = 0x021841a8
	USBMODE_SDIS             = 4
	USBMODE_SLOM             = 3
	USBMODE_CM               = 0
	USBMODE_CM_DEVICE        = 0b10
	USBMODE_CM_HOST          = 0b11

	USB_UOG1_ENDPTSETUPSTAT uint32 = 0x021841ac

	USB_UOG1_ENDPTPRIME uint32 = 0x021841b0
	ENDPTPRIME_PETB            = 16
	ENDPTPRIME_PERB            = 0

	USB_UOG1_ENDPTFLUSH uint32 = 0x021841b4
	ENDPTFLUSH_FETB            = 16
	ENDPTFLUSH_FERB            = 0

	USB_UOG1_ENDPTSTAT uint32 = 0x021841b8

	USB_UOG1_ENDPTCOMPLETE uint32 = 0x021841bc
	ENDPTCOMPLETE_ETBR            = 16
	ENDPTCOMPLETE_ERBR            = 0

	USB_UOG1_ENDPTCTRL uint32 = 0x021841c0
	ENDPTCTRL_TXE             = 23
	ENDPTCTRL_TXR             = 22
	ENDPTCTRL_TXI             = 21
	ENDPTCTRL_TXT             = 18
	ENDPTCTRL_TXS             = 16
	ENDPTCTRL_RXE             = 7
	ENDPTCTRL_RXR             = 6
	ENDPTCTRL_RXI             = 5
	ENDPTCTRL_RXT             = 2
	ENDPTCTRL_RXS             = 0
)

type usb struct {
	sync.Mutex

	ccgr     *uint32
	pll      *uint32
	ctrl     *uint32
	pwd      *uint32
	chrg     *uint32
	mode     *uint32
	otg      *uint32
	cmd      *uint32
	addr     *uint32
	sts      *uint32
	sc       *uint32
	ep       *uint32
	setup    *uint32
	flush    *uint32
	prime    *uint32
	stat     *uint32
	complete *uint32
	epctrl   uint32

	EP EndPointList
}

var USB1 = &usb{
	ccgr:     (*uint32)(unsafe.Pointer(uintptr(CCM_CCGR6))),
	pll:      (*uint32)(unsafe.Pointer(uintptr(CCM_ANALOG_PLL_USB1))),
	ctrl:     (*uint32)(unsafe.Pointer(uintptr(USBPHY1_CTRL))),
	pwd:      (*uint32)(unsafe.Pointer(uintptr(USBPHY1_PWD))),
	chrg:     (*uint32)(unsafe.Pointer(uintptr(USB_ANALOG_USB1_CHRG_DETECT))),
	mode:     (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBMODE))),
	otg:      (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_OTGSC))),
	cmd:      (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBCMD))),
	addr:     (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_DEVICEADDR))),
	sts:      (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBSTS))),
	sc:       (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_PORTSC1))),
	ep:       (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTLISTADDR))),
	setup:    (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTSETUPSTAT))),
	flush:    (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTFLUSH))),
	prime:    (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTPRIME))),
	stat:     (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTSTAT))),
	complete: (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTCOMPLETE))),
	epctrl:   USB_UOG1_ENDPTCTRL,
}

// Init initializes the USB controller.
func (hw *usb) Init() {
	hw.Lock()
	defer hw.Unlock()

	// enable clock
	reg.SetN(hw.ccgr, CCM_CCGR6_CG0, 0b11, 0b11)

	// power up PLL
	reg.Set(hw.pll, CCM_ANALOG_PLL_USB1_POWER)
	reg.Set(hw.pll, CCM_ANALOG_PLL_USB1_EN_USB_CLKS)

	// wait for lock
	log.Printf("imx6_usb: waiting for PLL lock\n")
	reg.Wait(hw.pll, CCM_ANALOG_PLL_USB1_LOCK, 0b1, 1)

	// remove bypass
	reg.Clear(hw.pll, CCM_ANALOG_PLL_USB1_BYPASS)

	// enable PLL
	reg.Set(hw.pll, CCM_ANALOG_PLL_USB1_ENABLE)

	// soft reset USB1 PHY
	reg.Set(hw.ctrl, USBPHY1_CTRL_SFTRST)
	reg.Clear(hw.ctrl, USBPHY1_CTRL_SFTRST)

	// enable clocks
	reg.Clear(hw.ctrl, USBPHY1_CTRL_CLKGATE)

	// clear power down
	*(hw.pwd) = 0x00000000

	// enable UTMI+
	reg.Set(hw.ctrl, USBPHY1_CTRL_ENUTMILEVEL3)
	reg.Set(hw.ctrl, USBPHY1_CTRL_ENUTMILEVEL2)
	// enable disconnection detect
	reg.Set(hw.ctrl, USBPHY1_CTRL_ENHOSTDISCONDETECT)

	// disable charger detector
	reg.Set(hw.chrg, USB_ANALOG_USB1_CHRG_DETECT_EN_B)
	reg.Set(hw.chrg, USB_ANALOG_USB1_CHRG_DETECT_CHK_CHRG_B)
}

// Reset waits for and handles a USB bus reset.
func (hw *usb) Reset() {
	hw.Lock()
	defer hw.Unlock()

	log.Printf("imx6_usb: waiting for bus reset\n")
	reg.Wait(hw.sts, USBSTS_URI, 0b1, 1)

	// p3792, 56.4.6.2.1 Bus Reset, IMX6ULLRM

	// read and write back to clear setup token semaphores
	*(hw.setup) |= *(hw.setup)
	// read and write back to clear setup status
	*(hw.complete) |= *(hw.complete)
	// flush endpoint buffers
	*(hw.flush) = 0xffffffff

	log.Printf("imx6_usb: waiting for port reset\n")
	reg.Wait(hw.sc, PORTSC_PR, 0b1, 0)

	// clear reset
	*(hw.sts) |= (1<<USBSTS_URI | 1<<USBSTS_UI)
}
