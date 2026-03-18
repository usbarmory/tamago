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

const (
	// Memory manager status
	HSCH_MMGT = HSCH_BASE + 0x8da4
	RESET_CFG = HSCH_MMGT + 0x8
)

// EnableSwitchCore resets and initializes the switching core.
func EnableSwitchCore() {
	// enable switch core
	reg.Write(RESET_CFG, 1)
}
