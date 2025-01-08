// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

// CPUID function numbers
//
// (Intel® Architecture Instruction Set Extensions
// and Future Features Programming Reference
// 1.5 CPUID INSTRUCTION).
const (
	CPUID_TSC_CCC = 0x15

	CPUID_APM         = 0x80000007
	APM_TSC_INVARIANT = 8
)

// CPUID hypervisor function numbers
//
// (https://docs.kernel.org/virt/acrn/cpuid.html)
const (
	CPUID_HYP_TSC_KHZ = 0x40000010
)

// defined in features.s
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
