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
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// GetAttestationReport sends a guest request for an AMD SEV-SNP attestation
// report through the Guest-Hypervisor Communication Block.
//
// The arguments represent guest provided data and the VM Communication Key
// (see [SNPSecrets.VPMCK]) payload and index for encrypting the request.
func (b *GHCB) GetAttestationReport(data []byte, key []byte, index int) (r *AttestationReport, err error) {
	if b.addr == 0 {
		return nil, errors.New("invalid instance")
	}

	if len(data) > 64 {
		return nil, errors.New("data length must not exceed %d bytes")
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

	res := &ReportResponse{}

	// update sequence number
	binary.LittleEndian.PutUint64(hdr.SeqNo[:], b.seqNo)

	// fill message data and update header with its size
	copy(msg.Data[:], data)
	msgData := msg.Bytes()
	hdr.MessageSize = uint16(len(msgData))

	if err = seal(key, hdr.SeqNo[0:12], msgData, hdr.Bytes()); err != nil {
		return
	}

	// concatenate updated header and encrypted message
	copy(hdr.AuthTag[:], msgData[len(msgData)-16:])
	req := hdr.Bytes()
	req = append(req, msgData[:len(msgData)-16]...)

	// allocate shared page for guest/hypervisor communication
	addr, shm := dma.Reserve(pageSize, pageSize) // FIXME: clear C-bit
	defer dma.Release(addr)

	// set up GHCB layout
	b.write(SW_EXITCODE, SNP_GUEST_REQUEST)
	b.write(SW_EXITINFO1, 0)
	b.write(SW_EXITINFO2, uint64(addr))

	// trigger NAE event
	b.Yield()

	if err = hdr.unmarshal(shm); err != nil {
		return
	}

	if hdr.Seq() != b.seqNo {
		return nil, errors.New("invalid response header")
	}

	// zero auth AuthTag before unseal
	copy(shm[headerSize:headerSize+32], make([]byte, 32))

	if err = unseal(key, hdr.SeqNo[0:12], shm[headerSize:headerSize+hdr.MessageSize], hdr.Bytes()); err != nil {
		return
	}

	b.seqNo += 2

	if err = res.unmarshal(shm[headerSize : headerSize+hdr.MessageSize]); err != nil {
		return
	}

	// TODO: validate response fields

	return &res.Report, nil
}
