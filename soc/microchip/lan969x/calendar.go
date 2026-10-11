// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// Calendar registers
const (
	QSYS_RAM_INIT = QSYS_BASE + 0x89c
	QSYS_CALCFG   = QSYS_BASE + 0x874

	QSYS_CAL_AUTO0 = QSYS_CALCFG + 0x00
	QSYS_CAL_AUTO1 = QSYS_CALCFG + 0x04
	QSYS_CAL_AUTO2 = QSYS_CALCFG + 0x08
	QSYS_CAL_AUTO3 = QSYS_CALCFG + 0x0c
	GBPS_1         = 1

	QSYS_CAL_CTRL     = QSYS_CALCFG + 0x24
	CAL_CTRL_CAL_MODE = 11
	MODE_HALT         = 10
	MODE_CAL_AUTO     = 8

	DSM_CFG      = DSM_BASE + 0x14
	TAXI_CAL_CFG = DSM_CFG + 0xc98
	CAL_SEL_STAT = 23
	CAL_SWITCH   = 22
	CAL_PGM_SEL  = 21
	CAL_IDX      = 15
	CAL_PGM_VAL  = 1
	CAL_PGM_ENA  = 0
)

const CalendarLength = 30

var TaxiCalendar [][]uint32 = [][]uint32{
	//  2G5: Port Group Manager IDs (0-9) separated by idle slots
	{0, 10, 10, 1, 10, 10, 2, 10, 10, 3, 10, 10, 4, 10, 10, 5, 10, 10, 6, 10, 10, 7, 10, 10, 8, 10, 10, 9, 10, 10},
	//   5G: Port Group Manager IDs (0-7) separated by idle slots
	{0, 10, 10, 1, 10, 10, 2, 10, 10, 3, 10, 10, 4, 10, 10, 5, 10, 10, 6, 10, 10, 7, 10, 10, 10, 10, 10, 10, 10, 10},
	{0, 10, 10, 1, 10, 10, 2, 10, 10, 3, 10, 10, 4, 10, 10, 5, 10, 10, 6, 10, 10, 7, 10, 10, 10, 10, 10, 10, 10, 10},
	// <=1G: Port Group Manager IDs (0-1) separated by idle slots
	{0, 10, 10, 1, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
	{0, 10, 10, 1, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
}

func calendarConfig(taxi uint32) {
	taxi_cal_cfg := TAXI_CAL_CFG + uint32(taxi*4)

	// select inactive calendar for update
	active := reg.Get(taxi_cal_cfg, CAL_SEL_STAT)
	reg.SetTo(taxi_cal_cfg, CAL_PGM_SEL, !active)
	reg.Set(taxi_cal_cfg, CAL_PGM_ENA)

	// initialize calendar entries
	for i := range uint32(CalendarLength) {
		reg.SetN(taxi_cal_cfg, CAL_IDX, 0x3f, i)
		reg.SetN(taxi_cal_cfg, CAL_PGM_VAL, 0xf, TaxiCalendar[taxi][i])
	}

	// disable update of selected calendar
	reg.Clear(taxi_cal_cfg, CAL_PGM_ENA)

	// switch calendar
	reg.Set(taxi_cal_cfg, CAL_SWITCH)
}
