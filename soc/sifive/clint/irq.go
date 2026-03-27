// SiFive Core-Local Interruptor (CLINT) driver
// https://github.com/usbarmory/tamago
//
// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in Go LICENSE file.

package clint

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// IPI sends an Inter-Processor Interrupt (IPI).
func (hw *CLINT) IPI(hart int) {
	off := MSIP + uint32(hart * 4)
	reg.Set(uint32(hw.Base) + off, 0)
}

// ClearIPI clears an Inter-Processor Interrupt (IPI).
func (hw *CLINT) ClearIPI(hart int) {
	off := MSIP + uint32(hart * 4)
	reg.Clear(uint32(hw.Base) + off, 0)
}
