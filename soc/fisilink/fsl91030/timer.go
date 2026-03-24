// Fisilink FSL91030 timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

// Counter returns the number of nanoseconds counted from the RTCCLK input.
// It reads the hardware CLINT mtime register and is always available
// regardless of the linknanotime build tag.
func Counter() uint64 {
	return uint64(CLINT.Nanotime())
}
