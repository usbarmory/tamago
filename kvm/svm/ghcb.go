// AMD virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	MSR_AMD_GHCB = 0xc0010130
	GHCBSize     = 4096
)

// GHCB represents a Guest-Hypervisor Communication Block instance, used to
// expose register state to an AMD SEV-ES hypervisor.
type GHCB struct {
	// DMA buffer
	addr uint
	buf  []byte
}

// Init initializes a Guest-Hypervisor Communication Block instance, mapping
// its memory location for guest/hypervisor access.
func (b *GHCB) Init() {
	b.addr, b.buf = dma.Reserve(GHCBSize, GHCBSize)
	reg.WriteMSR(MSR_AMD_GHCB, uint32(b.addr))
}

// Yield triggers an Automatic Exit (AE) event to transfer control to an AMD
// SEV-ES hypervisor for updated GHCB access.
func (b *GHCB) Yield() {
	vmgexit()
}
