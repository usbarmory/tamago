// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// The following build tag is to allow testing of this package using:
//	GOOS=tamago $TAMAGO test -tags fakenet,user_linux
//
//go:build !user_linux

// Package sev implements a driver for AMD Secure Encrypted Virtualization
// (SEV), following reference specifications:
//
//   - AMD64 Architecture Programmer’s Manual, Volume 2
//   - SEV-ES Guest-Hypervisor Communication Block Standardization
//   - SEV Secure Nested Paging Firmware ABI Specification
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package sev

import (
	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/bits"
)

// AMD64 Architecture Programmer’s Manual, Volume 2
// 15.34.10 SEV_STATUS MSR
const (
	MSR_AMD_SEV_STATUS = 0xc0010131
	SEV_STATUS_SEV_SNP = 2
	SEV_STATUS_SEV_ES  = 1
	SEV_STATUS_SEV     = 0
)

// SEVStatus represents active AMD Secure Encrypted Virtualization (SEV)
// features.
type SEVStatus struct {
	// SEV reports whether the KVM guest was run with SEV.
	SEV bool
	// SEV reports whether the KVM guest was run with SEV-ES.
	ES bool
	// SEV reports whether the KVM guest was run with SEV-SNP.
	SNP bool
}

// SVMFeatures represents the processor AMD Secure Virtual Machine
// capabilities.
type SVMFeatures struct {
	// SEV represents the AMD SEV-SNP features
	SEV SEVStatus
	// EncryptedBit represents the C-bit position
	EncryptedBit int
}

// Features returns the processor AMD Secure Virtual Machine capabilities.
func Features(cpu *amd64.CPU) (f *SVMFeatures) {
	_, _, ecx, _ := cpu.CPUID(amd64.CPUID_VENDOR, 0)

	if ecx != amd64.CPUID_VENDOR_ECX_AMD {
		return
	}

	status := uint32(cpu.MSR(MSR_AMD_SEV_STATUS))
	_, ebx, _, _ := cpu.CPUID(amd64.CPUID_AMD_ENCM, 0)

	f = &SVMFeatures{
		SEV: SEVStatus{
			SEV: bits.Get(&status, SEV_STATUS_SEV),
			ES:  bits.Get(&status, SEV_STATUS_SEV_ES),
			SNP: bits.Get(&status, SEV_STATUS_SEV_SNP),
		},
		EncryptedBit: int(ebx & 0b111111),
	}

	return
}
