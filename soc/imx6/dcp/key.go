// NXP Data Co-Processor (DCP) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dcp

import (
	"crypto/aes"
	"encoding/binary"
	"errors"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// The i.MX6 On-Chip RAM (OCRAM/iRAM) is used for passing DCP derived keys to
// its internal key RAM without touching external RAM.
const (
	iramStart = 0x00900000
	iramSize  = 0x20000
)

// DeriveKey derives a hardware unique key in a manner equivalent to PKCS#11
// C_DeriveKey with CKM_AES_CBC_ENCRYPT_DATA.
//
// The diversifier is AES-CBC encrypted using the internal OTPMK key (when SNVS
// is enabled).
//
// A negative index argument results in the derived key being computed and
// returned.
//
// An index argument equal or greater than 0 moves the derived key directly to
// the corresponding internal DCP key RAM slot (see SetKey()). This is
// accomplished through an iRAM reserved DMA buffer, to ensure that the key is
// never exposed to external RAM or the Go runtime. In this case no key is
// returned by the function.
func DeriveKey(diversifier []byte, iv []byte, index int) (key []byte, err error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("invalid IV size")
	}

	// prepare diversifier for in-place encryption
	key = pad(diversifier, false)

	region := dma.Default()

	if index >= 0 {
		// force use of iRAM if not already set as default DMA region
		if region.Start < iramStart || region.Start > iramStart+iramSize {
			region = &dma.Region{
				Start: iramStart,
				Size:  iramSize,
			}

			region.Init()
		}
	}

	pkt := &WorkPacket{}
	pkt.SetCipherDefaults()

	// Use device-specific hardware key for encryption.
	pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	pkt.Control0 |= 1 << DCP_CTRL0_OTP_KEY
	pkt.Control1 |= KEY_SELECT_UNIQUE_KEY << DCP_CTRL1_KEY_SELECT

	pkt.BufferSize = uint32(len(key))

	pkt.SourceBufferAddress = region.Alloc(key, aes.BlockSize)
	defer region.Free(pkt.SourceBufferAddress)

	pkt.DestinationBufferAddress = pkt.SourceBufferAddress

	pkt.PayloadPointer = region.Alloc(iv, 0)
	defer region.Free(pkt.PayloadPointer)

	ptr := region.Alloc(pkt.Bytes(), 0)
	defer region.Free(ptr)

	err = cmd(ptr, 1)

	if err != nil {
		return
	}

	if index >= 0 {
		err = setKeyData(index, nil, pkt.SourceBufferAddress)
	} else {
		region.Read(pkt.SourceBufferAddress, 0, key)
	}

	return
}

func setKeyData(index int, key []byte, addr uint32) (err error) {
	var keyLocation uint32
	var subword uint32

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	if key != nil && len(key) > aes.BlockSize {
		return errors.New("invalid key size")
	}

	bits.SetN(&keyLocation, KEY_INDEX, 0b11, uint32(index))

	mux.Lock()
	defer mux.Unlock()

	for subword < 4 {
		off := subword * 4

		bits.SetN(&keyLocation, KEY_SUBWORD, 0b11, subword)
		reg.Write(DCP_KEY, keyLocation)

		if key != nil {
			k := key[off : off+4]
			reg.Write(DCP_KEYDATA, binary.LittleEndian.Uint32(k))
		} else {
			reg.Move(DCP_KEYDATA, addr+off)
		}

		subword++
	}

	return
}

// SetKey configures an AES-128 key in one of the 4 available slots of the DCP
// key RAM.
func SetKey(index int, key []byte) (err error) {
	return setKeyData(index, key, 0)
}
