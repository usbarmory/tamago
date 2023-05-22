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

	CTYPE_FIFO_LOAD   = 0b00100
	CTYPE_STORE       = 0b01010
	CTYPE_OPERATION   = 0b10000
	CTYPE_HEADER      = 0b10110
	CTYPE_SEQ_IN_PTR  = 0b11110
	CTYPE_SEQ_OUT_PTR = 0b11111

	// ALGORITHM, PROTOCOL, PKHA OPERATION
	OPERATION_OPTYPE = 24
)

// p296, 6.6.11 FIFO LOAD command, IMX6ULSRM
const (
	FIFO_LOAD_CLASS = 25
	CLASS_2CHA      = 0b10

	FIFO_LOAD_EXT             = 22
	FIFO_LOAD_INPUT_DATA_TYPE = 16

	INPUT_DATA_TYPE_MESSAGE_DATA = 0b010000
	INPUT_DATA_TYPE_LC2          = 0b000100
)

// p306, 6.6.13 STORE command, IMX6ULSRM
const (
	STORE_CLASS = 25
	CLASS_2CCB  = 0b10

	STORE_SRC = 16
	SRC_CTX   = 0x20

	STORE_LENGTH = 0
)

// p328, 6.6.16 ALGORITHM OPERATION command, IMX6ULSRM
const (
	OPTYPE_ALG_CLASS1 = 0b010
	OPTYPE_ALG_CLASS2 = 0b100

	OPERATION_ALG = 16
	ALG_SHA256    = 0x43

	OPERATION_AS  = 2
	AS_UPDATE     = 0b00
	AS_INITIALIZE = 0b01
	AS_FINALIZE   = 0b10
)

// p333, 6.6.17 PROTOCOL OPERATION command, IMX6ULSRM
const (
	OPTYPE_PROT_ENC = 0b111

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

// p372, 6.6.22 SEQ  IN PTR command, IMX6ULSRM
// p375, 6.6.23 SEQ OUT PTR command, IMX6ULSRM
const (
	SEQ_PTR_LENGTH = 0
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

// FIFOLoad represents a FIFO LOAD command.
type FIFOLoad struct {
	Command
}

// SetDefaults initializes default values for the FIFO LOAD command.
func (c *FIFOLoad) SetDefaults() {
	c.Type = CTYPE_FIFO_LOAD
}

// Class sets the FIFO LOAD command CLASS field.
func (c *FIFOLoad) Class(class int) {
	bits.SetN(&c.Word0, FIFO_LOAD_CLASS, 0b11, uint32(class))
}

// DataType sets the FIFO LOAD command INPUT DATA TYPE field.
func (c *FIFOLoad) DataType(dt int) {
	bits.SetN(&c.Word0, FIFO_LOAD_INPUT_DATA_TYPE, 0x3f, uint32(dt))
}

// Pointer sets the FIFO LOAD command POINTER and EXT LENGTH fields.
func (c *FIFOLoad) Pointer(addr uint, n int) {
	bits.Set(&c.Word0, FIFO_LOAD_EXT)
	c.Words = []uint32{
		uint32(addr),
		uint32(n),
	}
}

// Store represents a STORE command.
type Store struct {
	Command
}

// SetDefaults initializes default values for the STORE command.
func (c *Store) SetDefaults() {
	c.Type = CTYPE_STORE
}

// Class sets the STORE command CLASS field.
func (c *Store) Class(class int) {
	bits.SetN(&c.Word0, STORE_CLASS, 0b11, uint32(class))
}

// Source sets the STORE command SRC field.
func (c *Store) Source(src int) {
	bits.SetN(&c.Word0, STORE_SRC, 0x7f, uint32(src))
}

// Pointer sets the STORE command LENGTH and POINTER fields.
func (c *Store) Pointer(addr uint, n int) {
	bits.SetN(&c.Word0, STORE_LENGTH, 0xff, uint32(n))
	c.Words = []uint32{uint32(addr)}
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

// Algorithm sets the ALGORITHM OPERATION command ALG field.
func (c *Operation) Algorithm(alg int) {
	bits.SetN(&c.Word0, OPERATION_ALG, 0xff, uint32(alg))
}

// Algorithm sets the ALGORITHM OPERATION command AS field.
func (c *Operation) State(as int) {
	bits.SetN(&c.Word0, OPERATION_AS, 0b11, uint32(as))
}

// Protocol sets the PROTOCOL OPERATION command PROTID and PROTINFO fields.
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

// SeqInPtr represents a CAAM SEQ IN PTR command.
type SeqInPtr struct {
	Command
}

// SetDefaults initializes default values for the SEQ IN PTR command.
func (c *SeqInPtr) SetDefaults() {
	c.Type = CTYPE_SEQ_IN_PTR
}

// Pointer sets the SEQ IN PTR command LENGHT and POINTER fields.
func (c *SeqInPtr) Pointer(addr uint, n int) {
	bits.SetN(&c.Word0, SEQ_PTR_LENGTH, 0xffff, uint32(n))
	c.Words = append(c.Words, uint32(addr))
}

// SeqOutPtr represents a CAAM SEQ OUT PTR command.
type SeqOutPtr struct {
	Command
}

// SetDefaults initializes default values for the SEQ OUT PTR command.
func (c *SeqOutPtr) SetDefaults() {
	c.Type = CTYPE_SEQ_OUT_PTR
}

// Pointer sets the SEQ OUT PTR command LENGTH and POINTER fields.
func (c *SeqOutPtr) Pointer(addr uint, n int) {
	bits.SetN(&c.Word0, SEQ_PTR_LENGTH, 0xffff, uint32(n))
	c.Words = []uint32{uint32(addr)}
}
