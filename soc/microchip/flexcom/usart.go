// Flexible Serial Communication Controller (FLEXCOM)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package flexcom

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

// USART registers
const (
	DEFAULT_BAUDRATE = 115200

	FLEX_US_CR = 0x200
	FLEX_US_MR = 0x204
)

// Tx transmits a single character to the serial port.
func (hw *FLEXCOM) Tx(c byte) {
	reg.Write(hw.Base+FLEX_THR, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *FLEXCOM) Rx() (c byte, valid bool) {
	return byte(reg.Read(hw.Base + FLEX_RHR)), true
}

// Write data from buffer to serial port.
func (hw *FLEXCOM) Write(buf []byte) (n int, _ error) {
	for n = 0; n < len(buf); n++ {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *FLEXCOM) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n = 0; n < len(buf); n++ {
		buf[n], valid = hw.Rx()

		if !valid {
			if n == 0 {
				runtime.Gosched()
			}

			break
		}
	}

	return
}
