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
	GCB_SOFT_RST = GCB_BASE + 0x0c
	SOFT_SWC_RST = 1
)

// Reset asserts the switch core reset field causing the SoC to restart with a
// soft reset.
//
// Note that only the SoC itself is guaranteed to restart as, depending on the
// board hardware layout, the system might remain powered (which might not be
// desirable). See respective board packages for cold reset options.
func Reset() {
	reg.Set(GCB_SOFT_RST, SOFT_SWC_RST)
}
