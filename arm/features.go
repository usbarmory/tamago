// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// ARM processor feature registers
const (
	ID_PFR0_ARM_MASK     = 0x000f
	ID_PFR0_THUMB_MASK   = 0x00f0
	ID_PFR0_THUMBEE_MASK = 0x0f00
	ID_PFR0_JAZELLE_MASK = 0xf000

	ID_PFR1_PROGRAMMERS_MODEL_MASK = 0x0000f
	ID_PFR1_SECURITY_MASK          = 0x000f0
	ID_PFR1_M_PROFILE_MODEL_MASK   = 0x00f00
	ID_PFR1_VIRTUALIZATION_MASK    = 0x0f000
	ID_PFR1_GENERIC_TIMER_MASK     = 0xf0000
)

// defined in features.s
func read_idpfr0() uint32
func read_idpfr1() uint32

func (cpu *CPU) initFeatures() {
	idpfr0 := read_idpfr0()
	idpfr1 := read_idpfr1()

	cpu.arm = (idpfr0 & ID_PFR0_ARM_MASK) != 0
	cpu.thumb = (idpfr0 & ID_PFR0_THUMB_MASK) != 0
	cpu.thumbee = (idpfr0 & ID_PFR0_THUMBEE_MASK) != 0
	cpu.jazelle = (idpfr0 & ID_PFR0_JAZELLE_MASK) != 0

	cpu.programmersModel = (idpfr1 & ID_PFR1_PROGRAMMERS_MODEL_MASK) != 0
	cpu.security = (idpfr1 & ID_PFR1_SECURITY_MASK) != 0
	cpu.mProfileModel = (idpfr1 & ID_PFR1_M_PROFILE_MODEL_MASK) != 0
	cpu.virtualization = (idpfr1 & ID_PFR1_VIRTUALIZATION_MASK) != 0
	cpu.genericTimer = (idpfr1 & ID_PFR1_GENERIC_TIMER_MASK) != 0
}
