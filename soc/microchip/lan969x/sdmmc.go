// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

const (
	gckConfigOffset = 0xb4
	sdmmc0ClockID   = 2

	sdmmc0ParentClock = 1_000_000_000
	sdmmc0TargetClock = 200_000_000
)
