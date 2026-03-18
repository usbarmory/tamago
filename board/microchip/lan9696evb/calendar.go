// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan9696evb

import (
	"github.com/usbarmory/tamago/internal/reg"
)

const CalendarLength = 30

var TaxiCalendar [][]uint32 = [][]uint32{
	//  2G5: Port Group Manager IDs (0-9) separated by idle slots
	[]uint32{0, 10, 10, 1, 10, 10, 2, 10, 10, 3, 10, 10, 4, 10, 10, 5, 10, 10, 6, 10, 10, 7, 10, 10, 8, 10, 10, 9, 10, 10},
	//   5G: Port Group Manager IDs (0-7) separated by idle slots
	[]uint32{0, 10, 10, 1, 10, 10, 2, 10, 10, 3, 10, 10, 4, 10, 10, 5, 10, 10, 6, 10, 10, 7, 10, 10, 10, 10, 10, 10, 10, 10},
	[]uint32{0, 10, 10, 1, 10, 10, 2, 10, 10, 3, 10, 10, 4, 10, 10, 5, 10, 10, 6, 10, 10, 7, 10, 10, 10, 10, 10, 10, 10, 10},
	// <=1G: Port Group Manager IDs (0-1) separated by idle slots
	[]uint32{0, 10, 10, 1, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
	[]uint32{0, 10, 10, 1, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
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
