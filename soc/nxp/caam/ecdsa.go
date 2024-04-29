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
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
)

// p451, Table 8-112, IMX7DSSRM
const (
	DSA_SIG_PDB_PD     = 22
	DSA_SIG_PDB_ECDSEL = 7
)

// p443, Table 8-101, IMX7DSSRM
const (
	// Table 8-101
	ECDSEL_P256   = 0x02
	ECDSEL_P256K1 = 0x20
)

// SignPDB represents an ECDSA sign protocol data block (PDB).
type SignPDB struct {
	// size of the group
	n int
	// elliptic curve domain selection
	ecdsel int
	// private key
	s uint
	// message hash
	f uint
	// signature buffer
	c uint
	// signature buffer (2nd part, n length)
	d uint

	// DMA buffer
	sig []byte
}

// Init initializes a PDB for ECDSA signing.
func (pdb *SignPDB) Init(priv *ecdsa.PrivateKey) (err error) {
	name := priv.PublicKey.Curve.Params().Name

	switch name {
	case "P-256":
		pdb.n = 32
		pdb.ecdsel = ECDSEL_P256
	case "P-256k1":
		pdb.n = 32
		pdb.ecdsel = ECDSEL_P256K1
	default:
		return fmt.Errorf("unsupported curve %s", name)
	}

	pdb.n = priv.PublicKey.Curve.Params().BitSize / 8

	pdb.s = dma.Alloc(make([]byte, pdb.n), 4)
	dma.Write(pdb.s, 0, priv.D.Bytes())

	pdb.f, _ = dma.Reserve(pdb.n, 4)

	pdb.c, pdb.sig = dma.Reserve(pdb.n*2, 4)
	pdb.d = pdb.c + uint(pdb.n)

	return
}

func (pdb *SignPDB) Hash(hash []byte) {
	dma.Write(pdb.f, 0, hash[0:pdb.n])
}

// Bytes converts the PDB to byte array format.
func (pdb *SignPDB) Bytes() []byte {
	var word0 uint32

	// p451, Table 8-112, IMX7DSSRM

	bits.Set(&word0, DSA_SIG_PDB_PD)
	bits.SetN(&word0, DSA_SIG_PDB_ECDSEL, 0x7f, uint32(pdb.ecdsel))

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(word0))
	binary.Write(buf, binary.LittleEndian, uint32(pdb.s))
	binary.Write(buf, binary.LittleEndian, uint32(pdb.f))
	binary.Write(buf, binary.LittleEndian, uint32(pdb.c))
	binary.Write(buf, binary.LittleEndian, uint32(pdb.d))

	return buf.Bytes()
}

// Free frees the memory allocated by the PDB.
func (pdb *SignPDB) Free() {
	dma.Release(pdb.c)
	dma.Release(pdb.f)
	dma.Free(pdb.s)
}

// Sign signs a hash (which should be the result of hashing a larger message)
// using the private key, priv. If the hash is longer than the bit-length of
// the private key's curve order, the hash will be truncated to that length. It
// returns the signature as a pair of integers.
//
// A previously initialized sign protocol data block (see SignPDB.Init()) may
// be passed to cache private key initialization, in this case priv is ignored.
func (hw *CAAM) Sign(priv *ecdsa.PrivateKey, hash []byte, pdb *SignPDB) (r, s *big.Int, err error) {
	if pdb == nil {
		pdb = &SignPDB{}
		defer pdb.Free()

		if err = pdb.Init(priv); err != nil {
			return
		}
	} else if pdb.n == 0 {
		return nil, nil, errors.New("pdb is not initialized")
	}

	pdb.Hash(hash)
	jd := pdb.Bytes()

	var info uint32
	bits.Set(&info, PROTINFO_ECC)
	bits.SetTo(&info, PROTINFO_SIGN_NO_TEQ, hw.DisableTimingEqualization)

	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_PROT_UNI)
	op.Protocol(PROTID_ECDSA_SIGN, info)

	hdr := &Header{}
	hdr.SetDefaults()
	hdr.StartIndex(1 + len(jd)/4)

	jd = append(jd, op.Bytes()...)
	hdr.Length(1 + len(jd)/4)

	if err = hw.job(hdr, jd); err != nil {
		return
	}

	r = &big.Int{}
	r.SetBytes(pdb.sig[0:pdb.n])

	s = &big.Int{}
	s.SetBytes(pdb.sig[pdb.n:])

	return
}
