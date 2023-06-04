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
	"errors"

	"github.com/usbarmory/tamago/dma"
)

func (hw *CAAM) cipher(buf []byte, key []byte, iv []byte, mode uint32, enc bool) (err error) {
	if len(buf)%aes.BlockSize != 0 {
		return errors.New("invalid input size")
	}

	switch len(key) {
	case 16, 24, 32:
		break
	default:
		return aes.KeySizeError(len(key))
	}

	if len(iv) != aes.BlockSize {
		return errors.New("invalid IV size")
	}

	keyBufferAddress := dma.Alloc(key, 4)
	defer dma.Free(keyBufferAddress)

	loadKey := Key{}
	loadKey.SetDefaults()
	loadKey.Class(1)
	loadKey.Pointer(keyBufferAddress, len(key))

	ivBufferAddress := dma.Alloc(iv, 4)
	defer dma.Free(ivBufferAddress)

	loadIV := Load{}
	loadIV.SetDefaults()
	loadIV.Class(1)
	loadIV.Destination(CTX)
	loadIV.Pointer(ivBufferAddress, len(iv))

	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_ALG_CLASS1)
	op.Algorithm(ALG_AES, mode)
	op.State(AS_INITIALIZE | AS_FINALIZE)
	op.Encrypt(enc)

	sourceBufferAddress := dma.Alloc(buf, 4)
	defer dma.Free(sourceBufferAddress)

	src := FIFOLoad{}
	src.SetDefaults()
	src.Class(1)
	src.DataType(INPUT_DATA_TYPE_MESSAGE_DATA | INPUT_DATA_TYPE_LC1)
	src.Pointer(sourceBufferAddress, len(buf))

	dst := FIFOStore{}
	dst.SetDefaults()
	dst.DataType(OUTPUT_DATA_TYPE_MESSAGE_DATA)
	dst.Pointer(sourceBufferAddress, len(buf))

	jd := loadKey.Bytes()
	jd = append(jd, loadIV.Bytes()...)
	jd = append(jd, op.Bytes()...)
	jd = append(jd, src.Bytes()...)
	jd = append(jd, dst.Bytes()...)

	if err = hw.job(nil, jd); err != nil {
		return
	}

	dma.Read(sourceBufferAddress, 0, buf)

	return
}

// Encrypt performs in-place buffer encryption using AES-CBC, the key argument
// should be the AES key, either 16, 24, or 32 bytes to select AES-128,
// AES-192, or AES-256.
//
// All argument buffers are copied to reserved buffers in the default DMA
// region, buffers previously created with dma.Reserve() can be used to avoid
// external RAM exposure, when placed in iRAM, as their pointers are directly
// passed to the CAAM without access by the Go runtime.
func (hw *CAAM) Encrypt(buf []byte, key []byte, iv []byte) (err error) {
	return hw.cipher(buf, key, iv, AAI_AES_CBC, true)
}

// Decrypt performs in-place buffer decryption using AES-CBC, the key argument
// should be the AES key, either 16, 24, or 32 bytes to select AES-128,
// AES-192, or AES-256.
//
// All argument buffers are copied to reserved buffers in the default DMA
// region, buffers previously created with dma.Reserve() can be used to avoid
// external RAM exposure, when placed in iRAM, as their pointers are directly
// passed to the CAAM without access by the Go runtime.
func (hw *CAAM) Decrypt(buf []byte, key []byte, iv []byte) (err error) {
	return hw.cipher(buf, key, iv, AAI_AES_CBC, false)
}
