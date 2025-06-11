// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"github.com/usbarmory/tamago/bits"
)

// CPUID function numbers
//
// (Intel® Architecture Instruction Set Extensions
// and Future Features Programming Reference
// 1.5 CPUID INSTRUCTION).
const (
	CPUID_VENDOR           = 0x00
	CPUID_VENDOR_ECX_INTEL = 0x6c65746e // GenuineI(ntel)
	CPUID_VENDOR_ECX_AMD   = 0x444d4163 // Authenti(cAMD)

	CPUID_INTEL_CACHE = 0x04

	CPUID_INTEL_APIC = 0x0b
	INTEL_APIC_LP    = 0

	CPUID_TSC_CCC = 0x15
	CPUID_CPU_FRQ = 0x16

	CPUID_APM         = 0x80000007
	APM_TSC_INVARIANT = 8
)

// CPUID function numbers
//
// (AMD64 Architecture Programmer’s Manual
// Volume 3 - Appendix E.4 Extended Feature Function Numbers.
const (
	CPUID_AMD_PROC = 0x80000008
	AMD_PROC_NC    = 0
)

// KVM CPUID function numbers
//
// (https://docs.kernel.org/virt/kvm/x86/cpuid.html)
const (
	KVM_CPUID_SIGNATURE = 0x40000000
	KVM_SIGNATURE       = 0x4b4d564b // "KVMK"

	KVM_CPUID_FEATURES    = 0x40000001
	FEATURES_CLOCKSOURCE  = 0
	FEATURES_CLOCKSOURCE2 = 3

	KVM_CPUID_TSC_KHZ = 0x40000010
)

// AMD MSRs
const (
	MSR_AMD_PSTATE = 0xc0010064
)

// kvmclock MSRs
const (
	MSR_KVM_SYSTEM_TIME     = 0x12
	MSR_KVM_SYSTEM_TIME_NEW = 0x4b564d01
)

// defined in features.s
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

// CPUID returns the processor capabilities.
func (cpu *CPU) CPUID(leaf, subleaf uint32) (eax, ebx, ecx, edx uint32) {
	return cpuid(leaf, subleaf)
}

func (cpu *CPU) initFeatures() {
	_, _, _, apmFeatures := cpuid(CPUID_APM, 0)
	cpu.invariant = bits.IsSet(&apmFeatures, APM_TSC_INVARIANT)

	_, kvmk, _, _ := cpuid(KVM_CPUID_SIGNATURE, 0)
	cpu.kvm = kvmk == KVM_SIGNATURE

	if !cpu.kvm {
		return
	}

	kvmFeatures, _, _, _ := cpuid(KVM_CPUID_FEATURES, 0)

	if bits.IsSet(&kvmFeatures, FEATURES_CLOCKSOURCE) {
		cpu.kvmclock = 0x12
	}

	if bits.IsSet(&kvmFeatures, FEATURES_CLOCKSOURCE2) {
		cpu.kvmclock = 0x4b564d01
	}
}

// Features represents the processor capabilities detected through the CPUID
// instruction.
type Features struct {
	// InvariantTSC indicates whether the Time Stamp Counter is guaranteed
	// to be at constant rate.
	InvariantTSC bool
	// KVM indicates whether a Kernel-base Virtual Machine is detected.
	KVM bool
	// KVMClockMSR returns the kvmclock Model Specific Register.
	KVMClockMSR uint32
}

// Features returns the processor capabilities.
func (cpu *CPU) Features() *Features {
	return &Features{
		InvariantTSC: cpu.invariant,
		KVM:          cpu.kvm,
		KVMClockMSR:  cpu.kvmclock,
	}
}
