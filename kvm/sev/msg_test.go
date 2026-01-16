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
	"encoding/hex"
	"strings"
	"testing"
)

const testKey = `
fe246add1ac33915ee4bf86e869b0ba5
72bd33f26e883404d7c70cca0b40b4b3
`

const testEncReq = `
ee858c71f018162db682ecee7244a4cd
00000000000000000000000000000000
01000000000000000000000000000000
01016000050160000000000001000000
00000000000000000000000000000000
00000000000000000000000000000000
a1a156d62d7eb58a8677e5fbf38aad4c
c3dc9322e16ed02e324e740580b76bf7
c671cb56fb030af6b345692677e7231e
f2100d39ef12b866701dadcfa95aa766
920616db36aa883089e7ecabd1640b9a
772a788858e9bfd7f122088f163af2ad
`

func TestAEAD(t *testing.T) {
	key, err := hex.DecodeString(strings.ReplaceAll(testKey, "\n", ``))

	if err != nil {
		t.Fatal(err)
	}

	encMsg, err := hex.DecodeString(strings.ReplaceAll(testEncReq, "\n", ``))

	if err != nil {
		t.Fatal(err)
	}

	b := &GHCB{
		seqNo: 1,
	}

	hdr := &MessageHeader{
		Algo:           AES_256_GCM,
		HeaderVersion:  headerVersion,
		HeaderSize:     headerSize,
		MessageType:    MSG_REPORT_REQ,
		MessageVersion: messageVersion,
		VMPCK:          1,
	}

	req := &ReportRequest{
		VMPL: 1,
	}

	hdr.SetSeq(1)

	// fill message data
	data := make([]byte, 96)
	copy(req.Data[:], data)

	sealedMsg, err := b.sealMessage(hdr, req.Bytes(), key)

	if err != nil {
		t.Errorf("could not seal message, %v", err)
	}

	if !bytes.Equal(sealedMsg, encMsg) {
		t.Errorf("encrypted message mismatch:\n%s\n%s", hex.Dump(sealedMsg), hex.Dump(encMsg))
	}
}
