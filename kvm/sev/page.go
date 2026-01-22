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

	"github.com/usbarmory/tamago/bits"
)

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 7: List of Supported Non-Automatic Events.
const PAGE_STATE_CHANGE = 0x80000010

// SEV-ES Guest-Hypervisor Communication Block Standardization
// Table 9: Page State Change Entry.
const (
	PAGE_SIZE_4K = 0
	PAGE_SIZE_2M = 1
)

type pscHeader struct {
	CurEntry uint16
	EndEntry uint16
	_        uint32
}

func (h *pscHeader) unmarshal(buf []byte) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, h)
	return
}

// Page State Change Entry constants
const (
	// offsets
	entryPageSize    = 56
	entryOperation   = 52
	entryGFN         = 12
	entrycurrentPage = 0

	// page types
	privatePage = 0x0001
	sharedPage  = 0x0002
)

type psc struct {
	Header  pscHeader
	Entries [253]uint64
}

func (r *psc) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r)
	return buf.Bytes()
}

// defined in page.s
func pvalidate(addr uint64, pageSize int, validate bool) (ret uint32)

// PageStateChange requests a page state change to either private/shared
// assignment, as pages are validated/invalidated accordingly any C-Bit
// transition [see SetEncryptedBit] must be performed after/before.
func (b *GHCB) PageStateChange(start uint64, end uint64, pageSize int, private bool) (err error) {
	var size uint64
	var n uint16

	switch pageSize {
	case PAGE_SIZE_4K:
		size = 1 << 12
	case PAGE_SIZE_2M:
		size = 2 << 20
	default:
		return errors.New("invalid page size")
	}

	req := &psc{}

	for gpa := start; gpa < end; gpa += size {
		var entry uint64

		bits.SetN64(&entry, entryGFN, 0xffffffffff, gpa>>12)
		bits.SetTo64(&entry, entryPageSize, pageSize == PAGE_SIZE_2M)

		if private {
			bits.SetN64(&entry, entryOperation, 0b1111, privatePage)
		} else {
			bits.SetN64(&entry, entryOperation, 0b1111, sharedPage)
		}

		if ret := pvalidate(gpa, pageSize, private); ret != 0 {
			return fmt.Errorf("pvalidate error, gpa:%#x %v ret:%d", gpa, private, ret)
		}

		n += 1

		if int(n) > len(req.Entries) {
			return errors.New("range exceeds page state change entries size")
		}

		req.Header.EndEntry = n
		req.Entries[n-1] = entry
	}

	num := req.Header.EndEntry
	buf := req.Bytes()

	copy(b.buf[SharedBuffer:], buf)

	if err = b.Exit(PAGE_STATE_CHANGE, 0, 0, uint64(b.addr)+SharedBuffer); err != nil {
		return
	}

	if err = req.Header.unmarshal(b.buf[SharedBuffer : SharedBuffer+8]); err != nil {
		return fmt.Errorf("could not parse response header, %v", err)
	}

	if req.Header.CurEntry <= num || req.Header.EndEntry > num {
		return fmt.Errorf("incomplete psc (num:%d cur:%d end:%d info2:%#x)", num, req.Header.CurEntry, req.Header.EndEntry, b.read(SW_EXITINFO2))
	}

	return
}
