// NXP i.MX6UL initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6ul

import (
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/dcp"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/usb"
)

// i.MX processor families
const (
	IMX6UL  = 0x64
	IMX6ULL = 0x65
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100

var (
	// Processor family
	Family uint32

	// Flag native or emulated processor
	Native bool

	// SDP flags whether Serial Download Protocol over USB has been used to
	// boot this runtime. The value is always false on non-secure (e.g.
	// TrustZone Normal World) processor modes.
	SDP bool
)

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup (e.g. runtime.hwinit).
func Init() {
	if ARM.Mode() != arm.SYS_MODE {
		// initialization required only when in PL1
		return
	}

	ramStart, _ := runtime.MemRegion()

	ARM.Init(ramStart)
	ARM.EnableVFP()

	// required when booting in SDP mode
	ARM.EnableSMP()

	// MMU initialization is required to take advantage of data cache
	ARM.InitMMU()
	ARM.EnableCache()

	_, fam, revMajor, revMinor := SiliconVersion()
	Family = fam

	if revMajor != 0 || revMinor != 0 {
		Native = true
	}

	initTimers()
}

func init() {
	// Initialize watchdogs, this must be done within 16 seconds to clear
	// their power-down counter event
	// (p4085, 59.5.3 Power-down counter event, IMX6ULLRM).
	WDOG1.Init()
	WDOG2.Init()
	WDOG3.Init()

	// use internal OCRAM (iRAM) as default DMA region
	dma.Init(OCRAM_START, OCRAM_SIZE)

	OCOTP.Init()
	model := Model()

	switch model {
	case "i.MX6UL":
		if Native {
			// Configuration and Manufacturing Info
			// p2203, 35.5.15 OCOTP_CFG3, IMX6ULRM
			cfg3, _ := OCOTP.Read(0, 4)

			// BEE_UNAVAILABLE
			if bits.Get(&cfg3, 25, 1) == 0 {
				// Bus Encryption Engine
				BEE = &bee.BEE{
					Base: BEE_BASE,
					SNVS: SNVS,
				}

				SNVS.DryIce = SNVS_LP_BASE
			}
		}

		OCOTP.Banks = 16
	case "i.MX6ULL", "i.MX6ULZ":
		// Data Co-Processor
		DCP = &dcp.DCP{
			Base:            DCP_BASE,
			CCGR:            CCM_CCGR0,
			CG:              CCGRx_CG5,
			DeriveKeyMemory: dma.Default(),
		}

		OCOTP.Banks = 8
	}

	switch model {
	case "i.MX6UL", "i.MX6ULL":
		// Ethernet MAC 1
		ENET1 = &enet.ENET{
			Index:     1,
			Base:      ENET1_BASE,
			CCGR:      CCM_CCGR0,
			CG:        CCGRx_CG6,
			Clock:     GetPeripheralClock,
			IRQ:       ENET1_IRQ,
			EnablePLL: EnableENETPLL,
		}

		// Ethernet MAC 2
		ENET2 = &enet.ENET{
			Index:     2,
			Base:      ENET2_BASE,
			CCGR:      CCM_CCGR0,
			CG:        CCGRx_CG6,
			Clock:     GetPeripheralClock,
			IRQ:       ENET2_IRQ,
			EnablePLL: EnableENETPLL,
		}
	}

	if !Native || ARM.NonSecure() {
		return
	}

	// OCOTP_ANA1 Temperature Sensor Calibration Data
	// p3531, 52.2 Software Usage Guidelines, IMX6ULLRM
	ana1, _ := OCOTP.Read(1, 6)
	// initialize temperature monitor
	TEMPMON.Init(ana1)

	// DryIce Trim Values
	// p1060, 8.5.2 DryIce Trim Value Register, IMX6ULSRM
	mem1, _ := OCOTP.Read(1, 1)
	// initialize security state machine (SSM)
	SNVS.Init(mem1)

	// On the i.MX6UL family the only way to detect if we are booting
	// through Serial Download Mode over USB is to check whether the USB
	// OTG1 controller was running in device mode prior to our own
	// initialization.
	if reg.Get(USB1_BASE+usb.USB_UOGx_USBMODE, usb.USBMODE_CM, 0b11) == usb.USBMODE_CM_DEVICE &&
		reg.Get(USB1_BASE+usb.USB_UOGx_USBCMD, usb.USBCMD_RS, 1) != 0 {
		SDP = true
	}
}
