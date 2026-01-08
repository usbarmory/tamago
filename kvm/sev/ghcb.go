// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

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
	MSR_AMD_GHCB = 0xc0010130

	GHCB_MSR_GHCB_REQ = 0x012
	GHCB_MSR_GHCB_RES = 0x013
	GHCB_MSR_PSC_REQ  = 0x014
	GHCB_MSR_PSC_RES  = 0x015
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// 2.6 GHCB Layout.
const (
	RAX          = 0x01f8
	RDX          = 0x0310
	SW_EXITCODE  = 0x0390
	SW_EXITINFO1 = 0x0398
	SW_EXITINFO2 = 0x03a0
	VALID_BITMAP = 0x03f0

	// 2032 bytes shared buffer area
	SharedBuffer = 0x0800
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 7: List of Supported Non-Automatic Events
const (
	RDTSC             = 0x6e
	SNP_GUEST_REQUEST = 0x80000011
)

// GHCB represents a Guest-Hypervisor Communication Block instance, used to
// expose register state to an AMD SEV-ES hypervisor.
type GHCB struct {
	// GHCBPage is a required unencrypted memory page for shared
	// guest/hypervisor access of the GHCB Layout.
	GHCBPage *dma.Region

	// RequestPage is a required unencrypted memory page for shared
	// guest/hypervisor access of SNP Guest Requests.
	RequestPage *dma.Region

	// ResponsePage is an unencrypted memory page for shared
	// guest/hypervisor access of SNP Guest responses, when not set
	// [RequestPage] is used.
	ResponsePage *dma.Region

	// DMA buffer
	addr uint
	buf  []byte

	seqNo uint64
}

// defined in sev.s
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

func (b *GHCB) valid(offsets []uint64) {
	for i := 0; i < 16; i++ {
		b.buf[VALID_BITMAP+i] = 0x00
	}

	// Each GHCB field set by the guest and returned by the hypervisor must
	// have the appropriate bit set in the GHCB VALID_BITMAP field
	// (4 GHCB Protocol).
	for _, off := range offsets {
		bit := off / 8
		b.buf[VALID_BITMAP+bit/8] |= 1 << (bit % 8)
	}
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
	if b.GHCBPage == nil {
		return errors.New("invalid instance, no GHCB page")
	}

	b.addr, b.buf = b.GHCBPage.Reserve(int(b.GHCBPage.Size()), pageSize)
	b.seqNo = 1

	if !register {
		return
	}

	if !b.msr(uint64(b.addr), GHCB_MSR_GHCB_REQ, GHCB_MSR_GHCB_RES) {
		return errors.New("could not register GHCB GPA")
	}

	for i := uint(0); i < b.GHCBPage.Size(); i += pageSize {
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
func (b *GHCB) Exit(code uint64, info1 uint64, info2 uint64) (err error) {
	if b.GHCBPage == nil {
		return errors.New("invalid instance, no GHCB page")
	}

	b.write(SW_EXITCODE, code)
	b.write(SW_EXITINFO1, info1)
	b.write(SW_EXITINFO2, info2)

	// set valid bitmap
	b.valid([]uint64{SW_EXITCODE, SW_EXITINFO1, SW_EXITINFO2})

	vmgexit()

	if exit := b.read(SW_EXITCODE); exit != code {
		return fmt.Errorf("exit code mismatch (%#x)", exit)
	}

	info1 = b.read(SW_EXITINFO1)
	info2 = b.read(SW_EXITINFO2)

	if info1 != 0 || info2 != 0 {
		return fmt.Errorf("exit error (info1:%#x info2:%#x)", info1, b.read(SW_EXITINFO2))
	}

	b.seqNo += 1

	return
}

// Dump returns a copy of the GHCB memory.
func (b *GHCB) Dump() (buf []byte) {
	buf = make([]byte, pageSize)
	copy(buf, b.buf)
	return
}
