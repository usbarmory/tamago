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
	SIO_CFG                 = 0x10
	CFG_SIO_PORT_WIDTH      = 3
	CFG_SIO_PORT_WIDTH_MASK = 0x3
	CFG_SIO_AUTO_REPEAT     = 6

	SIO_CLOCK                 = 0x14
	CLOCK_SYS_CLK_PERIOD      = 0
	CLOCK_SYS_CLK_PERIOD_MASK = 0xff
	CLOCK_SIO_CLK_FREQ        = 8
	CLOCK_SIO_CLK_FREQ_MASK   = 0xfff

	SIO_PORT_CFG             = 0x18
	PORT_CFG_BIT_SOURCE      = 12
	PORT_CFG_BIT_SOURCE_MASK = 0x7
	BIT_SOURCE_FORCED_LOW    = 0
	BIT_SOURCE_FORCED_HIGH   = 1

	SIO_PORT_ENA = 0x98

	PORT_COUNT     = 32
	PORT_WIDTH_MAX = 4
)

// SGPIO represents a serial GPIO controller instance.
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
	hw.Lock()
	defer hw.Unlock()

	switch {
	case hw.Base == 0:
		return errors.New("invalid SGPIO controller instance")
	case hw.PortWidth < 1 || hw.PortWidth > PORT_WIDTH_MAX:
		return fmt.Errorf("invalid SGPIO port width %d", hw.PortWidth)
	case hw.ClockPeriod == 0 || hw.ClockPeriod > CLOCK_SYS_CLK_PERIOD_MASK:
		return fmt.Errorf("invalid SGPIO clock period %d", hw.ClockPeriod)
	case hw.ClockDivider == 0 || hw.ClockDivider > CLOCK_SIO_CLK_FREQ_MASK:
		return fmt.Errorf("invalid SGPIO clock divider %d", hw.ClockDivider)
	}

	var config uint32
	bits.SetN(&config, CFG_SIO_PORT_WIDTH, CFG_SIO_PORT_WIDTH_MASK, uint32(hw.PortWidth-1))
	bits.SetTo(&config, CFG_SIO_AUTO_REPEAT, hw.AutoRepeat)
	reg.Write(hw.Base+SIO_CFG, config)

	var clock uint32
	bits.SetN(&clock, CLOCK_SYS_CLK_PERIOD, CLOCK_SYS_CLK_PERIOD_MASK, hw.ClockPeriod)
	bits.SetN(&clock, CLOCK_SIO_CLK_FREQ, CLOCK_SIO_CLK_FREQ_MASK, hw.ClockDivider)
	reg.Write(hw.Base+SIO_CLOCK, clock)

	return
}

// EnablePorts enables every port selected by mask without disabling ports that
// are already active.
func (hw *SGPIO) EnablePorts(mask uint32) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		return errors.New("invalid SGPIO controller instance")
	}

	reg.Or(hw.Base+SIO_PORT_ENA, mask)

	return
}

// SetBit configures one port output as forced high or forced low.
func (hw *SGPIO) SetBit(port, bit int, high bool) (err error) {
	hw.Lock()
	defer hw.Unlock()

	switch {
	case hw.Base == 0:
		return errors.New("invalid SGPIO controller instance")
	case port < 0 || port >= PORT_COUNT:
		return fmt.Errorf("invalid SGPIO port %d", port)
	case hw.PortWidth < 1 || hw.PortWidth > PORT_WIDTH_MAX:
		return fmt.Errorf("invalid SGPIO port width %d", hw.PortWidth)
	case bit < 0 || bit >= hw.PortWidth:
		return fmt.Errorf("invalid SGPIO bit %d", bit)
	}

	source := uint32(BIT_SOURCE_FORCED_LOW)

	if high {
		source = BIT_SOURCE_FORCED_HIGH
	}

	addr := hw.Base + SIO_PORT_CFG + uint32(port*4)
	reg.SetN(addr, PORT_CFG_BIT_SOURCE+bit*3, PORT_CFG_BIT_SOURCE_MASK, source)

	return
}
