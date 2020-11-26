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
	"errors"
	"io"

	"golang.org/x/sync/semaphore"
)

var sem = semaphore.NewWeighted(1)

// Hash is the common interface to DCP hardware backed hash functions.
//
//
// While similar to Go native hash.Hash, this interface is not fully compatible
// with it as hardware errors must be checked and checksum computation affects
// state.
type Hash interface {
	// Write (via the embedded io.Writer interface) adds more data to the running hash.
	// It can return an error.
	io.Writer

	// Sum appends the current hash to b and returns the resulting slice.
	// It does change the underlying hash state and can only be invoked once.
	Sum(b []byte) ([]byte, error)

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if all writes
	// are a multiple of the block size.
	BlockSize() int
}

type digest struct {
	mode uint32
	bs   int
	init bool
	term bool
	buf  []byte
}

// New256 returns a new Digest computing the SHA256 checksum.
//
// A single DCP channel is used for all operations, this entails that hash
// instances must be used exclusively, otherwise an error is returned.
//
// This exclusive access begins with New256() and ends when Sum() is invoked,
// after which the digest state can no longer be changed.
func New256() (d *digest, err error) {
	if !sem.TryAcquire(1) {
		return nil, errors.New("another digest instance is already in use")
	}

	d = &digest{
		mode: HASH_SELECT_SHA256,
		bs:   64,
		init: true,
	}

	return
}

func (d *digest) Write(p []byte) (n int, err error) {
	if len(d.buf) != 0 {
		_, err = hash(d.buf, d.mode, d.init, d.term)

		if err != nil {
			return
		}

		if d.init {
			d.init = false
		}
	}

	d.buf = p

	return len(p), nil
}

func (d *digest) Sum(in []byte) (sum []byte, err error) {
	if d.term {
		return nil, errors.New("digest instance can no longer be used")
	}

	defer sem.Release(1)
	d.term = true

	s, err := hash(d.buf, HASH_SELECT_SHA256, d.init, d.term)

	if err != nil {
		return
	}

	return append(in, s[:]...), nil
}

func (d *digest) BlockSize() int {
	return d.bs
}

// Sum256 returns the SHA256 checksum of the data.
func Sum256(data []byte) (sum [32]byte, err error) {
	s, err := hash(data, HASH_SELECT_SHA256, true, true)

	if err != nil {
		return
	}

	copy(sum[:], s)

	return
}
