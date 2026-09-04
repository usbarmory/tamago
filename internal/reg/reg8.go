// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

import (
	"runtime"
	"time"
)

// defined in reg_*.s
func Read8(addr uint32) uint8
func Write8(addr uint32, val uint8)

func GetN8(addr uint32, pos int, mask int) uint8 {
	val := Read8(addr)
	return (val >> pos) & uint8(mask)
}

func WaitFor8(timeout time.Duration, addr uint32, pos int, mask int, val uint8) bool {
	start := time.Now()

	for GetN8(addr, pos, mask) != val {
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}
