// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	_ "unsafe"
)

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
//go:linkname Init runtime.hwinit0
func Init() {}
