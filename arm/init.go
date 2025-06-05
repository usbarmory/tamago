// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	_ "unsafe"
)

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
//go:linkname Init runtime.hwinit0
func Init() {
	if int(read_cpsr()&0x1f) != SYS_MODE {
		// initialization required only when in PL1
		return
	}

	vfp_enable()
}
