// Fisilink FSL91030 timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

import (
	_ "unsafe"
)

// Counter returns the number of nanoseconds counted from the RTCCLK input.
func Counter() uint64 {
	return uint64(CLINT.Nanotime())
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return RV64.GetTime()
}
