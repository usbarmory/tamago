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
	"errors"
	"io"

	"golang.org/x/sync/semaphore"
)

// A single CAAM job ring is used for all operations, this entails that only
// one digest state can be kept at any given time.
var sem = semaphore.NewWeighted(1)

// Hash is the common interface to CAAM hardware backed hash functions.
//
// While similar to Go native hash.Hash, this interface is not fully compatible
// with it as hardware errors must be checked and checksum computation affects
// state.
type Hash interface {
	// Write (via the embedded io.Writer interface) adds more data to the running hash.
	// It can return an error. It returns an error if Sum has been already invoked.
	io.Writer

	// Sum appends the current hash to b and returns the resulting slice.
	// Its invocation terminates the digest instance, for this reason Write
	// will return errors after Sum is invoked.
	Sum(b []byte) ([]byte, error)

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if all writes
	// are a multiple of the block size.
	BlockSize() int
}

type digest struct {
	caam *CAAM
	mode int
	n    int
	bs   int
	init bool
	buf  []byte
	sum  []byte
}

// Write adds more data to the running hash. It returns an error if Sum has
// been already invoked or in case of hardware errors.
//
// There must be sufficient DMA memory allocated to hold the data, otherwise
// the function will panic.
func (d *digest) Write(p []byte) (n int, err error) {
	if len(d.sum) != 0 {
		return 0, errors.New("digest instance can no longer be used")
	}

	// If we still don't have enough data for a block, accumulate and early
	// out.
	if len(d.buf)+len(p) < d.bs {
		d.buf = append(d.buf, p...)
		return len(p), nil
	}

	pl := len(p)

	// top up partial block buffer, and process that
	cut := d.bs - len(d.buf)
	d.buf = append(d.buf, p[:cut]...)
	p = p[cut:]

	if _, err = d.caam.hash(d.buf, d.mode, d.n, d.init, false); err != nil {
		return
	}

	if d.init {
		d.init = false
	}

	// work through any more full blocks in p
	if l := len(p); l > d.bs {
		r := l % d.bs

		if _, err = d.caam.hash(p[:l-r], d.mode, d.n, d.init, false); err != nil {
			return
		}

		p = p[l-r:]
	}

	// save off any partial block remaining
	d.buf = append(d.buf[0:0], p...)

	return pl, nil
}

// Sum appends the current hash to in and returns the resulting slice.  Its
// invocation terminates the digest instance, for this reason Write will return
// errors after Sum is invoked.
func (d *digest) Sum(in []byte) (sum []byte, err error) {
	if len(d.sum) != 0 {
		return append(in, d.sum[:]...), nil
	}

	defer sem.Release(1)

	if d.init && len(d.buf) == 0 {
		d.sum = sha256.New().Sum(nil)
	} else {
		s, err := d.caam.hash(d.buf, ALG_SHA256, d.n, d.init, true)

		if err != nil {
			return nil, err
		}

		d.sum = s
	}

	return append(in, d.sum[:]...), nil
}

// BlockSize returns the hash's underlying block size.
func (d *digest) BlockSize() int {
	return d.bs
}

// New256 returns a new Digest computing the SHA256 checksum.
//
// A single CAAM channel is used for all operations, this entails that only one
// digest instance can be kept at any given time, if this condition is not met
// an error is returned.
//
// The digest instance starts with New256() and terminates when when Sum() is
// invoked, after which the digest state can no longer be changed.
func (hw *CAAM) New256() (Hash, error) {
	if !sem.TryAcquire(1) {
		return nil, errors.New("another digest instance is already in use")
	}

	d := &digest{
		caam: hw,
		mode: ALG_SHA256,
		n:    sha256.Size,
		bs:   sha256.BlockSize,
		init: true,
		buf:  make([]byte, 0, sha256.BlockSize),
	}

	return d, nil
}

// Sum256 returns the SHA256 checksum of the data.
//
// There must be sufficient DMA memory allocated to hold the data, otherwise
// the function will panic.
func (hw *CAAM) Sum256(data []byte) (sum [32]byte, err error) {
	s, err := hw.hash(data, ALG_SHA256, len(sum), true, true)

	if err != nil {
		return
	}

	copy(sum[:], s)

	return
}
