// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

// SecureTSC performs a read of the time-stamp counter through the
// Guest-Hypervisor Communication Block.
func (b *GHCB) SecureTSC() (tsc uint64, err error) {
	if err = b.Exit(RDTSC, 0, 0); err != nil {
		return
	}

	vmgexit()

	return b.read(RDX)<<32 | b.read(RAX), nil
}
