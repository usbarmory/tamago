// MC146818A RTC driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package microvm

import (
	"errors"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// RTC registers
//
// (IBM PC AT Technical Reference - March 1984)
const (
	CMOS_RTC_OUT = 0x70
	CMOS_RTC_IN  = 0x71

	SECONDS = 0x00
	MINUTES = 0x02
	HOURS   = 0x04
	DOW     = 0x07
	MONTH   = 0x08
	YEAR    = 0x09
	CENTURY = 0x32

	STATUSA     = 0x0a
	STATUSA_UIP = 7
)

// RTC represents a Real-Time Clock instance.
type RTC struct {
	// Time zone
	Location *time.Location
}

func (rtc *RTC) read(addr uint8) int {
	reg.Out8(CMOS_RTC_OUT, addr)
	return int(reg.In8(CMOS_RTC_IN))
}

func bcdToBin(val int) int {
	return (val & 0x0f) + ((val / 16) * 10)
}

// Now() returns the real-time clock information.
func (rtc *RTC) Now() (t time.Time, err error) {
	if rtc.Location == nil {
		if rtc.Location, err = time.LoadLocation(""); err != nil {
			return
		}
	}

	if a := rtc.read(STATUSA); (a>>STATUSA_UIP)&1 == 1 {
		err = errors.New("update in progress")
		return
	}

	// We assume that the RTC remains in its initialized state with Data
	// Mode set to BCD and 24-hour mode.

	ss := bcdToBin(rtc.read(SECONDS))
	mm := bcdToBin(rtc.read(MINUTES))
	dd := bcdToBin(rtc.read(DOW))
	MM := bcdToBin(rtc.read(MONTH))
	yy := bcdToBin(rtc.read(YEAR))
	cc := bcdToBin(rtc.read(CENTURY))

	hh := rtc.read(HOURS)
	hh = ((hh & 0x0f) + (((hh & 0x70) / 16) * 10)) | (hh & 0x80)

	return time.Date(cc*100+yy, time.Month(MM), dd, hh, mm, ss, 0, rtc.Location), nil
}
