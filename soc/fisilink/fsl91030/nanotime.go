// Fisilink FSL91030 runtime nanotime hook
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// When built without the linknanotime tag this file provides the default
// runtime/goos.Nanotime implementation, reading the CLINT mtime register.
// Building with -tags linknanotime excludes it so a board package can supply
// its own time source (see board/nuclei/eval_soc).
//
//go:build !linknanotime

package fsl91030

import (
	_ "unsafe"
)

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return RV64.GetTime()
}
