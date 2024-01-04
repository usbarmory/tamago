// NXP Temperature Monitor (TEMPMON) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bee implements a driver for the NXP Temperature Monitor (TEMPMON)
// adopting the following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package tempmon

import (
	"sync"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// TEMPMON registers
const (
	TEMPMON_TEMPSENSE0     = 0x00
	TEMPMON_TEMPSENSE0_SET = 0x04
	TEMPMON_TEMPSENSE0_CLR = 0x08

	TEMPSENSE0_TEMP_CNT     = 8
	TEMPSENSE0_FINISHED     = 2
	TEMPSENSE0_MEASURE_TEMP = 1
	TEMPSENSE0_POWER_DOWN   = 0

	TEMPMON_TEMPSENSE1     = 0x10
	TEMPMON_TEMPSENSE1_SET = 0x14
	TEMPMON_TEMPSENSE1_CLR = 0x18

	TEMPSENSE1_MEASURE_FREQ = 0
)

// TEMPMON represents the Temperature Monitor instance.
type TEMPMON struct {
	sync.Mutex

	// Base register
	Base uint32

	// control registers
	sense0     uint32
	sense0_set uint32
	sense0_clr uint32
	sense1     uint32
	sense1_clr uint32

	// calibration points
	hotTemp   uint32
	hotCount  uint32
	roomCount uint32
}

// Init initializes the Temperature Monitor instance, the calibration data is
// fused individually for each part and is required for correct measurements.
func (hw *TEMPMON) Init(calibrationData uint32) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		panic("invalid TEMPMON instance")
	}

	hw.sense0 = hw.Base + TEMPMON_TEMPSENSE0
	hw.sense0_set = hw.Base + TEMPMON_TEMPSENSE0_SET
	hw.sense0_clr = hw.Base + TEMPMON_TEMPSENSE0_CLR

	hw.sense1 = hw.Base + TEMPMON_TEMPSENSE1
	hw.sense1_clr = hw.Base + TEMPMON_TEMPSENSE1_CLR

	hw.hotTemp = bits.Get(&calibrationData, 0, 0xff)
	hw.hotCount = bits.Get(&calibrationData, 8, 0xfff)
	hw.roomCount = bits.Get(&calibrationData, 20, 0xfff)
}

// Read performs a single on-die temperature measurement.
func (hw *TEMPMON) Read() float32 {
	hw.Lock()
	defer hw.Unlock()

	if hw.sense0 == 0 {
		return 0
	}

	// enable sensor only during single measurement
	reg.Set(hw.sense0_clr, TEMPSENSE0_POWER_DOWN)
	defer reg.Set(hw.sense0_set, TEMPSENSE0_POWER_DOWN)

	// start and wait for a single measurement
	reg.SetN(hw.sense1_clr, TEMPSENSE1_MEASURE_FREQ, 0xffff, 0xffff)
	reg.Set(hw.sense0_set, TEMPSENSE0_MEASURE_TEMP)
	reg.Wait(hw.sense0, TEMPSENSE0_FINISHED, 1, 1)

	cnt := reg.Get(hw.sense0, TEMPSENSE0_TEMP_CNT, 0xfff)

	return temp(cnt, hw.hotTemp, hw.hotCount, hw.roomCount)
}

// p3531, 52.2 Software Usage Guidelines, IMX6ULLRM
func temp(cnt, hotTemp, hotCount, roomCount uint32) float32 {
	nm := float32(cnt)
	t1 := float32(25.0)
	t2 := float32(hotTemp)
	n1 := float32(roomCount)
	n2 := float32(hotCount)

	return t2 - (nm-n2)*((t2-t1)/(n1-n2))
}
