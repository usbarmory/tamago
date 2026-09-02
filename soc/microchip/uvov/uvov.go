// Microchip under-voltage and over-voltage monitor
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uvov implements read-only diagnostics for the Microchip
// under-voltage and over-voltage monitor integrated in LAN969x SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package uvov

import (
	"errors"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// Register definitions follow Microchip's BSD-3-Clause LAN969x support at
// https://github.com/microchip-ung/arm-trusted-firmware/tree/master/plat/microchip/lan969x.
const (
	UVOV_CTL         = 0x00
	CTL_OVR_DET_EN   = 1
	CTL_UVR_DET_EN   = 0
	UVOV_INT_STS     = 0x04
	UVOV_INT_EN      = 0x08
	UVOV_TUNE        = 0x0c
	TUNE_CDR         = 24
	TUNE_M           = 18
	TUNE_MAG         = 12
	TUNE_TRIM_EN     = 6
	TRIM_INTERRUPT   = 24
	UVOV_CFG0_V09    = 0x14
	UVOV_CFG1_V09    = 0x18
	UVOV_CFG0_V12    = 0x1c
	UVOV_CFG1_V12    = 0x20
	UVOV_CFG0_V18    = 0x24
	UVOV_CFG1_V18    = 0x28
	CFG0_OV_DEB_EN   = 28
	CFG0_UV_DEB_EN   = 27
	CFG0_OV_RST_EN   = 11
	CFG0_UV_RST_EN   = 10
	CFG0_OVR_RNG_SEL = 4
	CFG0_UVR_RNG_SEL = 2
	CFG0_OVR_DET_EN  = 1
	CFG0_UVR_DET_EN  = 0
	CFG1_DEBOUNCE    = 8
	CFG1_OV_STS      = 1
	CFG1_UV_STS      = 0
)

// Rail identifies one monitored nominal supply.
type Rail uint8

const (
	// V09 identifies the nominal 0.9 V supply.
	V09 Rail = iota
	// V12 identifies the nominal 1.2 V supply.
	V12
	// V18 identifies the nominal 1.8 V supply.
	V18
	railCount
)

// TrimSnapshot reports the applied and observed factory-trim fields.
type TrimSnapshot struct {
	Raw               uint32
	CDR               uint8
	Magnitude         uint8
	ObservedMagnitude uint8
	Active            bool
	Event             bool
	InterruptEnabled  bool
}

// RailSnapshot reports one detector's configuration and current state.
type RailSnapshot struct {
	Configuration         uint32
	DebounceConfiguration uint32
	UnderEnabled          bool
	OverEnabled           bool
	UnderDebounced        bool
	OverDebounced         bool
	UnderInterruptEnabled bool
	OverInterruptEnabled  bool
	UnderResetEnabled     bool
	OverResetEnabled      bool
	UnderRange            uint8
	OverRange             uint8
	UnderActive           bool
	OverActive            bool
	UnderEvent            bool
	OverEvent             bool
	Debounce              uint32
}

// Snapshot reports the complete read-only UVOV state used by diagnostics.
type Snapshot struct {
	Control         uint32
	InterruptStatus uint32
	InterruptEnable uint32
	UnderEnabled    bool
	OverEnabled     bool
	Trim            TrimSnapshot
	Rails           [railCount]RailSnapshot
}

// LiveFault reports whether a detector currently reports an invalid voltage.
func (snapshot Snapshot) LiveFault() bool {
	for _, rail := range snapshot.Rails {
		if rail.UnderActive || rail.OverActive {
			return true
		}
	}

	return false
}

// EventPending reports whether the interrupt status records a voltage event.
func (snapshot Snapshot) EventPending() bool {
	if snapshot.Trim.Event {
		return true
	}
	for _, rail := range snapshot.Rails {
		if rail.UnderEvent || rail.OverEvent {
			return true
		}
	}

	return false
}

// ChannelResetEnabled reports whether any detector channel has its local
// system-reset capability enabled.
func (snapshot Snapshot) ChannelResetEnabled() bool {
	for _, rail := range snapshot.Rails {
		if rail.UnderResetEnabled || rail.OverResetEnabled {
			return true
		}
	}

	return false
}

// UVOV represents one under-voltage and over-voltage monitor instance.
type UVOV struct {
	// Base is the register base.
	Base uint32
}

var railLayout = [railCount]struct {
	configuration         uint32
	debounceConfiguration uint32
	underEvent            int
	overEvent             int
}{
	V09: {UVOV_CFG0_V09, UVOV_CFG1_V09, 16, 20},
	V12: {UVOV_CFG0_V12, UVOV_CFG1_V12, 8, 12},
	V18: {UVOV_CFG0_V18, UVOV_CFG1_V18, 0, 4},
}

// Snapshot reads the complete monitor configuration and status without
// changing detector, interrupt, or reset state.
func (hw *UVOV) Snapshot() (snapshot Snapshot, err error) {
	if hw.Base == 0 {
		return snapshot, errors.New("invalid UVOV instance")
	}

	snapshot.Control = reg.Read(hw.Base + UVOV_CTL)
	snapshot.InterruptStatus = reg.Read(hw.Base + UVOV_INT_STS)
	snapshot.InterruptEnable = reg.Read(hw.Base + UVOV_INT_EN)
	snapshot.UnderEnabled = bits.Get(&snapshot.Control, CTL_UVR_DET_EN)
	snapshot.OverEnabled = bits.Get(&snapshot.Control, CTL_OVR_DET_EN)

	tune := reg.Read(hw.Base + UVOV_TUNE)
	snapshot.Trim = TrimSnapshot{
		Raw:               tune,
		CDR:               uint8(bits.GetN(&tune, TUNE_CDR, 0x7)),
		Magnitude:         uint8(bits.GetN(&tune, TUNE_M, 0x3f)),
		ObservedMagnitude: uint8(bits.GetN(&tune, TUNE_MAG, 0x3f)),
		Active:            bits.Get(&tune, TUNE_TRIM_EN),
		Event:             bits.Get(&snapshot.InterruptStatus, TRIM_INTERRUPT),
		InterruptEnabled:  bits.Get(&snapshot.InterruptEnable, TRIM_INTERRUPT),
	}

	for rail, layout := range railLayout {
		configuration := reg.Read(hw.Base + layout.configuration)
		debounce := reg.Read(hw.Base + layout.debounceConfiguration)
		snapshot.Rails[rail] = RailSnapshot{
			Configuration:         configuration,
			DebounceConfiguration: debounce,
			UnderEnabled:          bits.Get(&configuration, CFG0_UVR_DET_EN),
			OverEnabled:           bits.Get(&configuration, CFG0_OVR_DET_EN),
			UnderDebounced:        bits.Get(&configuration, CFG0_UV_DEB_EN),
			OverDebounced:         bits.Get(&configuration, CFG0_OV_DEB_EN),
			UnderInterruptEnabled: bits.Get(&snapshot.InterruptEnable, layout.underEvent),
			OverInterruptEnabled:  bits.Get(&snapshot.InterruptEnable, layout.overEvent),
			UnderResetEnabled:     bits.Get(&configuration, CFG0_UV_RST_EN),
			OverResetEnabled:      bits.Get(&configuration, CFG0_OV_RST_EN),
			UnderRange:            uint8(bits.GetN(&configuration, CFG0_UVR_RNG_SEL, 0x3)),
			OverRange:             uint8(bits.GetN(&configuration, CFG0_OVR_RNG_SEL, 0x3)),
			UnderActive:           bits.Get(&debounce, CFG1_UV_STS),
			OverActive:            bits.Get(&debounce, CFG1_OV_STS),
			UnderEvent:            bits.Get(&snapshot.InterruptStatus, layout.underEvent),
			OverEvent:             bits.Get(&snapshot.InterruptStatus, layout.overEvent),
			Debounce:              bits.GetN(&debounce, CFG1_DEBOUNCE, 0xffffff),
		}
	}

	return
}
