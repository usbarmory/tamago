// NXP Data Co-Processor (DCP) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dcp

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// DeriveKey derives a hardware unique key in a manner equivalent to PKCS#11
// C_DeriveKey with CKM_AES_CBC_ENCRYPT_DATA.
//
// The diversifier is AES-128-CBC encrypted using the internal OTPMK (when SNVS
// is enabled).
//
// *WARNING*: when SNVS is not enabled a default non-unique test vector is used
// and therefore key derivation is *unsafe*, see snvs.Available().
//
// A negative index argument results in the derived key being computed and
// returned.
//
// An index argument equal or greater than 0 moves the derived key, through
// DeriveKeyMemory, to the corresponding internal DCP key RAM slot (see
// SetKey()). In this case no key is returned by the function.
func (hw *DCP) DeriveKey(diversifier []byte, iv []byte, index int) (key []byte, err error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("invalid IV size")
	}

	// prepare diversifier for in-place encryption
	key = pad(diversifier, false)

	region := hw.DeriveKeyMemory

	if index >= 0 {
		if region == nil {
			return nil, errors.New("invalid DeriveKeyMemory")
		}
	}

	sourceBufferAddress := region.Alloc(key, aes.BlockSize)
	defer region.Free(sourceBufferAddress)

	payloadPointer := region.Alloc(iv, 0)
	defer region.Free(payloadPointer)

	pkt := &WorkPacket{}
	pkt.SetCipherDefaults()

	// Use device-specific hardware key for encryption.
	pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	pkt.Control0 |= 1 << DCP_CTRL0_OTP_KEY
	pkt.Control1 |= KEY_SELECT_UNIQUE_KEY << DCP_CTRL1_KEY_SELECT
	pkt.SourceBufferAddress = uint32(sourceBufferAddress)
	pkt.DestinationBufferAddress = pkt.SourceBufferAddress
	pkt.BufferSize = uint32(len(key))
	pkt.PayloadPointer = uint32(payloadPointer)

	ptr := region.Alloc(pkt.Bytes(), 0)
	defer region.Free(ptr)

	if err = hw.cmd(ptr, 1); err != nil {
		return nil, err
	}

	if index >= 0 {
		return nil, hw.setKeyData(index, nil, pkt.DestinationBufferAddress)
	} else {
		region.Read(sourceBufferAddress, 0, key)
	}

	return
}

func (hw *DCP) setKeyData(index int, key []byte, addr uint32) (err error) {
	var keyLocation uint32
	var subword uint32

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	if key != nil && len(key) > aes.BlockSize {
		return errors.New("invalid key size")
	}

	bits.SetN(&keyLocation, KEY_INDEX, 0b11, uint32(index))

	hw.Lock()
	defer hw.Unlock()

	for subword < 4 {
		off := subword * 4

		bits.SetN(&keyLocation, KEY_SUBWORD, 0b11, subword)
		reg.Write(hw.key, keyLocation)

		if key != nil {
			k := key[off : off+4]
			reg.Write(hw.keydata, binary.LittleEndian.Uint32(k))
		} else {
			reg.Move(hw.keydata, addr+off)
		}

		subword++
	}

	return
}

// SetKey configures an AES-128 key in one of the 4 available slots of the DCP
// key RAM.
func (hw *DCP) SetKey(index int, key []byte) (err error) {
	return hw.setKeyData(index, key, 0)
}

func pad(buf []byte, extraBlock bool) []byte {
	padLen := 0
	r := len(buf) % aes.BlockSize

	if r != 0 {
		padLen = aes.BlockSize - r
	} else if extraBlock {
		padLen = aes.BlockSize
	}

	padding := []byte{(byte)(padLen)}
	padding = bytes.Repeat(padding, padLen)

	return append(buf, padding...)
}
