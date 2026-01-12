// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"errors"
	"fmt"
)

// GetAttestationReport sends a guest request for an AMD SEV-SNP attestation
// report through the Guest-Hypervisor Communication Block.
//
// The arguments represent guest provided data and the VM Communication Key
// (see [SNPSecrets.VPMCK]) payload and index for encrypting the request.
func (b *GHCB) GetAttestationReport(data, key []byte, index int) (r *AttestationReport, err error) {
	var msg []byte

	if b.RequestPage == nil {
		return nil, errors.New("invalid instance, no request page")
	}

	if len(data) > 64 {
		return nil, errors.New("data length must not exceed 64 bytes")
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
		KeySel: 0, // sign with VLEK | VCEK
	}

	res := &ReportResponse{}

	// fill message data
	copy(req.Data[:], data)
	hdr.MessageSize = uint16(len(data))

	// encrypt request message
	if msg, err = b.sealMessage(hdr, req.Bytes(), key); err != nil {
		return
	}

	reqAddr, reqBuf := b.RequestPage.Reserve(pageSize, pageSize)
	defer b.RequestPage.Release(reqAddr)

	resAddr := reqAddr
	resBuf := reqBuf

	// re-use request buffer if no response page has been provided
	if b.ResponsePage != nil {
		resAddr, resBuf = b.ResponsePage.Reserve(pageSize, pageSize)
		defer b.RequestPage.Release(resAddr)
	}

	copy(reqBuf, msg)

	// yield to hypervisor
	if err = b.Exit(SNP_GUEST_REQUEST, uint64(reqAddr), uint64(resAddr)); err != nil {
		return
	}

	// copy response buffer as soon as possible as GHCB might overwrite it
	buf := make([]byte, pageSize)
	copy(buf, resBuf)

	if err = hdr.unmarshal(buf); err != nil {
		return nil, fmt.Errorf("could not parse response header, %v", err)
	}

	if msg, err = b.openMessage(hdr, buf[headerSize:headerSize+hdr.MessageSize], key); err != nil {
		return nil, fmt.Errorf("could not decrypt response message, %v", err)
	}

	if err = res.unmarshal(msg); err != nil {
		return nil, fmt.Errorf("could not parse report, %v", err)
	}

	if res.Report.Version != reportVersion {
		err = errors.New("unsupported report version")
	}

	return &res.Report, err
}
