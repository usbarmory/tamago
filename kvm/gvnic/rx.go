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
	"math/bits"

	"github.com/usbarmory/tamago/internal/reg"
)

// advanceRxSlot consumes one RX descriptor: clears it, bumps the expected
// sequence number, and advances the consumer / fill counters.
func (hw *GVE) advanceRxSlot(desc []byte) {
	clear(desc)
	hw.state.rxSeq = nextSeqNo(hw.state.rxSeq)
	hw.state.rxCnt++
	hw.state.rxFill++
}

// Receive reads the next packet into buf. Satisfies go-net NetworkDevice.
// Returns 0 if no completed descriptor is available. Large packets delivered
// across multiple descriptors (GVE_RXF_PKT_CONT) are reassembled into buf in
// place, with no per-packet allocation. GVE_RX_PAD applies only to the first
// fragment.
func (hw *GVE) Receive(buf []byte) (int, error) {
	if hw.rx == nil || hw.rx.qpl == nil {
		return 0, nil
	}

	qpl := hw.rx.qpl
	n := 0
	isFirst := true

	for {
		idx := hw.state.rxCnt & hw.state.rxMask
		descOff := int(idx) * rxDescSize
		desc := hw.rx.desc[descOff : descOff+rxDescSize]

		// Read flags_seq (BE16 at offset 62)
		flagsSeq := binary.BigEndian.Uint16(desc[62:64])
		if seqNo(flagsSeq) != hw.state.rxSeq {
			return 0, nil
		}

		pktLen := int(binary.BigEndian.Uint16(desc[60:62]))
		isLast := flagsSeq&GVE_RXF_PKT_CONT == 0
		hasErr := flagsSeq&GVE_RXF_ERR != 0

		if hasErr {
			hw.advanceRxSlot(desc)
			if isLast {
				hw.ringRxDoorbell()
				return 0, nil
			}
			n = 0
			isFirst = false
			continue
		}

		// Copy fragment from QPL pool directly into buf.
		pad := 0
		if isFirst {
			pad = GVE_RX_PAD
		}
		bufOff := int(idx) * pageSize
		fragStart := bufOff + pad
		fragEnd := bufOff + pad + pktLen
		if fragEnd > bufOff+pageSize {
			fragEnd = bufOff + pageSize
		}
		if fragEnd > fragStart && n < len(buf) {
			n += copy(buf[n:], qpl[fragStart:fragEnd])
		}

		hw.advanceRxSlot(desc)

		if isLast {
			hw.ringRxDoorbell()
			return n, nil
		}
		isFirst = false
	}
}

// fillRxRing arms the RX ring after queue creation by writing the initial
// fill count to the RX doorbell.
func (hw *GVE) fillRxRing() {
	if hw.rx == nil {
		return
	}
	hw.state.rxFill = hw.Info.RxQueueEntries
	hw.ringRxDoorbell()
}

// ringRxDoorbell writes the current fill count to the RX doorbell.
func (hw *GVE) ringRxDoorbell() {
	if hw.rx == nil || hw.rx.Resources == nil {
		return
	}
	reg.Write(hw.doorbells+hw.rx.Resources.DBIndex*4, bits.ReverseBytes32(uint32(hw.state.rxFill)))
}
