// SpacemiT K1 runtime nanotime hook
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// When built without the linknanotime tag this file provides the default
// runtime/goos.Nanotime implementation, reading the RISC-V `time` CSR (see
// [Counter]). Building with -tags linknanotime excludes it so that a board
// package can supply its own time source.
//
//go:build !linknanotime

package k1

import (
	_ "unsafe"
)

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return RV64.GetTime()
}
