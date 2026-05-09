// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

// QPL identifiers used by the driver. These match the literals used in
// GVE.Init for initRxQueue / initTxQueue.
const (
	txQPLID uint32 = 1
	rxQPLID uint32 = 2
)

// driverState holds runtime state required by the driver beyond the
// fields defined on GVE. It is embedded as a value (not a pointer) on
// GVE so its fields are zero-valued from the moment the GVE struct is
// constructed — describeDevice and configureDeviceResources both assign
// to hw.state.X during Init before setupState runs.
type driverState struct {
	options []DeviceOption

	// TX QPL FIFO management
	txFIFO    txFIFOState
	txPending []txBufferState
	txReq     uint32
	txDone    uint32

	// Cached masks (RxQueueEntries-1 / TxQueueEntries-1) for the hot path.
	rxMask uint32
	txMask uint32

	// RX state
	rxSeq  uint8
	rxCnt  uint32 // consumer index
	rxFill uint16 // fill count for doorbell

	// counter array for TX completion
	counterArray []byte

	// IRQ doorbell array — populated by configureDeviceResources, used by
	// unmaskAllIRQs to write BE32 0 into each notification block's doorbell
	// slot in BAR2 after queue creation. Without this unmask the device
	// holds inbound traffic even in polling mode (Linux gve_turnup; prior
	// `third_party/gvnic` Phase 9 v6 observation).
	irqDBArray []byte
	numIRQDBs  uint32
}

// setupState fills in the parts of driverState that depend on values
// learned during Init (rx/tx queues, device descriptor). Fields written
// during Init's earlier phases (counterArray, options) are already in
// place because state is a value, not a heap allocation.
func (hw *GVE) setupState() {
	txDataSize := uint32(hw.Info.TxPagesPerQpl) * pageSize
	hw.state.txFIFO = txFIFOState{
		avail: txDataSize,
		size:  txDataSize,
	}
	hw.state.txPending = make([]txBufferState, hw.Info.TxQueueEntries)

	hw.state.rxMask = uint32(hw.Info.RxQueueEntries - 1)
	hw.state.txMask = uint32(hw.Info.TxQueueEntries - 1)
	hw.state.rxSeq = 1
}
