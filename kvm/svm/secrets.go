// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// SEV Secure Nested Paging Firmware ABI Specification
// Table 71. Secrets Page Format
const (
	SNP_SECRETS_VERSION = 0x000
	SNP_SECRETS_VMPCK0  = 0x020
)

const snpSecretSize = 32

// SNPSecrets represents the SNP Secrets Page format version 2 or 3 (AMD
// SEV-SNP).
type SNPSecrets struct {
	// Address represents the Secrets Page address.
	Address uint
	// Size represents the Secrets Page size.
	Size   int

	// DMA buffer
	buf  []byte
}

// Init initializes a Secrets Page instance, mapping the argument memory
// location for guest/hypervisor access.
func (s *SNPSecrets) Init() (err error) {
	if s.Address == 0 {
		return errors.New("invalid address")
	}

	if s.Size < pageSize {
		return errors.New("invalid size")
	}

	r, err := dma.NewRegion(s.Address, s.Size, false)

	if err != nil {
		return
	}

	_, s.buf = r.Reserve(s.Size, 0)

	if binary.LittleEndian.Uint32(s.buf[0:4]) < 2 {
		return errors.New("invalid version")
	}

	return
}

// VMPCK returns a VM Communication Key (VMPCK).
func (s *SNPSecrets) VMPCK(index int) (vmpck []byte, err error) {
	if len(s.buf) == 0 {
		return nil, errors.New("invalid instance")
	}

	off := SNP_SECRETS_VMPCK0 + index*snpSecretSize
	vmpck = s.buf[off : off+snpSecretSize]

	return
}
