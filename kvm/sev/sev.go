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

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 1: FEATURES Bitmap.
const (
	FeatureSNP                      = 0
	FeatureAPCreation               = 1
	FeatureRestrictedInjection      = 2
	FeatureRestrictedInjectionTimer = 3
	FeatureAPICIDList               = 4
	FeatureMultiVMPL                = 5
	FeaturePSC                      = 6
	FeatureTIO                      = 7
	FeatureUnregister               = 8
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 7: List of Supported Non-Automatic Events.
const HV_FEATURE_SUPPORT = 0x8000fffd

// SEVStatus represents active AMD Secure Encrypted Virtualization (SEV)
// features.
type SEVStatus struct {
	// SEV reports whether the KVM guest was run with SEV.
	SEV bool
	// SEV reports whether the KVM guest was run with SEV-ES.
	ES bool
	// SEV reports whether the KVM guest was run with SEV-SNP.
	SNP bool

	// Features reports Guest-controlled SEV feature selection.
	Features uint64
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
			SEV:      bits.Get(&status, SEV_STATUS_SEV),
			ES:       bits.Get(&status, SEV_STATUS_SEV_ES),
			SNP:      bits.Get(&status, SEV_STATUS_SEV_SNP),
			Features: uint64(status >> 2),
		},
		EncryptedBit: int(ebx & 0b111111),
	}

	return
}

// HypervisorFeatures requests the hypervisor features support bitmap.
func (b *GHCB) HypervisorFeatures() (features uint64, err error) {
	if err = b.Exit(HV_FEATURE_SUPPORT, 0, 0, 0); err != nil {
		return
	}

	return b.read(SW_EXITINFO2), nil
}
