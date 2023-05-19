// NXP Cryptographic Acceleration and Assurance Module (CAAM) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package caam

import (
	"bytes"
	"encoding/binary"

	"github.com/usbarmory/tamago/bits"
)

// p266, 6.6.7.3 Command types, IMX6ULSRM
const (
	CTYPE = 27

	CTYPE_OPERATION   = 0b10000
	CTYPE_HEADER      = 0b10110
	CTYPE_SEQ_OUT_PTR = 0b11111
)

// p333, 6.6.17 PROTOCOL OPERATION command, IMX6ULSRM
const (
	OPERATION_OPTYPE = 24
	OPTYPE_PROT_ENC  = 0b111

	OPERATION_PROTID = 16
	PROTID_BLOB      = 0x0d

	OPERATION_PROTINFO   = 0
	PROTINFO_BLOB_FORMAT = 0
	BLOB_FORMAT_MKV      = 0b10
)

// p276, 6.6.8 HEADER command, IMX6ULSRM
const (
	HEADER_ONE     = 23
	HEADER_DESCLEN = 0
)

// p375, 6.6.23 SEQ OUT PTR command, IMX6ULSRM
const (
	SEQ_OUT_PTR_LENGTH = 0
)

// Command represents a CAAM command
// (p266, 6.6.7.3 Command types, IMX6ULSRM).
type Command struct {
	// CTYPE field
	Type uint32
	// Main fields
	Word0 uint32
	// Optional words
	Words []uint32
}

// Bytes converts the descriptor non-optional words structure to byte array
// format.
func (c *Command) Bytes() []byte {
	bits.SetN(&c.Word0, CTYPE, 0x1f, c.Type)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.Word0)
	binary.Write(buf, binary.LittleEndian, c.Words)

	return buf.Bytes()
}

// Operation represents a CAAM OPERATION command.
type Operation struct {
	Command
}

// SetDefaults initializes default values for the OPERATION command.
func (c *Operation) SetDefaults() {
	c.Type = CTYPE_OPERATION
}

// OpType sets the OPERATION command OPTYPE field.
func (c *Operation) OpType(op int) {
	bits.SetN(&c.Word0, OPERATION_OPTYPE, 0b111, uint32(op))
}

// Protocol sets the OPERATION command PROTID and PROTINFO fields.
func (c *Operation) Protocol(id int, info uint32) {
	bits.SetN(&c.Word0, OPERATION_PROTID, 0b111, uint32(id))
	bits.SetN(&c.Word0, OPERATION_PROTINFO, 0xffff, info)
}

// Header represents a CAAM HEADER command.
// (p276, 6.6.8 HEADER command, IMX6ULSRM).
type Header struct {
	Command
}

// SetDefaults initializes default values for the HEADER command.
func (c *Header) SetDefaults() {
	c.Type = CTYPE_HEADER
	bits.Set(&c.Word0, HEADER_ONE)
}

// Length sets the HEADER command DESCLEN field.
func (c *Header) Length(words int) {
	bits.SetN(&c.Word0, HEADER_DESCLEN, 0x7f, uint32(words))
}

// SeqOutPtr represents a CAAM SEQ OUT PTR command.
type SeqOutPtr struct {
	Command
}

// SetDefaults initializes default values for the SEQ OUT PTR command.
func (c *SeqOutPtr) SetDefaults() {
	c.Type = CTYPE_SEQ_OUT_PTR
}

// Length sets the SEQ OUT PTR command LENGTH field.
func (c *SeqOutPtr) Length(n int) {
	bits.SetN(&c.Word0, SEQ_OUT_PTR_LENGTH, 0xffff, uint32(n))
}

// Pointer sets the SEQ OUT PTR command optional POINTER field.
func (c *SeqOutPtr) Pointer(addr uint) {
	c.Words = append(c.Words, uint32(addr))
}
