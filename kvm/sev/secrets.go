// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

const secretsVersion = 0x4

// SecretsPage represents an AMD SEV-SNP Secrets Page format version 4
// (SEV Secure Nested Paging Firmware ABI Specification - Table 71).
type SecretsPage struct {
	Version         uint32
	_               uint32
	FMS             uint32
	_               uint32
	GOSVW           [16]byte
	VMPCK0          [32]byte
	VMPCK1          [32]byte
	VMPCK2          [32]byte
	VMPCK3          [32]byte
	GuestArea1      [96]byte
	VMSATweakBitmap [64]byte
	GuestArea2      [32]byte
	TSCFactor       uint32
	_               uint32
	LaunchMitVector uint64
}

// Init initializes a Secrets Page instance, mapping the argument memory
// location for guest/hypervisor access.
func (s *SecretsPage) Init(addr uint, size int) (err error) {
	if addr == 0 {
		return errors.New("invalid address")
	}

	if size < pageSize {
		return errors.New("invalid size")
	}

	r, err := dma.NewRegion(addr, size, false)

	if err != nil {
		return
	}

	_, buf := r.Reserve(size, 0)

	if _, err = binary.Decode(buf, binary.LittleEndian, s); err != nil {
		return
	}

	if s.Version != secretsVersion {
		return errors.New("unsupported version")
	}

	return
}
