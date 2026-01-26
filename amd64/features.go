// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"runtime"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
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

	CPUID_INFO        = 0x01
	INFO_HYPERVISOR   = 31
	INFO_TSC_DEADLINE = 24

	CPUID_INTEL_CACHE = 0x04

	CPUID_INTEL_APIC = 0x0b
	INTEL_APIC_LP    = 0

	CPUID_TSC_CCC = 0x15
	CPUID_CPU_FRQ = 0x16

	CPUID_APM         = 0x80000007
	APM_TSC_INVARIANT = 8
	APM_HW_PSTATE     = 7
)

// CPUID function numbers
//
// (AMD64 Architecture Programmer’s Manual
// Volume 3 - Appendix E.4 Extended Feature Function Numbers.
const (
	CPUID_AMD_PROC = 0x80000008
	AMD_PROC_NC    = 0

	CPUID_AMD_ENCM = 0x8000001f
)

// KVM CPUID function numbers
//
// (https://docs.kernel.org/virt/kvm/x86/cpuid.html)
const (
	CPUID_KVM_SIGNATURE = 0x40000000
	KVM_SIGNATURE       = 0x4b4d564b // "KVMK"

	CPUID_KVM_FEATURES    = 0x40000001
	FEATURES_CLOCKSOURCE  = 0
	FEATURES_CLOCKSOURCE2 = 3

	CPUID_KVM_TSC_KHZ = 0x40000010
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

// Features represents the processor capabilities detected through the CPUID
// instruction.
type Features struct {
	// HwPstate indicates whether Hardware P-State control MSRs are
	// supported.
	HwPstate bool

	// TSCInvariant indicates whether the Time Stamp Counter is guaranteed
	// to be at constant rate.
	TSCInvariant bool
	// TSCDeadline indicates whether TSC-Deadline Mode of operation is
	// available for the local-APIC timer to support [CPU.SetAlarm].
	TSCDeadline bool

	// KVM indicates whether a Kernel-base Virtual Machine is detected.
	KVM bool
	// KVMClockMSR returns the kvmclock Model Specific Register.
	KVMClockMSR uint32
}

// defined in features.s
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

// CPUID returns the processor capabilities.
func (cpu *CPU) CPUID(leaf, subleaf uint32) (eax, ebx, ecx, edx uint32) {
	return cpuid(leaf, subleaf)
}

// MSR returns a machine-specific register.
func (cpu *CPU) MSR(addr uint64) (val uint64) {
	return reg.ReadMSR(addr)
}

func (cpu *CPU) initFeatures() {
	_, _, _, apmFeatures := cpuid(CPUID_APM, 0)
	cpu.features.HwPstate = bits.IsSet(&apmFeatures, APM_HW_PSTATE)
	cpu.features.TSCInvariant = bits.IsSet(&apmFeatures, APM_TSC_INVARIANT)

	_, _, cpuFeatures, _ := cpuid(CPUID_INFO, 0)
	cpu.features.TSCDeadline = bits.IsSet(&cpuFeatures, INFO_TSC_DEADLINE)

	if _, kvmk, _, _ := cpuid(CPUID_KVM_SIGNATURE, 0); kvmk != KVM_SIGNATURE {
		return
	}

	cpu.features.KVM = true
	kvmFeatures, _, _, _ := cpuid(CPUID_KVM_FEATURES, 0)

	if bits.IsSet(&kvmFeatures, FEATURES_CLOCKSOURCE) {
		cpu.features.KVMClockMSR = MSR_KVM_SYSTEM_TIME
	}

	if bits.IsSet(&kvmFeatures, FEATURES_CLOCKSOURCE2) {
		cpu.features.KVMClockMSR = MSR_KVM_SYSTEM_TIME_NEW
	}
}

// Features returns the processor capabilities.
func (cpu *CPU) Features() *Features {
	return &cpu.features
}

// NumCPU returns the number of logical CPUs available on the platform.
func NumCPU() (n int) {
	_, _, ecx, _ := cpuid(CPUID_VENDOR, 0)

	switch ecx {
	case CPUID_VENDOR_ECX_AMD:
		// AMD64 Architecture Programmer’s Manual
		// Volume 3 - E4.7
		_, _, ecx, _ := cpuid(CPUID_AMD_PROC, 0)
		n = int(bits.Get(&ecx, AMD_PROC_NC, 0xff)) + 1
	case CPUID_VENDOR_ECX_INTEL:
		// Intel® Architecture Instruction Set Extensions and Future
		// Features Programming Reference
		// 5.1.12 x2APIC Features / Processor Topology (Function 0Bh)
		_, ebx, _, _ := cpuid(CPUID_INTEL_APIC, 1) // core sublevel
		n = int(bits.Get(&ebx, INTEL_APIC_LP, 0xffff))
	}

	if n == 0 {
		n = runtime.NumCPU()
	}

	return
}
