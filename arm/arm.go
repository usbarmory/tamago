// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package arm

const (
	// Processor modes
	// p1139, ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	USR_MODE = 0x10
	FIQ_MODE = 0x11
	IRQ_MODE = 0x12
	SVC_MODE = 0x13
	MON_MODE = 0x16
	ABT_MODE = 0x17
	HYP_MODE = 0x1a
	UND_MODE = 0x1b
	SYS_MODE = 0x1f
)

type CPU struct {
	// instruction sets
	arm     bool
	thumb   bool
	jazelle bool
	thumbee bool

	// extensions
	programmersModel bool
	security         bool
	mProfileModel    bool
	virtualization   bool
	genericTimer     bool
}

func (c *CPU) Init() {
	c.initFeatures()
}

// defined in arm.s
func read_cpsr() uint32
func read_scr() uint32

func GetProcessorMode() uint8 {
	return uint8(read_cpsr() & 0x1f)
}

func GetProcessorModeName() (mode string) {
	switch GetProcessorMode() {
	case USR_MODE:
		mode = "USR"
	case FIQ_MODE:
		mode = "FIQ"
	case IRQ_MODE:
		mode = "IRQ"
	case SVC_MODE:
		mode = "SVC"
	case MON_MODE:
		mode = "MON"
	case ABT_MODE:
		mode = "ABT"
	case HYP_MODE:
		mode = "HYP"
	case UND_MODE:
		mode = "UND"
	case SYS_MODE:
		mode = "SYS"
	default:
		mode = "Unknown"
	}

	return
}

// Checks SCR.NS bit
func NonSecure() bool {
	return (read_scr() & 1 == 1)
}

func Secure() bool {
	return !NonSecure()
}
