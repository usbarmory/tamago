// Flexible Serial Communication Controller (FLEXCOM)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package flexcom implements a driver for Flexible Serial Communication
// Controller (FLEXCOM) controllers, currently only USART mode is supported.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package flexcom

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// FLEXCOM registers
const (
	FLEX_MR  = 0x00
	MR_USART = 1

	FLEX_RHR = 0x10
	FLEX_THR = 0x20
)

// FLEXCOM represents a Flexible Serial Communication controller instance.
type FLEXCOM struct {
	// Controller index
	Index int
	// Base register
	Base uint32
}

// Init initializes and enables an FLEXCOM controller instance in USART mode.
func (hw *FLEXCOM) Init() {
	if hw.Base == 0 {
		panic("invalid FLEXCOM controller instance")
	}

	reg.Set(hw.Base+FLEX_MR, MR_USART)
}
