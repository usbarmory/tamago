// NXP Cryptographic Acceleration and Assurance Module (CAAM) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package caam

import (
	"crypto/sha256"

	"github.com/usbarmory/tamago/dma"
)

// MasterKeyVerification outputs an unencrypted blob key encryption key (BKEK)
// derived from the hardware unique key (internal OTPMK, when SNVS is enabled).
//
// *WARNING*: when SNVS is not enabled a default non-unique test vector is used
// and therefore key derivation is *unsafe*, see snvs.Available().
func (hw *CAAM) MasterKeyVerification() (key []byte, err error) {
	key = make([]byte, sha256.Size)

	destinationBufferAddress := dma.Alloc(key, len(key))
	defer dma.Free(destinationBufferAddress)

	// output sequence start address
	dst := SeqOutPtr{}
	dst.SetDefaults()
	dst.Length(len(key))
	dst.Pointer(destinationBufferAddress)

	// Encapsulation protocol, Master Key Verification Blob
	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_PROT_ENC)
	op.Protocol(PROTID_BLOB, (BLOB_FORMAT_MKV << PROTINFO_BLOB_FORMAT))

	jd := dst.Bytes()
	jd = append(jd, op.Bytes()...)

	hdr := Header{}
	hdr.SetDefaults()
	hdr.Length(1 + len(jd)/4)

	jd = append(hdr.Bytes(), jd...)

	if err = hw.job(jd); err != nil {
		return
	}

	dma.Read(destinationBufferAddress, 0, key)

	return
}
