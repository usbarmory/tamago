// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// defined in vfp.s
func vfp_enable()

// EnableVFP activates the ARM Vector-Floating-Point co-processor.
func (cpu *CPU) EnableVFP() {
	vfp_enable()
}
