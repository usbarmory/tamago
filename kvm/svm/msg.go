// AMD virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"bytes"
	"encoding/binary"
)

const (
	headerVersion = 0x1
	headerSize    = 96
)

// AEAD Algorithm Encodings
const (
	AES_256_GCM = 1
)

// MessageHeader represents an AMD SEV-SNP Message Header.
type MessageHeader struct {
	AuthTag        [32]byte
	SeqNo          [16]byte
	Algo           uint8
	HeaderVersion  uint8
	HeaderSize     uint16
	MessageType    uint8
	MessageVersion uint8
	MessageSize    uint16
	_              uint32
	VMPCK          uint8
	_              [35]byte
}

// Bytes converts the descriptor structure to byte array format.
func (h *MessageHeader) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, h)
	return buf.Bytes()
}
