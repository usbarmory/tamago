// Microchip LAN969x temperature sensor
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import "github.com/usbarmory/tamago/internal/reg"

// Temperature registers
const (
	TEMP_SENSOR_CFG  = CHIP_TOP_BASE + 0x104
	TEMP_SENSOR_STAT = CHIP_TOP_BASE + 0x108
)

// Configuration fields
const (
	CFG_SAMPLE_ENA      = 0
	CFG_CONTINUOUS_MODE = 1
	CFG_CLK_CYCLES_1US  = 15
)

// Status fields
const (
	STAT_TEMP       = 0
	STAT_TEMP_VALID = 12
)

// TEMP_CLK_CYCLES_1US matches the 328.125 MHz system clock.
// p501, 3.47.3 CLOCKING AND RESET, Microchip DS00005048E
const TEMP_CLK_CYCLES_1US = 328

type temp struct{}

// Init initializes and enables continuous temperature sampling.
func (*temp) Init() {
	reg.SetN(TEMP_SENSOR_CFG, CFG_CLK_CYCLES_1US, 0x1ff,
		TEMP_CLK_CYCLES_1US)
	reg.Set(TEMP_SENSOR_CFG, CFG_CONTINUOUS_MODE)
	reg.Set(TEMP_SENSOR_CFG, CFG_SAMPLE_ENA)
}

// Read returns the latest temperature sample in degrees Celsius and whether
// it is valid.
func (*temp) Read() (celsius float32, valid bool) {
	if !reg.Get(TEMP_SENSOR_STAT, STAT_TEMP_VALID) {
		return
	}

	raw := reg.GetN(TEMP_SENSOR_STAT, STAT_TEMP, 0xfff)

	// p768, 3.47.11.17 Temperature Sensor, Microchip DS00005048E
	celsius = float32(raw)/4096*352.3 - 109.4
	valid = true

	return
}
