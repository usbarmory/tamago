// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
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
func read_midr() uint32
func read_idpfr0() uint32
func read_idpfr1() uint32

// MIDR returns the Main ID Register (CP15 c0, c0, 0), which identifies the
// processor implementer, architecture, part number and revision.
func (cpu *CPU) MIDR() uint32 {
	return read_midr()
}

func (cpu *CPU) initFeatures() {
	// Do not set read features on cores with fixed exception vectors (e.g.
	// ARMv5) which have no CP15 feature ID registers.
	if cpu.vbar == 0 {
		return
	}

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
