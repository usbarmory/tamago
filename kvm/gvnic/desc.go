// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

// GVE_RX_PAD is the number of padding bytes before the ethernet frame in RX
// buffers. gVNIC adds this so both DMA and L3/L4 headers are aligned.
const GVE_RX_PAD = 2

// Descriptor and ring layout sizes.
const (
	rxDescSize   = 64 // sizeof(rxDesc)
	txDescSize   = 16 // sizeof(txDesc)
	irqDBStride  = 64 // ntfy-block doorbell stride in the IRQ DB array
	rxPacketSize = pageSize / 2
)

// TX descriptor type flags from gve_desc.h.
const (
	GVE_TXD_STD = 0x0 << 4 // Standard with host address
	GVE_TXD_TSO = 0x1 << 4 // TSO with host address
	GVE_TXD_SEG = 0x2 << 4 // Segment with host address
	GVE_TXD_MTD = 0x3 << 4 // Metadata

	GVE_TXF_L4CSUM = 1 << 0 // Need L4 checksum offload
)

// RX descriptor flags from gve_desc.h.
const (
	GVE_RXF_FRAG     = 1 << (3 + 3)  // IP fragment
	GVE_RXF_IPV4     = 1 << (3 + 4)  // IPv4
	GVE_RXF_IPV6     = 1 << (3 + 5)  // IPv6
	GVE_RXF_TCP      = 1 << (3 + 6)  // TCP
	GVE_RXF_UDP      = 1 << (3 + 7)  // UDP
	GVE_RXF_ERR      = 1 << (3 + 8)  // Packet error
	GVE_RXF_PKT_CONT = 1 << (3 + 10) // Multi-fragment packet
)

// GVE IRQ doorbell bits from gve_desc.h.
const (
	GVE_IRQ_ACK   = 1 << 31
	GVE_IRQ_MASK  = 1 << 30
	GVE_IRQ_EVENT = 1 << 29
)

// seqNo extracts the 3-bit sequence number from flags_seq (host-endian).
func seqNo(flagsSeq uint16) uint8 {
	return uint8(flagsSeq & 0x7)
}

// nextSeqNo advances the GQI sequence number (1-7, wrapping).
func nextSeqNo(seq uint8) uint8 {
	next := seq + 1
	if next == 8 {
		return 1
	}
	return next
}

// Device option IDs from gve_adminq.h.
const (
	DevOptGqiRawAddr  = 0x1
	DevOptGqiRDA      = 0x2
	DevOptGqiQPL      = 0x3
	DevOptDqoRDA      = 0x4
	DevOptModifyRing  = 0x6
	DevOptDqoQPL      = 0x7
	DevOptJumboFrames = 0x8
	DevOptBufferSizes = 0xa
	DevOptFlowSteer   = 0xb
)

// DeviceOption represents a single option entry following the device
// descriptor (struct gve_device_option, 8 bytes).
type DeviceOption struct {
	OptionID             uint16
	OptionLength         uint16
	RequiredFeaturesMask uint32
}
