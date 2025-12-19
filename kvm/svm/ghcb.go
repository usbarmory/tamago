// AMD virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"encoding/binary"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	MSR_AMD_GHCB = 0xc0010130
	pageSize     = 4096
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
	seqNo uint64

	// DMA buffer
	addr uint
	buf  []byte
}

// Init initializes a Guest-Hypervisor Communication Block instance, mapping
// its memory location for guest/hypervisor access.
//
// The initialization will panic unless a default DMA region is allocated with
// sufficient size for the desired guest/hypervisor request/responses payloads.
func (b *GHCB) Init() {
	b.seqNo = 1
	b.addr, b.buf = dma.Reserve(pageSize, pageSize)
	reg.WriteMSR(MSR_AMD_GHCB, uint32(b.addr))
}

// Yield triggers an Automatic Exit (AE) event to transfer control to an AMD
// SEV-ES hypervisor for updated GHCB access.
func (b *GHCB) Yield() {
	vmgexit()
}

// Clear clears any previous GHCB field invocation data.
func (b *GHCB) Clear() {
	// TODO
}

func (b *GHCB) write(off uint, val uint64) {
	binary.LittleEndian.PutUint64(b.buf[off:off+8], val)
}
