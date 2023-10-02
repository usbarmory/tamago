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
	"crypto/aes"
	"crypto/sha256"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// mkv requests the CAAM Master Key Verification Blob, the resulting
// unencrypted blob key encryption key (BKEK), derived from the hardware unique
// key (internal OTPMK, when SNVS is enabled), is written to buf.
func (hw *CAAM) mkv(buf []byte) (err error) {
	if len(buf) != sha256.Size {
		return errors.New("invalid input size")
	}

	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_PROT_ENC)
	op.Protocol(PROTID_BLOB, (BLOB_FORMAT_MKV << PROTINFO_BLOB_FORMAT))

	destinationBufferAddress := dma.Alloc(buf, 4)
	defer dma.Free(destinationBufferAddress)

	dst := SeqOutPtr{}
	dst.SetDefaults()
	dst.Pointer(destinationBufferAddress, len(buf))

	jd := dst.Bytes()
	jd = append(jd, op.Bytes()...)

	if err = hw.job(nil, jd); err != nil {
		return
	}

	dma.Read(destinationBufferAddress, 0, buf)

	return
}

// DeriveKey derives a hardware unique key in a manner equivalent to NXP
// Symmetric key diversifications guidelines (AN10922 - Rev. 2.2) for AES-256
// keys.
//
// The diversifier is used as message for AES-256-CMAC authentication using a
// blob key encryption key (BKEK) derived from the hardware unique key
// (internal OTPMK, when SNVS is enabled, through Master Key Verification
// Blob).
//
// *WARNING*: when SNVS is not enabled a default non-unique test vector is used
// and therefore key derivation is *unsafe*, see snvs.Available().
//
// The unencrypted BKEK is used through DeriveKeyMemory. An output key buffer
// previously created with DeriveKeyMemory.Reserve() can be used to avoid
// external RAM exposure, when placed in iRAM, as its pointer is directly
// passed to the CAAM without access by the Go runtime.
func (hw *CAAM) DeriveKey(diversifier []byte, key []byte) (err error) {
	if len(diversifier) > sha256.Size {
		return errors.New("invalid diversifier size")
	}

	if len(key) != sha256.Size {
		return errors.New("invalid key size")
	}

	region := hw.DeriveKeyMemory

	if region == nil {
		return errors.New("invalid DeriveKeyMemory")
	}

	// Key Diversification (p11, 2.4 AES-256 key, AN10922)
	//
	//   AES-256-CMAC(key, 0x41 || msg || padding) ||
	//   AES-256-CMAC(key, 0x42 || msg || padding)

	// padding is added by the CAAM
	d1 := append([]byte{0x41}, diversifier...)
	d2 := append([]byte{0x42}, diversifier...)

	mkvBufferAddress, mkv := region.Reserve(sha256.Size, 4)
	defer region.Release(mkvBufferAddress)

	if err = hw.mkv(mkv); err != nil {
		return
	}

	d1BufferAddress := region.Alloc(d1, aes.BlockSize)
	defer region.Free(d1BufferAddress)

	dkaBufferAddress, dka := region.Reserve(aes.BlockSize, 4)
	defer region.Free(dkaBufferAddress)

	if err = hw.cmac(d1, mkv, dka); err != nil {
		return
	}

	d2BufferAddress := region.Alloc(d2, aes.BlockSize)
	defer region.Free(d2BufferAddress)

	dkbBufferAddress, dkb := region.Reserve(aes.BlockSize, 4)
	defer region.Free(dkbBufferAddress)

	if err = hw.cmac(d2, mkv, dkb); err != nil {
		return
	}

	copy(key[0:16], dka)
	copy(key[16:32], dkb)

	return
}
