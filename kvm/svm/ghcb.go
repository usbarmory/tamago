// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	sharedPage = 2 << 52
	pageSize   = 4096
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// 2.3.1 GHCB MSR Protocol.
const (
	GHCB_MSR_GHCB_REQ = 0x012
	GHCB_MSR_GHCB_RES = 0x013
	GHCB_MSR_PSC_REQ  = 0x014
	GHCB_MSR_PSC_RES  = 0x015
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// 2.6 GHCB Layout.
const (
	SW_EXITCODE  = 0x0390
	SW_EXITINFO1 = 0x0398
	SW_EXITINFO2 = 0x03a0
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 7: List of Supported Non-Automatic Events
const (
	SNP_GUEST_REQUEST = 0x80000011
)

// GHCB represents a Guest-Hypervisor Communication Block instance, used to
// expose register state to an AMD SEV-ES hypervisor.
type GHCB struct {
	// SharedMemory is a required unencrypted memory region for shared
	// guest/hypervisor access.
	SharedMemory *dma.Region

	// DMA buffer
	addr uint
	buf  []byte

	seqNo uint64
}

// defined in svm.s
func vmgexit()
func pvalidate(addr uint64, validate bool) (ret uint32)

func (b *GHCB) msr(val uint64, req uint64, res uint64) (valid bool) {
	// 2.3.1 GHCB MSR Protocol
	reg.WriteMSR(MSR_AMD_GHCB, val|req)
	vmgexit()

	return reg.ReadMSR(MSR_AMD_GHCB) == (val | res)
}

func (b *GHCB) write(off uint, val uint64) {
	binary.LittleEndian.PutUint64(b.buf[off:off+8], val)
}

func (b *GHCB) read(off uint) (val uint64) {
	return binary.LittleEndian.Uint64(b.buf[off : off+8])
}

// Init initializes a Guest-Hypervisor Communication Block instance, mapping
// its memory location for guest/hypervisor access.
//
// The argument DMA region must be initialized and have been previously
// allocated as unencrypted for hypervisor access (e.g. C-bit disabled).
func (b *GHCB) Init(register bool) (err error) {
	if b.SharedMemory == nil {
		return errors.New("invalid instance, no shared memory")
	}

	b.addr, b.buf = b.SharedMemory.Reserve(pageSize, pageSize)

	if !register {
		return
	}

	if !b.msr(uint64(b.addr), GHCB_MSR_GHCB_REQ, GHCB_MSR_GHCB_RES) {
		return errors.New("could not register GHCB GPA")
	}

	for i := uint(0); i < b.SharedMemory.Size(); i += pageSize {
		gpa := uint64(b.addr + i)

		if ret := pvalidate(gpa, false); ret != 0 {
			return fmt.Errorf("could not rescind page validation (%d)", ret)
		}

		if !b.msr(gpa|sharedPage, GHCB_MSR_PSC_REQ, GHCB_MSR_PSC_RES) {
			return errors.New("could not change page state")
		}
	}

	return
}

// Exit triggers an Automatic Exit (AE) event to transfer control to an AMD
// SEV-ES hypervisor for updated GHCB access. The arguments represent guest
// state towards the hypervisor, the return values represent hypervisor state
// towards the guest.
func (b *GHCB) Exit(code uint64, info1 uint64, info2 uint64) (uint64, uint64, uint64) {
	b.write(SW_EXITCODE, code)
	b.write(SW_EXITINFO1, info1)
	b.write(SW_EXITINFO2, info2)

	vmgexit()
	b.seqNo += 1

	return b.read(SW_EXITCODE), b.read(SW_EXITINFO1), b.read(SW_EXITINFO2)
}
