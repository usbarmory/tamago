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

// A single DCP channel is used for all operations, this entails that only one
// digest state can be kept at any given time.
var sem = semaphore.NewWeighted(1)

// Hash is the common interface to DCP hardware backed hash functions.
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
	mode uint32
	bs   int
	init bool
	term bool
	buf  []byte
	sum  []byte
}

// New256 returns a new Digest computing the SHA256 checksum.
//
// A single DCP channel is used for all operations, this entails that only one
// digest instance can be kept at any given time, if this condition is not met
// an error is returned.
//
// The digest instance starts with New256() and terminates when when Sum() is
// invoked, after which the digest state can no longer be changed.
func New256() (Hash, error) {
	if !sem.TryAcquire(1) {
		return nil, errors.New("another digest instance is already in use")
	}

	d := &digest{
		mode: HASH_SELECT_SHA256,
		bs:   64,
		init: true,
	}

	return d, nil
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

// Sum appends the current hash to in and returns the resulting slice.  Its
// invocation terminates the digest instance, for this reason Write will return
// errors after Sum is invoked.
func (d *digest) Sum(in []byte) (sum []byte, err error) {
	if len(d.sum) != 0 {
		return append(in, d.sum[:]...), nil
	}

	defer sem.Release(1)
	d.term = true

	s, err := hash(d.buf, HASH_SELECT_SHA256, d.init, d.term)

	if err != nil {
		return
	}

	d.sum = s

	return append(in, d.sum[:]...), nil
}

// BlockSize returns the hash's underlying block size.
func (d *digest) BlockSize() int {
	return d.bs
}

// Sum256 returns the SHA256 checksum of the data.
//
// There must be sufficient DMA memory allocated to hold the data, otherwise
// the function will panic.
func Sum256(data []byte) (sum [32]byte, err error) {
	s, err := hash(data, HASH_SELECT_SHA256, true, true)

	if err != nil {
		return
	}

	copy(sum[:], s)

	return
}
