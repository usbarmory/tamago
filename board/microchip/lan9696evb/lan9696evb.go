// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package lan9696evb provides hardware initialization, automatically on
// import, for the Microchip LAN969x 24-port EVB (ev23x71a).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package lan9696evb

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/microchip/lan969x"
)

// Peripheral instances
var (
	FLEXCOM0 = lan969x.FLEXCOM0
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	lan969x.Init()

	// initialize console
	lan969x.FLEXCOM0.Init()
}

func init() {
	// initialize switch memories, calendar, VLANs and core
	initializeSwitchCore()
}
