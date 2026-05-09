// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

import (
	"encoding/binary"
	"testing"
)

func TestSeqNo(t *testing.T) {
	tests := []struct {
		flagsSeq uint16
		want     uint8
	}{
		{0x0001, 1},
		{0x0007, 7},
		{0x0000, 0},
		{0xFF03, 3},
		{0x1234, 4},
	}
	for _, tt := range tests {
		got := seqNo(tt.flagsSeq)
		if got != tt.want {
			t.Errorf("seqNo(%#x) = %d, want %d", tt.flagsSeq, got, tt.want)
		}
	}
}

func TestNextSeqNo(t *testing.T) {
	tests := []struct {
		seq  uint8
		want uint8
	}{
		{1, 2},
		{6, 7},
		{7, 1}, // wraps
		{0, 1},
	}
	for _, tt := range tests {
		got := nextSeqNo(tt.seq)
		if got != tt.want {
			t.Errorf("nextSeqNo(%d) = %d, want %d", tt.seq, got, tt.want)
		}
	}
}

func TestTxFIFOAllocOneFrag(t *testing.T) {
	fifo := txFIFOState{head: 0, avail: 4096 * 64, size: 4096 * 64}

	// Small packet fits in one fragment, no padding needed
	pad := fifo.txFIFOPadAllocOneFrag(100)
	if pad != 0 {
		t.Errorf("expected pad=0 for small packet, got %d", pad)
	}

	// Packet at end of FIFO wraps
	fifo.head = fifo.size - 50
	pad = fifo.txFIFOPadAllocOneFrag(100)
	if pad != 50 {
		t.Errorf("expected pad=50 for wrapping packet, got %d", pad)
	}
}

func TestTxFIFOAllocFrags(t *testing.T) {
	fifo := txFIFOState{head: 0, avail: 4096 * 64, size: 4096 * 64}

	// Single fragment
	nfrags := fifo.previewAllocFrags(100)
	if nfrags != 1 {
		t.Errorf("expected 1 fragment, got %d", nfrags)
	}

	// Zero bytes
	nfrags = fifo.previewAllocFrags(0)
	if nfrags != 0 {
		t.Errorf("expected 0 fragments for 0 bytes, got %d", nfrags)
	}

	// Two fragments (wrapping)
	fifo.head = fifo.size - 50
	nfrags = fifo.previewAllocFrags(100)
	if nfrags != 2 {
		t.Errorf("expected 2 fragments for wrapping, got %d", nfrags)
	}
}

func TestTxAllocFIFO(t *testing.T) {
	fifo := txFIFOState{head: 0, avail: 4096 * 64, size: 4096 * 64}
	var iov [txMaxAllocFrags]txIovec

	// Allocate a small packet
	nfrags := fifo.txAllocFIFO(100, &iov)
	if nfrags != 1 {
		t.Fatalf("expected 1 fragment, got %d", nfrags)
	}
	if iov[0].off != 0 {
		t.Errorf("expected offset 0, got %d", iov[0].off)
	}
	if iov[0].length != 100 {
		t.Errorf("expected length 100, got %d", iov[0].length)
	}
	// head should be aligned
	if fifo.head%txFIFOAlign != 0 {
		t.Errorf("expected aligned head, got %d", fifo.head)
	}
}

func TestParseDeviceOptions(t *testing.T) {
	// Build a buffer with 2 device options
	buf := make([]byte, 24)
	// Option 1: GQI-QPL (id=3), length=0, features=0
	binary.BigEndian.PutUint16(buf[0:2], DevOptGqiQPL)
	binary.BigEndian.PutUint16(buf[2:4], 0) // length
	binary.BigEndian.PutUint32(buf[4:8], 0) // features

	// Option 2: DQO-RDA (id=4), length=0, features=0
	binary.BigEndian.PutUint16(buf[8:10], DevOptDqoRDA)
	binary.BigEndian.PutUint16(buf[10:12], 0)
	binary.BigEndian.PutUint32(buf[12:16], 0)

	opts := parseDeviceOptions(buf, 2)
	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}
	if opts[0].OptionID != DevOptGqiQPL {
		t.Errorf("option 0: expected id %d, got %d", DevOptGqiQPL, opts[0].OptionID)
	}
	if opts[1].OptionID != DevOptDqoRDA {
		t.Errorf("option 1: expected id %d, got %d", DevOptDqoRDA, opts[1].OptionID)
	}
}

func TestSupportsGQI_QPL(t *testing.T) {
	opts := []DeviceOption{
		{OptionID: DevOptGqiRawAddr},
		{OptionID: DevOptGqiQPL},
	}
	if !supportsGQI_QPL(opts) {
		t.Error("expected GQI-QPL support to be detected")
	}

	opts = []DeviceOption{
		{OptionID: DevOptGqiRawAddr},
		{OptionID: DevOptDqoRDA},
	}
	if supportsGQI_QPL(opts) {
		t.Error("expected GQI-QPL support to NOT be detected")
	}
}

func TestStatusString(t *testing.T) {
	if s := statusString(COMMAND_PASSED); s != "passed" {
		t.Errorf("expected 'passed', got %q", s)
	}
	if s := statusString(COMMAND_ERROR_INVALID_ARGUMENT); s != "invalid argument" {
		t.Errorf("expected 'invalid argument', got %q", s)
	}
	if s := statusString(0x42); s == "" {
		t.Error("expected non-empty string for unknown status")
	}
}

func TestFillTxPktDesc(t *testing.T) {
	desc := make([]byte, 16)
	fillTxPktDesc(desc, 2, 100, 0x1000, 200)

	if desc[0] != GVE_TXD_STD {
		t.Errorf("expected type %#x, got %#x", GVE_TXD_STD, desc[0])
	}
	if desc[3] != 2 {
		t.Errorf("expected desc_cnt 2, got %d", desc[3])
	}
	if pktLen := binary.BigEndian.Uint16(desc[4:6]); pktLen != 200 {
		t.Errorf("expected pktLen 200, got %d", pktLen)
	}
	if segLen := binary.BigEndian.Uint16(desc[6:8]); segLen != 100 {
		t.Errorf("expected segLen 100, got %d", segLen)
	}
	if segAddr := binary.BigEndian.Uint64(desc[8:16]); segAddr != 0x1000 {
		t.Errorf("expected segAddr 0x1000, got %#x", segAddr)
	}
}

func TestFillTxSegDesc(t *testing.T) {
	desc := make([]byte, 16)
	fillTxSegDesc(desc, 50, 0x2000)

	if desc[0] != GVE_TXD_SEG {
		t.Errorf("expected type %#x, got %#x", GVE_TXD_SEG, desc[0])
	}
	if segLen := binary.BigEndian.Uint16(desc[6:8]); segLen != 50 {
		t.Errorf("expected segLen 50, got %d", segLen)
	}
	if segAddr := binary.BigEndian.Uint64(desc[8:16]); segAddr != 0x2000 {
		t.Errorf("expected segAddr 0x2000, got %#x", segAddr)
	}
}

func TestCloseNilIdempotent(t *testing.T) {
	// A GVE with nil aq should return nil from Close
	hw := &GVE{}
	if err := hw.Close(); err != nil {
		t.Errorf("Close on uninitialized GVE: %v", err)
	}
	// Double close should also be safe
	if err := hw.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}
