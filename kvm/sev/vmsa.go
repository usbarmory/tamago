// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"bytes"
	"encoding/binary"
)

// Segment represents a segment descriptor
// (AMD64 Architecture Programmer’s Manual, Volume 2 - Table B-2).
type Segment struct {
	Selector uint16
	Attrib   uint16
	Limit    uint32
	Base     uint64
}

// VMSA represents an AMD SEV-SNP Virtual Machine Save Area page
// (AMD64 Architecture Programmer’s Manual, Volume 2 - Table B-4).
type VMSA struct {
	ES   Segment
	CS   Segment
	SS   Segment
	DS   Segment
	FS   Segment
	GS   Segment
	GDTR Segment
	LDTR Segment
	IDTR Segment
	TR   Segment

	PL0SSP uint64
	PL1SSP uint64
	PL2SSP uint64
	PL3SSP uint64
	U_CET  uint64
	_      [2]byte

	VMPL uint8
	CPL  uint8
	_    [4]byte

	EFER uint64
	_    [104]byte

	XSS         uint64
	CR4         uint64
	CR3         uint64
	CR0         uint64
	DR7         uint64
	DR6         uint64
	RFLAGS      uint64
	RIP         uint64
	DR0         uint64
	DR1         uint64
	DR2         uint64
	DR3         uint64
	DR0AddrMask uint64
	DR1AddrMask uint64
	DR2AddrMask uint64
	DR3AddrMask uint64
	_           [24]byte

	RSP          uint64
	S_CET        uint64
	SSP          uint64
	ISST_ADDR    uint64
	RAX          uint64
	STAR         uint64
	LSTAR        uint64
	CSTAR        uint64
	SFMASK       uint64
	KernelGsBase uint64
	SYSENTER_CS  uint64
	SYSENTER_ESP uint64
	SYSENTER_EIP uint64
	CR2          uint64
	_            [32]byte

	G_PAT        uint64
	DBGCTL       uint64
	BR_FROM      uint64
	BR_TO        uint64
	LASTEXCPFROM uint64
	LASTEXCPTO   uint64
	_            [80]byte

	PKRU    uint32
	TSC_AUX uint32
	_       [24]byte

	RCX uint64
	RDX uint64
	RBX uint64
	_   [8]byte
	RBP uint64
	RSI uint64
	RDI uint64
	R8  uint64
	R9  uint64
	R10 uint64
	R11 uint64
	R12 uint64
	R13 uint64
	R14 uint64
	R15 uint64
	_   [16]byte

	GUEST_EXITINFO1   uint64
	GUEST_EXITINFO2   uint64
	GUEST_EXITINTINFO uint64
	GUEST_NRIP        uint64
	SEV_FEATURES      uint64
	VINTR_CTRL        uint64
	GUEST_EXITCODE    uint64
	VIRTUAL_TOM       uint64
	TLB_ID            uint64
	PCPU_ID           uint64
	EVENTINJ          uint64
	XCR0              uint64
	_                 [16]byte

	X87_DP    uint64
	MXCSR     uint32
	X87_FTW   uint16
	X87_FSW   uint16
	X87_FCW   uint16
	X87_FOP   uint16
	X87_DS    uint16
	X87_CS    uint16
	X87_RIP   uint64
	FPREG_X87 [80]uint8
	FPREG_XMM [256]uint8
	FPREG_YMM [256]uint8
	_         [2448]byte
}

// Bytes converts the descriptor structure to byte array format.
func (v *VMSA) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, v)
	return buf.Bytes()
}

// Init sets the VMSA state to safe defaults for AP initialization by
// [amd64.CPU.InitSMP].
func (v *VMSA) Init(pc uint64) {
	vector := pc >> 12

	// AMD64 Architecture Programmer’s Manual, Volume 2
	// Canonicalization and Consistency Checks (p505).

	v.CS.Selector = uint16(vector) << 8
	v.CS.Base = uint64(vector) << 12
	v.CS.Limit = 0xffff
	v.CS.Attrib = 0x009b // Present, Code, Executable, Read, Accessed

	v.DS.Limit = 0xffff
	v.DS.Attrib = 0x0093 // Present, Data, Read/Write, Accessed

	v.ES = v.DS
	v.FS = v.DS
	v.GS = v.DS
	v.SS = v.DS

	v.GDTR.Limit = 0xffff
	v.IDTR.Limit = 0xffff

	v.LDTR.Limit = 0xffff
	v.LDTR.Attrib = 0x0082 // Present, LDT

	v.TR.Limit = 0xffff
	v.TR.Attrib = 0x008B // Present, Busy TSS

	v.CR0 = 0x60000010 // Cache Disabled, Not Writethrough
	v.EFER = 0x1000    // SVME Secure Virtual Machine Enable

	v.RIP = pc & 0xfff

	// set reserved bits
	v.RFLAGS = 0x2
	v.DR6 = 0xffff0ff0
	v.DR7 = 0x400

	// set architectural defaults
	v.G_PAT = 0x0007040600070406
	v.XCR0 = 0x1
	v.MXCSR = 0x1f80
	v.X87_FCW = 0x0040
	v.X87_FTW = 0x5555

	return
}
