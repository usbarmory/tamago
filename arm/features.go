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

const (
	IDPFR0_ARM_MASK     uint32 = 0x000f
	IDPFR0_THUMB_MASK          = 0x00f0
	IDPFR0_THUMBEE_MASK        = 0x0f00
	IDPFR0_JAZELLE_MASK        = 0xf000

	IDPFR1_PROGRAMMERS_MODEL_MASK uint32 = 0x0000f
	IDPFR1_SECURITY_MASK                 = 0x000f0
	IDPFR1_M_PROFILE_MODEL_MASK          = 0x00f00
	IDPFR1_VIRTUALIZATION_MASK           = 0x0f000
	IDPFR1_GENERIC_TIMER_MASK            = 0xf0000
)

// defined in arm.s
func read_idpfr0() uint32
func read_idpfr1() uint32

func (c *CPU) initFeatures() {
	idpfr0 := read_idpfr0()
	idpfr1 := read_idpfr1()

	c.arm = (idpfr0 & IDPFR0_ARM_MASK) != 0
	c.thumb = (idpfr0 & IDPFR0_THUMB_MASK) != 0
	c.thumbee = (idpfr0 & IDPFR0_THUMBEE_MASK) != 0
	c.jazelle = (idpfr0 & IDPFR0_JAZELLE_MASK) != 0

	c.programmersModel = (idpfr1 & IDPFR1_PROGRAMMERS_MODEL_MASK) != 0
	c.security = (idpfr1 & IDPFR1_SECURITY_MASK) != 0
	c.mProfileModel = (idpfr1 & IDPFR1_M_PROFILE_MODEL_MASK) != 0
	c.virtualization = (idpfr1 & IDPFR1_VIRTUALIZATION_MASK) != 0
	c.genericTimer = (idpfr1 & IDPFR1_GENERIC_TIMER_MASK) != 0
}
