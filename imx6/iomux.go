// NXP i.MX6 IOMUX driver
// https://github.com/f-secure-foundry/tamago
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

	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	IOMUXC_START uint32 = 0x020e0000
	IOMUXC_END   uint32 = 0x0203ffff

	SW_PAD_CTL_HYS                = 16

	SW_PAD_CTL_PUS                = 14
	SW_PAD_CTL_PUS_PULL_DOWN_100K = 0
	SW_PAD_CTL_PUS_PULL_UP_47K    = 1
	SW_PAD_CTL_PUS_PULL_UP_100K   = 2
	SW_PAD_CTL_PUS_PULL_UP_22K    = 3

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

	SW_MUX_CTL_SION = 4
	SW_MUX_CTL_MUX_MODE = 0
)

type Pad struct {
	// mux register (IOMUXC_SW_MUX_CTL_PAD_*)
	mux uint32
	// pad register (IOMUXC_SW_PAD_CTL_PAD_*)
	pad uint32
	// daisy register (IOMUXC_*_SELECT_INPUT)
	daisy uint32
}

// NewPad initializes a pad.
func NewPad(mux uint32, pad uint32, daisy uint32) (*Pad, error) {
	for _, r := range []uint32{mux, pad, daisy} {
		if !(r >= IOMUXC_START || r <= IOMUXC_END) {
			return nil, fmt.Errorf("invalid IOMUXC register %#x", r)
		}
	}

	return &Pad{
		mux:   mux,
		pad:   pad,
		daisy: daisy,
	}, nil
}

// Mode configures the pad iomux mode.
func (pad *Pad) Mode(mode uint32) {
	reg.SetN(pad.mux, SW_MUX_CTL_MUX_MODE, 0b1111, mode)
}

// SoftwareInput configures the pad SION bit.
func (pad *Pad) SoftwareInput(enabled bool) {
	if enabled {
		reg.Set(pad.mux, SW_MUX_CTL_SION)
	} else {
		reg.Clear(pad.mux, SW_MUX_CTL_SION)
	}
}

// Ctl configures the pad control register.
func (pad *Pad) Ctl(ctl uint32) {
	reg.Write(pad.pad, ctl)
}

// Select configures the pad daisy chain register.
func (pad *Pad) Select(input uint32) {
	reg.Write(pad.daisy, input)
}
