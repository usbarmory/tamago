// AI Foundry Erbium initialization
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

// Soft Reset registers
const (
	SOFT_RESET = 0x02000028
	RESET_SOFT = 0
	RESET_WARM = 1
	RESET_MRAM = 2
)

// Reset asserts a system reset signal causing the processor to restart.
//
// Note that only the processor itself is guaranteed to restart as, depending
// on the board hardware layout, the system might remain powered (which might
// not be desirable). See respective board packages for cold reset options.
func Reset(soft bool) {
	if soft {
		reg.Set(SOFT_RESET, RESET_SOFT)
	} else {
		reg.Set(SOFT_RESET, RESET_WARM)
	}
}
