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
	"errors"
	"fmt"
)

const (
	MSG_REPORT_REQ = 5
	MSG_REPORT_RSP = 6
)

const reportVersion = 0x5

// ReportRequest represents an AMD SEV-SNP Report Request Message
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

// ReportResponse represents an AMD SEV-SNP Report Request Response
// (SEV Secure Nested Paging Firmware ABI Specification
// Table 25. MSG_REPORT_RSP Message Structure).
type ReportResponse struct {
	Status uint32
	Size   uint32
	_      [24]byte
	Report AttestationReport
}

func (r *ReportResponse) unmarshal(buf []byte) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, r)
	return
}

// AttestationReport represents an AMD SEV-SNP attestation report
// (SEV Secure Nested Paging Firmware ABI Specification
// Table 23. ATTESTATION_REPORT Structure).
type AttestationReport struct {
	Version          uint32
	GuestSVN         uint32
	Policy           uint64
	FamilyID         [16]byte
	ImageID          [16]byte
	VMPL             uint32
	SignatureAlgo    uint32
	CurrentTCB       uint64
	PlatformInfo     uint64
	SignerInfo       uint32
	_                uint32
	ReportData       [64]byte
	Measurement      [48]byte
	HostData         [32]byte
	IDKeyDigest      [48]byte
	AuthorKeyDigest  [48]byte
	ReportID         [32]byte
	ReportIDMA       [32]byte
	ReportedTCB      uint64
	CPUIDFamID       uint8
	CPUIDModID       uint8
	CPUIDStep        uint8
	_                [21]byte
	ChipID           [64]byte
	CommittedTCB     uint64
	CurrentBuild     uint8
	CurrentMinor     uint8
	CurrentMajor     uint8
	_                uint8
	CommittedBuild   uint8
	CommittedMinor   uint8
	CommittedMajor   uint8
	_                uint8
	LaunchTCB        uint64
	LaunchMitVector  uint64
	CurrentMitVector uint64
	_                [152]byte
	Signature        [512]byte
}

// Bytes converts the descriptor structure to byte array format.
func (r *AttestationReport) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r)
	return buf.Bytes()
}

// GetAttestationReport sends a guest request for an AMD SEV-SNP attestation
// report through the Guest-Hypervisor Communication Block.
//
// The arguments represent guest provided data and the VM Communication Key
// (see [SNPSecrets.VMPCK]) payload and index for encrypting the request.
func (b *GHCB) GetAttestationReport(data, key []byte, index int) (r *AttestationReport, err error) {
	var buf []byte

	req := &ReportRequest{
		VMPL:   0,
		KeySel: 0, // sign with VLEK | VCEK
	}

	res := &ReportResponse{}

	// fill message data
	copy(req.Data[:], data)

	if buf, err = b.GuestRequest(index, key, req.Bytes(), MSG_REPORT_REQ); err != nil {
		return
	}

	if err = res.unmarshal(buf); err != nil {
		return nil, fmt.Errorf("could not parse report, %v", err)
	}

	if res.Report.Version != reportVersion {
		err = errors.New("unsupported report version")
	}

	return &res.Report, err
}
