// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package svm

import (
	"errors"
)

// GetAttestationReport sends a guest request for an AMD SEV-SNP attestation
// report through the Guest-Hypervisor Communication Block.
//
// The arguments represent guest provided data and the VM Communication Key
// (see [SNPSecrets.VPMCK]) payload and index for encrypting the request.
func (b *GHCB) GetAttestationReport(data, key []byte, index int) (r *AttestationReport, err error) {
	var msg []byte

	if b.addr == 0 {
		return nil, errors.New("invalid instance")
	}

	if len(data) > 64 {
		return nil, errors.New("data length must not exceed %d bytes")
	}

	// SEV Secure Nested Paging Firmware ABI Specification
	// 8.26 SNP_GUEST_REQUEST

	hdr := &MessageHeader{
		Algo:           AES_256_GCM,
		HeaderVersion:  headerVersion,
		HeaderSize:     headerSize,
		MessageType:    MSG_REPORT_REQ,
		MessageVersion: messageVersion,
		VMPCK:          uint8(index),
	}

	req := &ReportRequest{
		VMPL:   0,
		KeySel: 0,
	}

	res := &ReportResponse{}

	// fill message data
	copy(req.Data[:], data)
	hdr.MessageSize = uint16(len(data))

	// encrypt request message
	if msg, err = b.sealMessage(hdr, req.Bytes(), key); err != nil {
		return
	}

	// allocate shared page for guest/hypervisor communication
	addr, shm := b.SharedMemory.Reserve(pageSize, pageSize)
	defer b.SharedMemory.Release(addr)

	// set up GHCB layout and yield to hypervisor
	copy(shm, msg)
	b.Exit(SNP_GUEST_REQUEST, 0, uint64(addr))

	// decode response header
	if err = hdr.unmarshal(shm); err != nil {
		return
	}

	// decrypt response message
	if msg, err = b.openMessage(hdr, shm[headerSize:headerSize+hdr.MessageSize], key); err != nil {
		return
	}

	// TODO: validate response fields
	err = res.unmarshal(msg)

	return &res.Report, nil
}
