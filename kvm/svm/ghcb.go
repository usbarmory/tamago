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
	GHCB_MSR_REG_GPA_REQ = 0x012
	GHCB_MSR_REG_GPA_RES = 0x013
	GHCB_MSR_PGS_CHG_REQ = 0x014
	GHCB_MSR_PGS_CHG_RES = 0x015

        MSR_AMD_GHCB = 0xc0010130
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
func pvalidate(addr uint64, validate bool) uint32

func (b *GHCB) msr(val uint64, req uint64, res uint64) (valid bool) {
	// 2.3.1 GHCB MSR Protocol
	reg.WriteMSR(MSR_AMD_GHCB, val|req)
	vmgexit()

	return reg.ReadMSR(MSR_AMD_GHCB) == (val | res)
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
	gpa := uint64(b.addr)

	if !register {
		return
	}

	if !b.msr(gpa|sharedPage, GHCB_MSR_PGS_CHG_REQ, GHCB_MSR_PGS_CHG_RES) {
		return errors.New("could not change GHCB GPA page state")
	}

	if pvalidate(gpa, true) != 0 {
		return errors.New("could not PVALIDATE GHCB GPA")
	}

	if !b.msr(gpa, GHCB_MSR_REG_GPA_REQ, GHCB_MSR_REG_GPA_RES) {
		return errors.New("could not register GHCB GPA")
	}

	return
}

func (b *GHCB) write(off uint, val uint64) {
	binary.LittleEndian.PutUint64(b.buf[off:off+8], val)
}

func (b *GHCB) read(off uint) (val uint64) {
	return binary.LittleEndian.Uint64(b.buf[off:off+8])
}

// Exit triggers an Automatic Exit (AE) event to transfer control to an AMD
// SEV-ES hypervisor for updated GHCB access.
func (b *GHCB) Exit(code uint64, info1 uint64, info2 uint64) (exit1 uint64, exit2 uint64) {
	b.write(SW_EXITCODE, code)
	b.write(SW_EXITINFO1, info1)
	b.write(SW_EXITINFO2, info2)

	vmgexit()
	b.seqNo += 1

	exit1 = b.read(SW_EXITINFO1)
	exit2 = b.read(SW_EXITINFO2)

	return
}

// Clear clears any previous GHCB field invocation data.
func (b *GHCB) Clear() {
	// TODO
}
