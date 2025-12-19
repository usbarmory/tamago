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
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

const reportSize = 128

const MSG_REPORT_REQ = 1

// ReportRequest represents an AMD SEV-SNP MSG_REPORT_REQ Message
// (SEV Secure Nested Paging Firmware ABI Specification
// Table 22. MSG_REPORT_REQ Message Structure).
type ReportRequest struct {
	Data   [64]byte
	VMPL   uint32
	KeySel uint32
	_      [24]byte
}

// Bytes converts the descriptor structure to byte array format.
func (m *ReportRequest) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, m)
	return buf.Bytes()
}

// GetAttestationReport sends a guest request for an attestation report through
// the Guest-Hypervisor Communication Block. The arguments represent guest
// provided data and the VM Communication Key (see [SNPSecrets.VPMCK]) payload
// and index for encrypting the request.
func (b *GHCB) GetAttestationReport(data []byte, key []byte, index int) (res []byte, err error) {
	if b.addr == 0 {
		return nil, errors.New("invalid instance")
	}

	if len(data) > 64 {
		return nil, errors.New("data length must not exceed %d bytes")
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return
	}

	hdr := &MessageHeader{
		Algo:          AES_256_GCM,
		HeaderVersion: headerVersion,
		HeaderSize:    headerSize,
		MessageType:   MSG_REPORT_REQ,
		VMPCK:         uint8(index),
	}

	msg := &ReportRequest{
		VMPL:   0,
		KeySel: 0,
	}

	// update sequence number
	binary.LittleEndian.PutUint64(hdr.SeqNo[:], b.seqNo)
	b.seqNo += 2

	// fill message data and update header with its size
	copy(msg.Data[:], data)
	msgData := msg.Bytes()
	hdr.MessageSize = uint16(len(msgData))

	// protect with AEAD
	ciphertext := aesgcm.Seal(nil, hdr.SeqNo[0:12], msgData, hdr.Bytes())
	copy(hdr.AuthTag[:], ciphertext[0:16])

	// concatenate header and encrypted message
	req := hdr.Bytes()
	req = append(req, ciphertext...)

	// allocate shared page for guest/hypervisor communication
	addr, shm := dma.Reserve(pageSize, pageSize)
	defer dma.Release(addr)

	// set up GHCB layout
	b.write(SW_EXITCODE, SNP_GUEST_REQUEST)
	b.write(SW_EXITINFO1, 0)
	b.write(SW_EXITINFO2, uint64(addr))

	// trigger NAE event
	b.Yield()

	// copy response as DMA buffer will be released
	res = make([]byte, pageSize)
	copy(res, shm)

	return
}
