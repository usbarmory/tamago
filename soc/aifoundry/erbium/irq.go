// AI Foundry Erbium support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package erbium

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// Interrupt ESRs
const (
	IPI_TRIGGER       = 0x80f40090
	IPI_TRIGGER_CLEAR = 0x80f40098
)

// IPI sends an Inter-Processor Interrupt (IPI).
func IPI(hart int) {
	reg.Set64(IPI_TRIGGER, hart)
}

// ClearIPI clears an Inter-Processor Interrupt (IPI).
func ClearIPI(hart int) {
	reg.Set64(IPI_TRIGGER_CLEAR, hart)
}
