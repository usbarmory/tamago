// NuMaker-IIoT-NUC980G2 board support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package nuc980iiot provides hardware initialization, automatically on import,
// for the Nuvoton NuMaker-IIoT-NUC980G2 board (NUC980DK71YC, 128 MB DDR2).
//
// This package is only meant to be used with
// `GOOS=tamago GOARCH=arm GOARM=5` as supported by the TamaGo framework for
// bare metal Go, see https://github.com/usbarmory/tamago.
package nuc980iiot

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nuvoton/nuc980"
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	nuc980.Init()
}
