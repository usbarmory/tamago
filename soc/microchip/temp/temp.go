// Microchip LAN969x temperature sensor
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package temp implements a driver for Microchip Temperature sensors adopting
// the following specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package temp

import (
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// Temperature sensor registers
const (
	TEMP_SENSOR_CFG     = 0x04
	CFG_CLK_CYCLES_1US  = 15
	CFG_CONTINUOUS_MODE = 1
	CFG_SAMPLE_ENA      = 0

	TEMP_SENSOR_STAT = 0x08
	STAT_TEMP_VALID  = 12
	STAT_TEMP        = 0
)

// TEMP_CLK_CYCLES_1US matches the 328.125 MHz system clock.
// p501, 3.47.3 CLOCKING AND RESET, Microchip DS00005048E
const TEMP_CLK_CYCLES_1US = 328

// SENSOR represents the Temperature sensor instance.
type SENSOR struct {
	sync.Mutex

	// Base register
	Base uint32

	run bool
}

// Init initializes and enables continuous temperature sampling.
func (hw *SENSOR) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		panic("invalid SENSOR instance")
	}

	reg.SetN(hw.Base+TEMP_SENSOR_CFG, CFG_CLK_CYCLES_1US, 0x1ff, TEMP_CLK_CYCLES_1US)
	reg.Set(hw.Base+TEMP_SENSOR_CFG, CFG_CONTINUOUS_MODE)
	reg.Set(hw.Base+TEMP_SENSOR_CFG, CFG_SAMPLE_ENA)

	hw.run = true
}

// Read returns the latest temperature sample in degrees Celsius and whether it
// is valid.
func (hw *SENSOR) Read() (celsius float32, valid bool) {
	hw.Lock()
	defer hw.Unlock()

	if !hw.run || !reg.Get(hw.Base+TEMP_SENSOR_STAT, STAT_TEMP_VALID) {
		return
	}

	raw := reg.GetN(hw.Base+TEMP_SENSOR_STAT, STAT_TEMP, 0xfff)

	// p768, 3.47.11.17 Temperature Sensor, Microchip DS00005048E
	celsius = float32(raw)/4096*352.3 - 109.4
	valid = true

	return
}
