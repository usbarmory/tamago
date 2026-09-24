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

// Outbound interrupt controller registers
const (
	INTR              = 0x128
	INTR_STICKY_BASE  = INTR + 0x40
	INTR_ENA_BASE     = INTR + 0x60
	INTR_ENA_CLR_BASE = INTR + 0x70
	INTR_ENA_SET_BASE = INTR + 0x80
)

// EnableInterrupt enables an outbound interrupt controller source.
func EnableInterrupt(id int) {
	group := id / 32
	index := id % 32
	reg.Write(CPU_BASE+INTR_ENA_SET_BASE+uint32(group*4), 1<<index)
}

// DisableInterrupt disables an outbound interrupt controller source.
func DisableInterrupt(id int) {
	group := id / 32
	index := id % 32
	reg.Write(CPU_BASE+INTR_ENA_CLR_BASE+uint32(group*4), 1<<index)
}

// ClearInterrupt clears the sticky event of an outbound interrupt controller
// source.
func ClearInterrupt(id int) {
	group := id / 32
	index := id % 32
	reg.Write(CPU_BASE+INTR_STICKY_BASE+uint32(group*4), 1<<index)
}
