// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package arm

import (
	_ "unsafe"
)

// defined in vfp.s
func vfp_enable()

//go:linkname EnableVFP runtime.vfp_enable
func (c *CPU) EnableVFP() {
	vfp_enable()
}
