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
	MSG_REPORT_REQ = 5
	MSG_REPORT_RSP = 6
)

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
	Version         uint32
	GuestSVN        uint32
	Policy          uint64
	FamilyID        [16]byte
	ImageID         [16]byte
	VMPL            uint32
	SignatureAlgo   uint32
	PlatformVersion uint64
	PlatformInfo    uint64
	Flags           uint32
	Reserved0       uint32
	Measurement     [48]byte
	HostData        [32]byte
	IDKeyDigest     [48]byte
	AuthorKeyDigest [48]byte
	ReportData      [64]byte
	Reserved1       [192]byte
	ChipID          [64]byte
	CommittedTCB    uint64
	CurrentBuild    uint8
	CurrentMinor    uint8
	CurrentMajor    uint8
	Reserved2       uint8
	CommittedBuild  uint8
	CommittedMinor  uint8
	CommittedMajor  uint8
	Reserved3       uint8
	LaunchTCB       uint64
	Reserved4       [168]byte
	Signature       [512]byte
}
