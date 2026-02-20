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
)

const (
	sharedPageOp = 2 << 52
	pageSize     = 4096
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// 2.3.1 GHCB MSR Protocol.
const (
	MSR_AMD_GHCB = 0xc0010130
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// 2.6 GHCB Layout.
const (
	SW_EXITCODE  = 0x0390
	SW_EXITINFO1 = 0x0398
	SW_EXITINFO2 = 0x03a0
	SW_SCRATCH   = 0x03a8
	VALID_BITMAP = 0x03f0

	SharedBuffer     = 0x0800
	SharedBufferSize = 0x7f0
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 7: List of Supported Non-Automatic Events.
const SNP_GUEST_REQUEST = 0x80000011

// GHCB represents a Guest-Hypervisor Communication Block instance, used to
// expose register state to an AMD SEV-ES hypervisor.
type GHCB struct {
	// Layout is a required unencrypted memory page for shared
	// guest/hypervisor access of the GHCB Layout.
	Layout *dma.Region

	// Region is an unencrypted memory region for shared guest/hypervisor
	// buffers, required for any function issuing [GHCB.GuestRequest].
	Region *dma.Region

	// DMA buffer
	addr uint
	buf  []byte

	seqNo uint64
}

// defined in sev.s
func vmgexit()

func (b *GHCB) write(off uint, val uint64) {
	binary.LittleEndian.PutUint64(b.buf[off:off+8], val)
}

func (b *GHCB) valid(offsets []uint64) {
	for i := range 16 {
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

func (b *GHCB) read(off uint64) (val uint64) {
	return binary.LittleEndian.Uint64(b.buf[off : off+8])
}

// Init initializes a Guest-Hypervisor Communication Block instance, mapping
// its memory location for guest/hypervisor access.
//
// The argument DMA region must be initialized and have been previously
// allocated as unencrypted for hypervisor access (e.g. C-bit disabled).
func (b *GHCB) Init() (err error) {
	if b.Layout == nil {
		return errors.New("invalid instance, no GHCB page")
	}

	b.addr, b.buf = b.Layout.Reserve(int(b.Layout.Size()), pageSize)
	b.seqNo = 1

	return
}

// Exit triggers an Automatic Exit (AE) event to transfer control to an AMD
// SEV-ES hypervisor for updated GHCB access. The arguments represent guest
// state towards the hypervisor, the return values represent hypervisor state
// towards the guest.
func (b *GHCB) Exit(code, info1, info2, scratch uint64) (err error) {
	if b.Layout == nil {
		return errors.New("invalid instance, no GHCB page")
	}

	b.write(SW_EXITCODE, code)
	b.write(SW_EXITINFO1, info1)
	b.write(SW_EXITINFO2, info2)
	b.write(SW_SCRATCH, scratch)

	valid := []uint64{SW_EXITCODE, SW_EXITINFO1, SW_EXITINFO2}

	if scratch > 0 {
		valid = append(valid, SW_SCRATCH)
	}

	b.valid(valid)
	vmgexit()

	if exit := b.read(SW_EXITCODE); exit != code {
		return fmt.Errorf("exit code mismatch (%#x)", exit)
	}

	info1 = b.read(SW_EXITINFO1)
	info2 = b.read(SW_EXITINFO2)

	if info1 != 0 || info2 != 0 {
		return fmt.Errorf("exit error (info1:%#x info2:%#x)", info1, b.read(SW_EXITINFO2))
	}

	return
}

// Dump returns a copy of the GHCB memory.
func (b *GHCB) Dump() (buf []byte) {
	buf = make([]byte, pageSize)
	copy(buf, b.buf)
	return
}

// GuestRequest issues an SNP Guest Request from an SEV-SNP guest to the
// SEV-SNP firmware running inside the Platform Security Processor (PSP).
//
// The message is protected with authenticated AES-256 GCM encryption, using
// the argument key index and value (see [SNPSecrets.VMPCK]).
//
// See [SEV Secure Nested Paging Firmware ABI Specification - Chapter 7].
func (b *GHCB) GuestRequest(index int, key, req []byte, messageType int) (res []byte, err error) {
	var msg []byte

	if b.Region == nil {
		return nil, errors.New("invalid instance, nil DMA Region")
	}

	// SEV Secure Nested Paging Firmware ABI Specification
	// 8.26 SNP_GUEST_REQUEST

	hdr := &MessageHeader{
		Algo:           AES_256_GCM,
		HeaderVersion:  headerVersion,
		HeaderSize:     headerSize,
		MessageType:    uint8(messageType),
		MessageVersion: messageVersion,
		VMPCK:          uint8(index),
	}

	// encrypt request message
	if msg, err = b.sealMessage(hdr, req, key); err != nil {
		return
	}

	reqAddr, reqBuf := b.Region.Reserve(pageSize, pageSize)
	defer b.Region.Release(reqAddr)

	resAddr, resBuf := b.Region.Reserve(pageSize, pageSize)
	defer b.Region.Release(resAddr)

	// zero out response buffer flush speculative read
	copy(resBuf, make([]byte, pageSize))
	copy(reqBuf, msg)

	// yield to hypervisor
	if err = b.Exit(SNP_GUEST_REQUEST, uint64(reqAddr), uint64(resAddr), 0); err != nil {
		return
	}

	b.seqNo += 1

	// copy response buffer as soon as possible as GHCB might overwrite it
	buf := make([]byte, pageSize)
	copy(buf, resBuf)

	if err = hdr.unmarshal(buf); err != nil {
		return nil, fmt.Errorf("could not parse response header, %v", err)
	}

	if msg, err = b.openMessage(hdr, buf[headerSize:headerSize+hdr.MessageSize], key); err != nil {
		return nil, fmt.Errorf("could not decrypt response message, %v", err)
	}

	return msg, nil
}
