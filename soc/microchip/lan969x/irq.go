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

// Interrupt controller registers
const (
	INTR          = 0x128
	INTR_ENA_BASE = INTR + 0x60
)

// EnableInterrupt enables propagation of an individual interrupt source.
func EnableInterrupt(id int) {
	group := id / 32
	index := id % 32
	reg.Set(CPU_BASE+INTR_ENA_BASE+uint32(group*4), index)
}

// DisableInterrupt disables propagation of an individual interrupt source.
func DisableInterrupt(id int) {
	group := id / 32
	index := id % 32
	reg.Clear(CPU_BASE+INTR_ENA_BASE+uint32(group*4), index)
}
