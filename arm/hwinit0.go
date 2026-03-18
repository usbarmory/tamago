// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !arm.6

package arm

import (
	_ "unsafe"
)

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
// On GOARM=5 (soft-float ABI) no VFP initialization is required at pre-World
// start; any board-level VFP enable happens later in Hwinit1 if needed.
//
//go:linkname Init runtime/goos.Hwinit0
func Init() {}
