// Raspberry Pi 1 support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pi1 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize
// +build !linkramsize

package pi1

import (
	_ "unsafe"
)

//go:linkname ramSize runtime.ramSize

// the following models have 256 MB RAM:
// - model A  (all)
// - model A+ (before 10th August 2016)
// - model B  (before 15th October 2012)
// var ramSize uint32 = 0x10000000 - 0x04000000 // 256 MB - 64MB GPU (VideoCore)

// the following models have 512 MB RAM:
// - model A+ (after 10th August 2016)
// - model B  (after 15th October 2012)
// - model B+ (all)
var ramSize uint32 = 0x20000000 - 0x04000000 // 512 MB - 64MB GPU (VideoCore)
