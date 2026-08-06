// Microchip Serial GPIO support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sgpio implements the Serial GPIO controller found on Microchip
// LAN969x SoCs under the following specification:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package sgpio

import (
	"errors"
	"fmt"
	"sync"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	sioConfig       = 0x10
	configWidth     = 3
	configRepeat    = 6
	sioClock        = 0x14
	clockPeriod     = 0
	clockPeriodMSK  = 0xff
	clockDivider    = 8
	clockDividerMSK = 0xfff
	sioPortConfig   = 0x18
	portStride      = 4
	bitSource       = 12
	bitSourceMSK    = 0x7
	sioPortEnable   = 0x98

	ports            = 32
	maximumPortWidth = 4
	forcedLow        = 0
	forcedHigh       = 1
)

// SGPIO represents a Microchip Serial GPIO controller instance.
type SGPIO struct {
	sync.Mutex

	// Base register of the SIO_CTRL group.
	Base uint32
	// PortWidth configures one to four serial GPIO bits per port.
	PortWidth int
	// ClockPeriod is the system clock period in units of 100 ps.
	ClockPeriod uint32
	// ClockDivider divides the system clock to produce the shift clock.
	ClockDivider uint32
	// AutoRepeat continuously emits configured output bursts when enabled.
	AutoRepeat bool
}

// Init configures the port width and shift clock. Pin routing is configured
// separately through the GPIO controller.
func (hw *SGPIO) Init() (err error) {
	switch {
	case hw.Base == 0:
		return errors.New("invalid SGPIO controller instance")
	case hw.PortWidth < 1 || hw.PortWidth > maximumPortWidth:
		return fmt.Errorf("invalid SGPIO port width %d", hw.PortWidth)
	case hw.ClockPeriod == 0 || hw.ClockPeriod > clockPeriodMSK:
		return fmt.Errorf("invalid SGPIO clock period %d", hw.ClockPeriod)
	case hw.ClockDivider == 0 || hw.ClockDivider > clockDividerMSK:
		return fmt.Errorf("invalid SGPIO clock divider %d", hw.ClockDivider)
	}

	hw.Lock()
	defer hw.Unlock()

	var config uint32
	bits.SetN(&config, configWidth, 0x3, uint32(hw.PortWidth-1))
	bits.SetTo(&config, configRepeat, hw.AutoRepeat)
	reg.Write(hw.Base+sioConfig, config)

	var clock uint32
	bits.SetN(&clock, clockPeriod, clockPeriodMSK, hw.ClockPeriod)
	bits.SetN(&clock, clockDivider, clockDividerMSK, hw.ClockDivider)
	reg.Write(hw.Base+sioClock, clock)

	return
}

// EnablePorts enables every port selected by mask without disabling ports that
// are already active.
func (hw *SGPIO) EnablePorts(mask uint32) (err error) {
	if hw.Base == 0 {
		return errors.New("invalid SGPIO controller instance")
	}

	hw.Lock()
	defer hw.Unlock()

	reg.Or(hw.Base+sioPortEnable, mask)

	return
}

// SetBit configures one port output as forced high or forced low.
func (hw *SGPIO) SetBit(port, bit int, high bool) (err error) {
	switch {
	case hw.Base == 0:
		return errors.New("invalid SGPIO controller instance")
	case port < 0 || port >= ports:
		return fmt.Errorf("invalid SGPIO port %d", port)
	case hw.PortWidth < 1 || hw.PortWidth > maximumPortWidth:
		return fmt.Errorf("invalid SGPIO port width %d", hw.PortWidth)
	case bit < 0 || bit >= hw.PortWidth:
		return fmt.Errorf("invalid SGPIO bit %d", bit)
	}

	source := uint32(forcedLow)
	if high {
		source = forcedHigh
	}

	hw.Lock()
	defer hw.Unlock()

	addr := hw.Base + sioPortConfig + uint32(port*portStride)
	reg.SetN(addr, bitSource+bit*3, bitSourceMSK, source)

	return
}
