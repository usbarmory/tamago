// SpacemiT K1 Multi-Function Pin Register (MFPR) support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package k1

import (
	"fmt"
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

// Alternate functions
// (p135, Section 4.7.2 MFPR Functional Description, K1 Datasheet).
const (
	AF0 = iota
	AF1
	AF2
	AF3
	AF4
	AF5
	AF6
	AF7
)

// Drive strength settings for the MFPR DRIVE[2:0] field
// (p134, Section 4.7.2 MFPR Functional Description, K1 Datasheet).
const (
	// DRIVE_SLOW selects the slowest slew rate/drive strength.
	DRIVE_SLOW = 0b000
	// DRIVE_MEDIUM is the setting recommended for all GPIOs except the SD
	// card interface.
	DRIVE_MEDIUM = 0b010
	// DRIVE_FAST is the setting recommended for the SD card interface.
	DRIVE_FAST = 0b110
)

// MFPR fields
// (p133, Section 4.7.2 MFPR Functional Description, K1 Datasheet).
const (
	MFPR_PULL_SEL      = 15
	MFPR_PULLUP_EN     = 14
	MFPR_PULLDN_EN     = 13
	MFPR_DRIVE_LO      = 11 // DRIVE[1:0]
	MFPR_DRIVE_HI      = 10 // DRIVE[2]
	MFPR_ST            = 8  // ST[1:0]
	MFPR_SLE           = 7
	MFPR_EDGE_CLEAR    = 6
	MFPR_EDGE_FALL_EN  = 5
	MFPR_EDGE_RISE_EN  = 4
	MFPR_SPU           = 3
	MFPR_AF_SEL        = 0
	MFPR_AF_SEL_MASK   = 0b111
	MFPR_ST_MASK       = 0b11
	MFPR_DRIVE_LO_MASK = 0b11
)

// MFPR offsets for pads which are not named after a GPIO number
// (p128, Section 4.7 Multi-Function Pin Register (MFPRs), K1 Datasheet).
const (
	MFPR_PRI_TDI    = 0x11c
	MFPR_PRI_TMS    = 0x120
	MFPR_PRI_TCK    = 0x124
	MFPR_PRI_TDO    = 0x128
	MFPR_QSPI_DAT0  = 0x168
	MFPR_QSPI_DAT1  = 0x16c
	MFPR_QSPI_DAT2  = 0x170
	MFPR_QSPI_DAT3  = 0x174
	MFPR_QSPI_CS1   = 0x178
	MFPR_QSPI_CLK   = 0x17c
	MFPR_MMC1_DAT3  = 0x1b8
	MFPR_MMC1_DAT2  = 0x1bc
	MFPR_MMC1_DAT1  = 0x1c0
	MFPR_MMC1_DAT0  = 0x1c4
	MFPR_MMC1_CMD   = 0x1c8
	MFPR_MMC1_CLK   = 0x1cc
	MFPR_PWR_SCL    = 0x1d4
	MFPR_PWR_SDA    = 0x1d8
	MFPR_VCXO_EN    = 0x1dc
	MFPR_DVL0       = 0x1e0
	MFPR_DVL1       = 0x1e4
	MFPR_PMIC_INT_N = 0x1e8
)

// Pad represents a K1 multi-function I/O pad instance, controlled through its
// Multi-Function Pin Register (MFPR).
type Pad struct {
	// Register is the pad MFPR absolute address.
	Register uint32
}

// NewPad returns the pad instance for the argument MFPR offset from
// [MFPR_BASE] (p128, Section 4.7, K1 Datasheet).
func NewPad(offset uint32) *Pad {
	return &Pad{
		Register: MFPR_BASE + offset,
	}
}

// GPIO returns the pad instance for the argument GPIO number.
//
// The GPIO numbers 93 to 109 are not accepted as the K1 Datasheet assigns them
// as alternate functions of dedicated pads (p118, Section 4.5 Multi-Function
// I/O Pin Assignments), whose MFPRs are inconsistently shared between the
// PWR/QSPI/SD and eMMC groups, such pads must therefore be selected explicitly
// through [NewPad] and the relevant MFPR offset constant.
func GPIO(num int) (pad *Pad, err error) {
	var offset uint32

	switch {
	case num >= 0 && num <= 85:
		// GPIO_00 (0x004) to GPIO_85 (0x158), the JTAG pads PRI_TDI,
		// PRI_TMS, PRI_TCK and PRI_TDO fill the GPIO 70 to 73 slots.
		offset = 0x004 + uint32(num)*4
	case num >= 86 && num <= 92:
		// GPIO_86 (0x1ec) to GPIO_92 (0x204)
		offset = 0x1ec + uint32(num-86)*4
	case num == 110:
		// GPIO_110 sits apart from the other GPIO MFPRs
		offset = 0x1d0
	case num >= 111 && num <= 127:
		// GPIO_111 (0x20c) to GPIO_127 (0x24c)
		offset = 0x20c + uint32(num-111)*4
	default:
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	return NewPad(offset), nil
}

// ConfigureGPIO selects the alternate function for the pad controlling the
// argument GPIO number, see [GPIO] and [Pad.Mode].
func ConfigureGPIO(num int, af int) (err error) {
	pad, err := GPIO(num)

	if err != nil {
		return
	}

	pad.Mode(af)

	return
}

// setN updates a field of the pad MFPR, see [read] on why the internal/reg
// read-modify-write helpers are not used.
func (p *Pad) setN(pos int, mask uint32, val uint32) {
	r := reg.Read(p.Register)
	r = (r & ^(mask << pos)) | ((val & mask) << pos)

	reg.Write(p.Register, r)
}

// setTo sets or clears a bit of the pad MFPR.
func (p *Pad) setTo(pos int, val bool) {
	var bit uint32

	if val {
		bit = 1
	}

	p.setN(pos, 1, bit)
}

// Mode selects the pad alternate function (AF0 - AF7).
func (p *Pad) Mode(af int) {
	p.setN(MFPR_AF_SEL, MFPR_AF_SEL_MASK, uint32(af))
}

// Pull enables or disables the pad internal pull-up and pull-down resistors,
// overriding the configuration of the selected alternate function.
func (p *Pad) Pull(up bool, down bool) {
	p.setTo(MFPR_PULL_SEL, true)
	p.setTo(MFPR_PULLUP_EN, up)
	p.setTo(MFPR_PULLDN_EN, down)
}

// Drive sets the pad drive strength and slew rate (DRIVE[2:0]), see
// [DRIVE_SLOW], [DRIVE_MEDIUM] and [DRIVE_FAST].
//
// The field is not contiguous, DRIVE[1:0] is held at bits 12:11 while DRIVE[2]
// is held at bit 10 (p134, Section 4.7.2, K1 Datasheet).
func (p *Pad) Drive(strength int) {
	ds := uint32(strength)

	p.setN(MFPR_DRIVE_LO, MFPR_DRIVE_LO_MASK, ds)
	p.setTo(MFPR_DRIVE_HI, ds>>2&1 == 1)
}

// SchmittTrigger sets the pad input threshold, a zero value selects the plain
// buffer input while values 1 to 3 enable the Schmitt trigger with increasing
// hysteresis (p132, Section 4.7.1.1 I/O PAD Paramenter Definition, K1
// Datasheet).
func (p *Pad) SchmittTrigger(threshold int) {
	p.setN(MFPR_ST, MFPR_ST_MASK, uint32(threshold))
}

// SlewRate enables or disables the pad slew rate output control, which slows
// down the output ramp for EMI considerations.
func (p *Pad) SlewRate(enable bool) {
	p.setTo(MFPR_SLE, enable)
}

// StrongPull enables or disables the pad strong pull resistor, required by I2C
// and SD card pads.
func (p *Pad) StrongPull(enable bool) {
	p.setTo(MFPR_SPU, enable)
}

// Value returns the pad MFPR value.
func (p *Pad) Value() uint32 {
	return reg.Read(p.Register)
}

// SetValue sets the pad MFPR value.
func (p *Pad) SetValue(val uint32) {
	reg.Write(p.Register, val)
}
