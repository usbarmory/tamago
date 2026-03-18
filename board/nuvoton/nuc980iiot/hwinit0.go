// NuMaker-IIoT-NUC980G2 board support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !arm.6

package nuc980iiot

import (
	_ "unsafe"
)

// Init0 takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
// ARM926EJ-S has no VFP unit; the default arm.Init would fault.
//
//go:linkname Init0 runtime/goos.Hwinit0
func Init0() {}
