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

	CTYPE_KEY         = 0b00000
	CTYPE_LOAD        = 0b00010
	CTYPE_FIFO_LOAD   = 0b00100
	CTYPE_STORE       = 0b01010
	CTYPE_FIFO_STORE  = 0b01100
	CTYPE_OPERATION   = 0b10000
	CTYPE_JUMP        = 0b10100
	CTYPE_HEADER      = 0b10110
	CTYPE_SEQ_IN_PTR  = 0b11110
	CTYPE_SEQ_OUT_PTR = 0b11111
)

// Fields common across multiple commands
const (
	// KEY, LOAD, STORE, JUMP commands
	CLASS = 25

	// ALGORITHM, PROTOCOL, PKHA OPERATION commands
	OPERATION_OPTYPE = 24

	// LOAD, STORE commands
	EXT       = 22
	DATA_TYPE = 16
	LENGTH    = 0
)

// Field values common across multiple commands
const (
	// LOAD, STORE commands
	CLRW = 0x08
	CTX  = 0x20
)

// p285, 6.6.10 JUMP command, IMX6ULSRM
const (
	LOAD_IMM = 23
)

// p296, 6.6.11 FIFO LOAD command, IMX6ULSRM
// p313, 6.6.14 FIFO STORE command, IMX6ULSRM
const (
	INPUT_DATA_TYPE_PKHA_Ax      = 0b000000
	INPUT_DATA_TYPE_IV           = 0b100000
	INPUT_DATA_TYPE_MESSAGE_DATA = 0b010000
	INPUT_DATA_TYPE_LC2          = 1 << 2
	INPUT_DATA_TYPE_LC1          = 1 << 1

	OUTPUT_DATA_TYPE_MESSAGE_DATA = 0x30
)

// p328, 6.6.16 ALGORITHM OPERATION command, IMX6ULSRM
const (
	OPTYPE_ALG_CLASS1 = 0b010
	OPTYPE_ALG_CLASS2 = 0b100

	OPERATION_ALG = 16
	ALG_AES       = 0x10
	ALG_SHA256    = 0x43
	ALG_SHA512    = 0x45
	ALG_RNG       = 0x50

	OPERATION_AAI = 4
	AAI_AES_CBC   = 0x10
	AAI_AES_CMAC  = 0x60
	AAI_RNG_SK    = 8

	OPERATION_AS  = 2
	AS_UPDATE     = 0b00
	AS_INITIALIZE = 0b01
	AS_FINALIZE   = 0b10

	OPERATION_ENC = 0
)

// p333, 6.6.17 PROTOCOL OPERATION command, IMX6ULSRM
const (
	OPTYPE_PROT_UNI = 0b000
	OPTYPE_PROT_DEC = 0b110
	OPTYPE_PROT_ENC = 0b111

	OPERATION_PROTID  = 16
	PROTID_BLOB       = 0x0d
	PROTID_ECDSA_SIGN = 0x15

	OPERATION_PROTINFO = 0

	PROTINFO_BLOB_FORMAT = 0
	BLOB_FORMAT_MKV      = 0b10

	PROTINFO_SIGN_NO_TEQ = 12
	PROTINFO_ECC         = 1
)

// p356, 6.6.20 JUMP command, IMX6ULSRM
const (
	JUMP_OFFSET = 0
)

// p276, 6.6.8 HEADER command, IMX6ULSRM
const (
	HEADER_ONE         = 23
	HEADER_START_INDEX = 16
	HEADER_DESCLEN     = 0
)

// Command represents a CAAM command
// (p266, 6.6.7.3 Command types, IMX6ULSRM).
type Command struct {
	// Main fields
	Word0 uint32
	// Optional words
	Words []uint32
}

// Class sets the command CLASS field.
func (c *Command) Class(class int) {
	bits.SetN(&c.Word0, CLASS, 0b11, uint32(class))
}

// Bytes converts the descriptor non-optional words structure to byte array
// format.
func (c *Command) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.Word0)
	binary.Write(buf, binary.LittleEndian, c.Words)

	return buf.Bytes()
}

// LengthCommand represents a CAAM command with Length and Pointer fields.
type LengthCommand struct {
	Command
}

// Pointer sets the command POINTER and EXT LENGTH fields.
func (c *LengthCommand) Pointer(addr uint, n int) {
	bits.SetN(&c.Word0, LENGTH, 0xff, uint32(n))
	c.Words = []uint32{uint32(addr)}
}

// ExtendedLengthCommand represents a CAAM command with Extended Length and
// Pointer fields.
type ExtendedLengthCommand struct {
	Command
}

// Pointer sets the command POINTER and EXT LENGTH fields.
func (c *ExtendedLengthCommand) Pointer(addr uint, n int) {
	bits.Set(&c.Word0, EXT)
	c.Words = []uint32{
		uint32(addr),
		uint32(n),
	}
}

// Key represents a KEY command
// (p281, 6.6.9 KEY commands, IMX6ULSRM).
type Key struct {
	LengthCommand
}

// SetDefaults initializes default values for the KEY command.
func (c *Key) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_KEY)
}

// Load represents a LOAD command
// (p285, 6.6.10 LOAD commands, IMX6ULSRM).
type Load struct {
	LengthCommand
}

// SetDefaults initializes default values for the LOAD command.
func (c *Load) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_LOAD)
}

// Destination sets the LOAD command DST field.
func (c *Load) Destination(dst int) {
	bits.SetN(&c.Word0, DATA_TYPE, 0x7f, uint32(dst))
}

// Immediate sets the LOAD command IMM, LENGTH and Value fields.
func (c *Load) Immediate(imm uint32) {
	bits.Set(&c.Word0, LOAD_IMM)
	bits.SetN(&c.Word0, LENGTH, 0xff, 4)
	c.Words = []uint32{imm}
}

// FIFOLoad represents a FIFO LOAD command
// (p296, 6.6.11 FIFO LOAD command, IMX6ULSRM).
type FIFOLoad struct {
	ExtendedLengthCommand
}

// SetDefaults initializes default values for the FIFO LOAD command.
func (c *FIFOLoad) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_FIFO_LOAD)
}

// DataType sets the FIFO LOAD command INPUT DATA TYPE field.
func (c *FIFOLoad) DataType(dt int) {
	bits.SetN(&c.Word0, DATA_TYPE, 0x3f, uint32(dt))
}

// Store represents a STORE command
// (p306, 6.6.13 STORE command, IMX6ULSRM).
type Store struct {
	LengthCommand
}

// SetDefaults initializes default values for the STORE command.
func (c *Store) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_STORE)
}

// Source sets the STORE command SRC field.
func (c *Store) Source(src int) {
	bits.SetN(&c.Word0, DATA_TYPE, 0x7f, uint32(src))
}

// FIFOStore represents a FIFO STORE command
// (p313, 6.6.14 FIFO STORE command, IMX6ULSRM).
type FIFOStore struct {
	ExtendedLengthCommand
}

// SetDefaults initializes default values for the FIFO STORE command.
func (c *FIFOStore) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_FIFO_STORE)
}

// DataType sets the FIFO STORE command OUTPUT DATA TYPE field.
func (c *FIFOStore) DataType(dt int) {
	bits.SetN(&c.Word0, DATA_TYPE, 0x3f, uint32(dt))
}

// Operation represents a CAAM OPERATION command.
type Operation struct {
	Command
}

// SetDefaults initializes default values for the OPERATION command.
func (c *Operation) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_OPERATION)
}

// OpType sets the OPERATION command OPTYPE field.
func (c *Operation) OpType(op int) {
	bits.SetN(&c.Word0, OPERATION_OPTYPE, 0b111, uint32(op))
}

// Algorithm sets the ALGORITHM OPERATION command ALG and AAI fields.
func (c *Operation) Algorithm(alg int, aai uint32) {
	bits.SetN(&c.Word0, OPERATION_ALG, 0xff, uint32(alg))
	bits.SetN(&c.Word0, OPERATION_AAI, 0x1ff, aai)
}

// Algorithm sets the ALGORITHM OPERATION command AS field.
func (c *Operation) State(as int) {
	bits.SetN(&c.Word0, OPERATION_AS, 0b11, uint32(as))
}

// Algorithm sets the ALGORITHM OPERATION command ENC field.
func (c *Operation) Encrypt(enc bool) {
	bits.SetTo(&c.Word0, OPERATION_ENC, enc)
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
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_HEADER)
	bits.Set(&c.Word0, HEADER_ONE)
}

// Length sets the HEADER command DESCLEN field.
func (c *Header) Length(words int) {
	bits.SetN(&c.Word0, HEADER_DESCLEN, 0x7f, uint32(words))
}

// StartIndex sets the HEADER command START INDEX field.
func (c *Header) StartIndex(off int) {
	bits.SetN(&c.Word0, HEADER_START_INDEX, 0x3f, uint32(off))
}

// Jump represents a CAAM JUMP command.
// (p356, 6.6.20 JUMP command, IMX6ULSRM).
type Jump struct {
	Command
}

// SetDefaults initializes default values for the JUMP command.
func (c *Jump) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_JUMP)
}

// Offset sets the JUMP command LOCAL OFFSET field.
func (c *Jump) Offset(off int) {
	bits.SetN(&c.Word0, JUMP_OFFSET, 0xff, uint32(off))
}

// SeqInPtr represents a CAAM SEQ IN PTR command
// (p372, 6.6.22 SEQ  IN PTR command, IMX6ULSRM).
type SeqInPtr struct {
	ExtendedLengthCommand
}

// SetDefaults initializes default values for the SEQ IN PTR command.
func (c *SeqInPtr) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_SEQ_IN_PTR)
}

// SeqOutPtr represents a CAAM SEQ OUT PTR command
// (p375, 6.6.23 SEQ OUT PTR command, IMX6ULSRM).
type SeqOutPtr struct {
	ExtendedLengthCommand
}

// SetDefaults initializes default values for the SEQ OUT PTR command.
func (c *SeqOutPtr) SetDefaults() {
	bits.SetN(&c.Word0, CTYPE, 0x1f, CTYPE_SEQ_OUT_PTR)
}
