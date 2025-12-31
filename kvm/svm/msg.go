// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
)

const (
	headerVersion = 0x1
	headerSize    = 96

	messageVersion = 0x1
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

func (h *MessageHeader) unmarshal(buf []byte) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, h)
	return
}

// Bytes converts the descriptor structure to byte array format.
func (h *MessageHeader) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, h)
	return buf.Bytes()
}

// Seq returns the 64-bit sequence number for this message.
func (h *MessageHeader) Seq() uint64 {
	return binary.LittleEndian.Uint64(h.SeqNo[:])
}

// SetSeq sets the 64-bit sequence number for this message.
func (h *MessageHeader) SetSeq(seq uint64) {
	binary.LittleEndian.PutUint64(h.SeqNo[:], seq)
}

func seal(key, nonce, plaintext, additionalData []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintext, additionalData)

	return
}

func unseal(key, nonce, ciphertext, additionalData []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return
	}

	return aesgcm.Open(nil, nonce, ciphertext, additionalData)
}

func (b *GHCB) sealMessage(hdr *MessageHeader, plaintext, key []byte) (msg []byte, err error) {
	// update header
	hdr.MessageSize = uint16(len(plaintext))
	hdr.SetSeq(b.seqNo)

	// encrypt request
	ciphertext, err := seal(key, hdr.SeqNo[0:12], plaintext, hdr.Bytes()[48:])

	if err != nil {
		return
	}

	// set authentication tag
	tagOffset := len(ciphertext) - 16
	authTag := ciphertext[tagOffset:]
	copy(hdr.AuthTag[:], authTag)

	// concatenate header and encrypted message
	msg = hdr.Bytes()
	msg = append(msg, ciphertext[:tagOffset]...)

	return
}

func (b *GHCB) openMessage(hdr *MessageHeader, ciphertext, key []byte) (plaintext []byte, err error) {
	if hdr.Seq() != b.seqNo {
		return nil, errors.New("invalid response header")
	}

	// append authentication tag to ciphertext
	ciphertext = append(ciphertext, hdr.AuthTag[:16]...)

	// zero AuthTag header before unseal
	copy(hdr.AuthTag[:], make([]byte, 32))

	// decrypt response
	if plaintext, err = unseal(key, hdr.SeqNo[0:12], ciphertext, hdr.Bytes()[48:]); err != nil {
		return
	}

	return
}
