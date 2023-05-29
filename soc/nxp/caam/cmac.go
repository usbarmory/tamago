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

func (hw *CAAM) cmac(msg []byte, key []byte, mac []byte) (err error) {
	switch len(key) {
	case 16, 24, 32:
		break
	default:
		return aes.KeySizeError(len(key))
	}

	if len(mac) != aes.BlockSize {
		return errors.New("invalid mac size")
	}

	keyBufferAddress := dma.Alloc(key, 4)
	defer dma.Free(keyBufferAddress)

	loadKey := Key{}
	loadKey.SetDefaults()
	loadKey.Class(1)
	loadKey.Pointer(keyBufferAddress, len(key))

	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_ALG_CLASS1)
	op.Algorithm(ALG_AES, AAI_AES_CMAC)
	op.State(AS_INITIALIZE | AS_FINALIZE)

	sourceBufferAddress := dma.Alloc(msg, 4)
	defer dma.Free(sourceBufferAddress)

	src := FIFOLoad{}
	src.SetDefaults()
	src.Class(1)
	src.DataType(INPUT_DATA_TYPE_MESSAGE_DATA | INPUT_DATA_TYPE_LC1)
	src.Pointer(sourceBufferAddress, len(msg))

	destinationBufferAddress := dma.Alloc(mac, 4)
	defer dma.Free(destinationBufferAddress)

	dst := Store{}
	dst.SetDefaults()
	dst.Class(1)
	dst.Source(CTX)
	dst.Pointer(destinationBufferAddress, len(mac))

	jd := loadKey.Bytes()
	jd = append(jd, op.Bytes()...)
	jd = append(jd, src.Bytes()...)
	jd = append(jd, dst.Bytes()...)

	if err = hw.job(nil, jd); err != nil {
		return
	}

	dma.Read(destinationBufferAddress, 0, mac)

	return
}

// SumAES returns the AES Cipher-based message authentication code (CMAC) of
// the input message, the key argument should be the AES key, either 16, 24, or
// 32 bytes to select AES-128, AES-192, or AES-256.
//
// There must be sufficient DMA memory allocated to hold the message, otherwise
// the function will panic.
func (hw *CAAM) SumAES(msg []byte, key []byte) (sum [16]byte, err error) {
	err = hw.cmac(msg, key, sum[:])
	return
}
