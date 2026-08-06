// Microchip LAN969x fan controller
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package fan implements the fan controller found on Microchip LAN969x SoCs
// under the following specification:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package fan

import (
	"errors"
	"sync"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	FAN_CFG              = 0x00
	CFG_FAN_STAT_CFG     = 0
	CFG_PWM_OPEN_COL_ENA = 1
	CFG_GATE_ENA         = 2
	CFG_INV_POL          = 3
	CFG_DUTY_CYCLE       = 16
	CFG_DUTY_CYCLE_MASK  = 0xff

	PWM_FREQ             = 0x04
	FREQ_PWM_FREQ        = 0
	FREQ_PWM_FREQ_MASK   = 0xffff
	FREQ_CLK_CYCLES_10US = 16
	FREQ_CLK_CYCLES_MASK = 0xfff

	FAN_CNT          = 0x08
	CNT_FAN_CNT      = 0
	CNT_FAN_CNT_MASK = 0xffff

	FAN_CLOCK_HZ          = 328125000
	FAN_CLOCK_CYCLES_10US = FAN_CLOCK_HZ / 100000
)

// CounterMode controls how the tachometer counter is updated.
type CounterMode uint8

const (
	// PulsesPerSecond updates the counter once per second with the number of
	// tachometer pulses received during the previous second.
	PulsesPerSecond CounterMode = iota
	// TotalPulses configures the counter as a wrapping total of tachometer
	// pulses.
	TotalPulses
)

// FAN represents a Microchip fan controller instance.
type FAN struct {
	sync.Mutex

	// Base register
	Base uint32
	// PWM must be configures with the output frequency in hertz.
	PWM uint32
	// Inverted configures the PWM output polarity.
	Inverted bool
	// OpenCollector configures the PWM pin as an open-collector output.
	OpenCollector bool
	// GateTacho counts tachometer pulses only while the PWM output is active.
	GateTacho bool
	// CounterMode selects how the tachometer counter is updated.
	CounterMode CounterMode
}

// Init initializes the fan controller with its output disabled. Pin routing is
// configured separately through the GPIO controller.
func (hw *FAN) Init() (err error) {
	if hw.Base == 0 {
		return errors.New("invalid fan controller instance")
	}

	if hw.PWM == 0 {
		return errors.New("invalid fan PWM frequency")
	}

	divider := FAN_CLOCK_HZ / hw.PWM / 256

	if divider == 0 || divider > FREQ_PWM_FREQ_MASK {
		return errors.New("invalid fan PWM frequency")
	}

	if hw.CounterMode > TotalPulses {
		return errors.New("invalid fan counter mode")
	}

	hw.Lock()
	defer hw.Unlock()

	var freq uint32
	bits.SetN(&freq, FREQ_PWM_FREQ, FREQ_PWM_FREQ_MASK, divider)
	bits.SetN(&freq, FREQ_CLK_CYCLES_10US, FREQ_CLK_CYCLES_MASK, FAN_CLOCK_CYCLES_10US)
	reg.Write(hw.Base+PWM_FREQ, freq)

	var cfg uint32
	bits.SetTo(&cfg, CFG_INV_POL, hw.Inverted)
	bits.SetTo(&cfg, CFG_PWM_OPEN_COL_ENA, hw.OpenCollector)
	bits.SetTo(&cfg, CFG_GATE_ENA, hw.GateTacho)
	bits.SetTo(&cfg, CFG_FAN_STAT_CFG, hw.CounterMode == TotalPulses)
	reg.Write(hw.Base+FAN_CFG, cfg)

	return
}

// SetDuty sets the PWM duty cycle, where 0 is always off and 255 is always on.
func (hw *FAN) SetDuty(duty uint8) {
	reg.SetN(hw.Base+FAN_CFG, CFG_DUTY_CYCLE, CFG_DUTY_CYCLE_MASK, uint32(duty))
}

// Duty returns the configured PWM duty cycle.
func (hw *FAN) Duty() (duty uint8) {
	if hw.Base == 0 {
		return
	}

	return uint8(reg.GetN(hw.Base+FAN_CFG, CFG_DUTY_CYCLE, CFG_DUTY_CYCLE_MASK))
}

// Frequency returns the configured PWM frequency.
func (hw *FAN) Frequency() (freq uint32) {
	if hw.Base == 0 {
		return
	}

	divider := reg.GetN(hw.Base+PWM_FREQ, FREQ_PWM_FREQ, FREQ_PWM_FREQ_MASK)

	if divider == 0 {
		return
	}

	return FAN_CLOCK_HZ / divider / 256
}

// TachoCount returns the raw 16-bit tachometer count.
func (hw *FAN) TachoCount() (count uint16) {
	if hw.Base == 0 {
		return
	}

	return uint16(reg.GetN(hw.Base+FAN_CNT, CNT_FAN_CNT, CNT_FAN_CNT_MASK))
}
