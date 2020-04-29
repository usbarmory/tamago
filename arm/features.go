// ARM processor features
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

type features struct {
	// instruction sets
	arm     bool
	thumb   bool
	jazelle bool
	thumbee bool

	// extensions
	programmersModel bool
	security         bool
	mProfileModel    bool
	virtualization   bool
	genericTimer     bool
}

// defined in arm.s
func read_idpfr0() uint32
func read_idpfr1() uint32

func (f *features) init() {
	var idpfr0 uint32
	var idpfr1 uint32

	idpfr0 = read_idpfr0()
	idpfr1 = read_idpfr1()

	f.arm     = (idpfr0 & IDPFR0_ARM_MASK) != 0
	f.thumb   = (idpfr0 & IDPFR0_THUMB_MASK) != 0
	f.thumbee = (idpfr0 & IDPFR0_THUMBEE_MASK) != 0
	f.jazelle = (idpfr0 & IDPFR0_JAZELLE_MASK) != 0

	f.programmersModel = (idpfr1 & IDPFR1_PROGRAMMERS_MODEL_MASK) != 0
	f.security         = (idpfr1 & IDPFR1_SECURITY_MASK) != 0
	f.mProfileModel    = (idpfr1 & IDPFR1_M_PROFILE_MODEL_MASK) != 0
	f.virtualization   = (idpfr1 & IDPFR1_VIRTUALIZATION_MASK) != 0
	f.genericTimer     = (idpfr1 & IDPFR1_GENERIC_TIMER_MASK) != 0
}

func (f *features) print() {
	if f.arm {
		print("ARM instruction set implemented.\n")
	}
	if f.thumb {
		print("Thumb instruction set implemented.\n")
	}
	if f.jazelle {
		print("Jazelle extension implemented.\n")
	}
	if f.thumbee {
		print("ThumbEE extension implemented.\n")
	}
	if f.programmersModel {
		print("Programmers' model support.\n")
	}
	if f.security {
		print("Security Extensions implemented.\n")
	}
	if f.mProfileModel {
		print("M Profile programmers model implemented.\n")
	}
	if f.virtualization {
		print("Virtualization Extensions implemented.\n")
	}
	if f.genericTimer {
		print("Generic Timer Extensions implemented.\n")
	}
}
