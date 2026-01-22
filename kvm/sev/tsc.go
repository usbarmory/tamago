// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sev

import (
	"encoding/binary"
	"fmt"
)

const (
	MSG_TSC_INFO_REQ = 17
	MSG_TSC_INFO_RES = 18
)

// TSCInfo represents an AMD SEV-SNP TSC Information Response
// (SEV Secure Nested Paging Firmware ABI Specification
// Table 39: MSG_TSC_INFO_RES Message Structure).
type TSCInfo struct {
	Status         uint32
	_              uint32
	GuestTSCScale  uint64
	GuestTSCOffset uint64
	TSCFactor      uint32
	_              [100]byte
}

func (r *TSCInfo) unmarshal(buf []byte) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, r)
	return
}

// TSCInfo sends an AMD SEV-SNP guest request for TSC information. The
// arguments represent guest provided request parameters, the VM Communication
// Key (see [SNPSecrets.VMPCK]) payload and index for encrypting the request.
func (b *GHCB) TSCInfo(key []byte, index int) (res *TSCInfo, err error) {
	var buf []byte

	req := make([]byte, 128)
	res = &TSCInfo{}

	if buf, err = b.GuestRequest(index, key, req, MSG_TSC_INFO_REQ); err != nil {
		return
	}

	if err = res.unmarshal(buf); err != nil {
		return nil, fmt.Errorf("could not parse response, %v", err)
	}

	if res.Status != 0 {
		return nil, fmt.Errorf("request error, %#x", res.Status)
	}

	return
}
