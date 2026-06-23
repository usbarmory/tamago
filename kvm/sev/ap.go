// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/bits"
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 7: List of Supported Non-Automatic Events.
const SNP_AP_CREATION = 0x80000013

const (
	add           = 0
	addAndRun     = 1
	stopAndRemove = 2
)

// CreateAP creates a vCPU, using the provided Virtual Machine Save Area
// information (VMSA), to be used as Application Processor (AP) for SMP
// operation.
func (b *GHCB) CreateAP(apicid int, v *VMSA) (err error) {
	var features uint64
	var attrs uint64

	if features, err = b.HypervisorFeatures(); err != nil {
		return
	} else {
		bits.Set64(&features, FeatureSNP)
		bits.Set64(&features, FeatureAPCreation)
	}

	if b.VMSA == nil {
		return errors.New("invalid instance, nil VMSA Region")
	}

	addr, buf := b.VMSA.Reserve(pageSize, 0)
	copy(buf, v.Bytes())
	gpa := uint64(addr)

	// ensure page is mapped as 4K
	if ret := pvalidate(gpa, PAGE_SIZE_2M, false); ret == 0 {
		for i := uint64(0); i < (2 << 20); i += (1 << 12) {
			pvalidate(gpa+i, PAGE_SIZE_4K, true)
		}
	}

	bits.Set64(&attrs, RMP_VMSA)
	bits.SetN64(&attrs, RMP_TARGET_VMPL, 0xff, 1)

	if ret := rmpadjust(gpa, PAGE_SIZE_4K, attrs); ret != 0 {
		return fmt.Errorf("rmpadjust error, gpa:%#x attrs:%x ret:%d\n", gpa, attrs, ret)
	}

	// match features
	b.write(RAX, v.SEV_FEATURES)

	info1 := uint64(add)
	info1 |= uint64(v.VMPL) << 16
	info1 |= uint64(apicid) << 32

	return b.Exit(SNP_AP_CREATION, info1, gpa, 0)
}

// RemoveAP removes a vCPU from runnable state.
func (b *GHCB) RemoveAP(apicid int) (err error) {
	info1 := uint64(stopAndRemove)
	info1 |= uint64(apicid) << 32

	return b.Exit(SNP_AP_CREATION, info1, 0, 0)
}
