// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

const (
	FPEXC_EN = 30
)

// defined in vfp.s
func vfp_enable()

// EnableVFP activates the ARM Vector-Floating-Point co-processor.
func (cpu *CPU) EnableVFP() {
	vfp_enable()
}
