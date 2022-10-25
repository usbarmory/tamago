// NXP IOMUXC support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package iomuxc implements helpers for IOMUX configuration on NXP SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package iomuxc

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// IOMUXC registers
const (
	SW_PAD_CTL_HYS = 16

	SW_PAD_CTL_PUS                = 14
	SW_PAD_CTL_PUS_PULL_DOWN_100K = 0b00
	SW_PAD_CTL_PUS_PULL_UP_47K    = 0b01
	SW_PAD_CTL_PUS_PULL_UP_100K   = 0b10
	SW_PAD_CTL_PUS_PULL_UP_22K    = 0b11

	SW_PAD_CTL_PUE = 13
	SW_PAD_CTL_PKE = 12
	SW_PAD_CTL_ODE = 11

	SW_PAD_CTL_SPEED        = 6
	SW_PAD_CTL_SPEED_50MHZ  = 0b00
	SW_PAD_CTL_SPEED_100MHZ = 0b10
	SW_PAD_CTL_SPEED_200MHZ = 0b11

	SW_PAD_CTL_DSE                        = 3
	SW_PAD_CTL_DSE_OUTPUT_DRIVER_DISABLED = 0b000
	SW_PAD_CTL_DSE_2_R0_2                 = 0b010
	SW_PAD_CTL_DSE_2_R0_3                 = 0b011
	SW_PAD_CTL_DSE_2_R0_4                 = 0b100
	SW_PAD_CTL_DSE_2_R0_5                 = 0b101
	SW_PAD_CTL_DSE_2_R0_6                 = 0b110
	SW_PAD_CTL_DSE_2_R0_7                 = 0b111

	SW_PAD_CTL_SRE = 0

	SW_MUX_CTL_SION     = 4
	SW_MUX_CTL_MUX_MODE = 0
)

// Pad instance.
type Pad struct {
	// Mux register (e.g. IOMUXC_SW_MUX_CTL_PAD_*)
	Mux uint32
	// Pad register (e.g. IOMUXC_SW_PAD_CTL_PAD_*)
	Pad uint32
	// Daisy register (e.g. IOMUXC_*_SELECT_INPUT)
	Daisy uint32
}

// Init initializes a pad.
func Init(mux uint32, pad uint32, mode uint32) (p *Pad) {
	p = &Pad{
		Mux: mux,
		Pad: pad,
	}

	p.Mode(mode)

	return
}

// Mode configures the pad iomux mode.
func (pad *Pad) Mode(mode uint32) {
	reg.SetN(pad.Mux, SW_MUX_CTL_MUX_MODE, 0b1111, mode)
}

// SoftwareInput configures the pad SION bit.
func (pad *Pad) SoftwareInput(enabled bool) {
	reg.SetTo(pad.Mux, SW_MUX_CTL_SION, enabled)
}

// Ctl configures the pad control register.
func (pad *Pad) Ctl(ctl uint32) {
	reg.Write(pad.Pad, ctl)
}

// Select configures the pad daisy chain register.
func (pad *Pad) Select(input uint32) {
	if pad.Daisy == 0 {
		return
	}

	reg.Write(pad.Daisy, input)
}
