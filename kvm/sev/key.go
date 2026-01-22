// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MSG_KEY_REQ = 3
	MSG_KEY_RSP = 4
)

// KeySelect masks
const (
	KeySelVLEK       = 2 << 1 // Versioned Loaded Endorsement Key
	KeySelVCEK       = 1 << 1 // Versioned Chip Endorsement Key (VCEK)
	KeySelVLEKOrVCEK = 0 << 1
	RootKeySelVMRK   = 1 << 0 // VM root key
	RootKeySelVCEK   = 0 << 0 // Versioned Chip Endorsement Key
)

// GuestFieldSelect masks
const (
	LaunchMitVector = 1 << 6 // Launch Mitigation Vector
	TCBVersion      = 1 << 5 // Trusted Computing Base Version
	GuestSVN        = 1 << 4 // Guest Security Version Number
	Measurement     = 1 << 3
	FamilyID        = 1 << 2
	ImageID         = 1 << 1
	GuestPolicy     = 1 << 0
)

// KeyRequest represents an AMD SEV-SNP Key Request Message
// (SEV Secure Nested Paging Firmware ABI Specification
// Table 19: MSG_KEY_REQ Message Structure).
type KeyRequest struct {
	KeySelect        uint32
	_                uint32
	GuestFieldSelect uint64
	VMPL             uint32
	GuestSVN         uint32
	TCBVersion       uint64
	LaunchMitVector  uint64
	_                [24]byte
}

// Bytes converts the descriptor structure to byte array format.
func (r *KeyRequest) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r)
	return buf.Bytes()
}

// KeyResponse represents an AMD SEV-SNP Key Request Response
// (SEV Secure Nested Paging Firmware ABI Specification
// Table 21: MSG_KEY_RSP Message Structure).
type KeyResponse struct {
	Status     uint32
	_          [28]byte
	DerivedKey [32]byte
}

func (r *KeyResponse) unmarshal(buf []byte) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, r)
	return
}

// DeriveKey sends an AMD SEV-SNP guest request for key derivation. The
// arguments represent guest provided request parameters, the VM Communication
// Key (see [SNPSecrets.VMPCK]) payload and index for encrypting the request.
func (b *GHCB) DeriveKey(req *KeyRequest, key []byte, index int) (dk []byte, err error) {
	var buf []byte

	res := &KeyResponse{}

	if buf, err = b.GuestRequest(index, key, req.Bytes(), MSG_KEY_REQ); err != nil {
		return
	}

	if err = res.unmarshal(buf); err != nil {
		return nil, fmt.Errorf("could not parse response, %v", err)
	}

	if res.Status != 0 {
		return nil, fmt.Errorf("request error, %#x", res.Status)
	}

	return res.DerivedKey[:], nil
}
