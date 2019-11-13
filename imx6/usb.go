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

package imx6

import (
	"fmt"
	"sync"
	"unsafe"
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

	USB_ANALOG_USB1_CHRG_DETECT      uint32 = 0x020c81b0
	USB_ANALOG_USB1_CHRG_DETECT_EN_B        = 20

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
	USBCMD_RST             = 1
	USBCMD_RS              = 0

	USB_UOG1_USBSTS uint32 = 0x02184144
	USBSTS_URI             = 6
	USBSTS_UI              = 0

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
	ENDPTFLUSH_PETB            = 16
	ENDPTFLUSH_PERB            = 0

	USB_UOG1_ENDPTFLUSH uint32 = 0x021841b4
	ENDPTFLUSH_FETB            = 16
	ENDPTFLUSH_FERB            = 0

	USB_UOG1_ENDPTCOMPLETE uint32 = 0x021841bc
	ENDPTFLUSH_ETCE               = 16
	ENDPTFLUSH_ERCE               = 0
)

type usb struct {
	sync.Mutex

	ccgr     *uint32
	pll      *uint32
	ctrl     *uint32
	pwd      *uint32
	chrg     *uint32
	cmd      *uint32
	ep       *uint32
	sts      *uint32
	sc       *uint32
	setup    *uint32
	flush    *uint32
	prime    *uint32
	complete *uint32

	EP EndPointList
}

var USB1 = &usb{
	ccgr:     (*uint32)(unsafe.Pointer(uintptr(CCM_CCGR6))),
	pll:      (*uint32)(unsafe.Pointer(uintptr(CCM_ANALOG_PLL_USB1))),
	ctrl:     (*uint32)(unsafe.Pointer(uintptr(USBPHY1_CTRL))),
	pwd:      (*uint32)(unsafe.Pointer(uintptr(USBPHY1_PWD))),
	chrg:     (*uint32)(unsafe.Pointer(uintptr(USB_ANALOG_USB1_CHRG_DETECT))),
	cmd:      (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBCMD))),
	ep:       (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTLISTADDR))),
	sts:      (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBSTS))),
	sc:       (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_PORTSC1))),
	setup:    (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTSETUPSTAT))),
	flush:    (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTFLUSH))),
	prime:    (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTPRIME))),
	complete: (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTCOMPLETE))),
}

// Initialize the USB controller.
func (hw *usb) Init() {
	hw.Lock()
	defer hw.Unlock()

	// enable clock
	setN(hw.ccgr, CCM_CCGR6_CG0, 0b11, 0b11)

	// power up PLL
	set(hw.pll, CCM_ANALOG_PLL_USB1_POWER)
	set(hw.pll, CCM_ANALOG_PLL_USB1_EN_USB_CLKS)

	// wait for lock
	print("imx6_usb: waiting for PLL lock...")
	wait(hw.pll, CCM_ANALOG_PLL_USB1_LOCK, 0b1, 1)
	print("done\n")

	// remove bypass
	clear(hw.pll, CCM_ANALOG_PLL_USB1_BYPASS)

	// enable PLL
	set(hw.pll, CCM_ANALOG_PLL_USB1_ENABLE)

	// soft reset USB1 PHY
	set(hw.ctrl, USBPHY1_CTRL_SFTRST)
	clear(hw.ctrl, USBPHY1_CTRL_SFTRST)

	// enable clocks
	clear(hw.ctrl, USBPHY1_CTRL_CLKGATE)

	// clear power down
	*(hw.pwd) = 0x00000000

	// enable UTMI+
	set(hw.ctrl, USBPHY1_CTRL_ENUTMILEVEL3)
	set(hw.ctrl, USBPHY1_CTRL_ENUTMILEVEL2)
	// enable disconnection detect
	set(hw.ctrl, USBPHY1_CTRL_ENHOSTDISCONDETECT)

	// disable charger detector
	clear(hw.chrg, USB_ANALOG_USB1_CHRG_DETECT_EN_B)
}

func (hw *usb) controlTransaction() {
	print("imx6_usb: waiting for bus reset...")
	wait(hw.sts, USBSTS_URI, 0b1, 1)
	print("done\n")

	// p3792, 56.4.6.2.1 Bus Reset, IMX6ULLRM

	// read and write back to clear setup token semaphores
	*(hw.setup) |= *(hw.setup)
	// read and write back to clear setup status
	*(hw.complete) |= *(hw.complete)
	// flush endpoint buffers
	*(hw.flush) = 0xffffffff

	print("imx6_usb: waiting for port reset...")
	wait(hw.sc, PORTSC_PR, 0b1, 0)
	print("done\n")

	// clear reset
	*(hw.sts) |= (1<<USBSTS_URI | 1<<USBSTS_UI)
	// FIXME
	v7_flush_dcache_all()

	print("imx6_usb: waiting for setup packet...")
	wait(hw.setup, 0, 0b1, 1)
	print("done\n")

	// p3801, 56.4.6.4.2.1 Setup Phase, IMX6ULLRM

	// clear setup status
	set(hw.setup, 0)
	// set tripwire
	set(hw.cmd, USBCMD_SUTW)
	// retrieve setup packet
	setup := (&SetupData{}).ParseQH(hw.EP.Get(0, OUT))

	// repeat if necessary
	for get(hw.cmd, USBCMD_SUTW, 0b1) == 0 {
		fmt.Printf("imx6_usb: retrying setup\n")
		set(hw.cmd, USBCMD_SUTW)
		setup = (&SetupData{}).ParseQH(hw.EP.Get(0, OUT))
	}

	// clear tripwire
	clear(hw.cmd, USBCMD_SUTW)
	// flush endpoint buffers
	*(hw.flush) = 0xffffffff

	fmt.Printf("imx6_usb: OUT setup buffer %+v\n", setup)

	return
}
