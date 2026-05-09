// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

import (
	"runtime"
)

// Start1 performs one iteration of TX completion processing for consumers
// that drive their own combined RX+poll loop.
func (hw *GVE) Start1() {
	hw.Lock()
	defer hw.Unlock()

	hw.cleanTxDone()
}

// Start runs the internal RX/TX polling loop forever. Packet retrieval is
// via Receive(buf) from the consumer.
func (hw *GVE) Start() {
	idle := 0

	for {
		hw.Lock()
		hw.cleanTxDone()
		hw.Unlock()

		idle++
		if idle >= 64 {
			runtime.Gosched()
			idle = 0
		}
	}
}

// Close performs ordered teardown of the device: destroy queues, unregister
// page lists, deconfigure resources, and reset driver status. Errors during
// teardown are ignored to allow best-effort cleanup.
func (hw *GVE) Close() error {
	hw.Lock()
	defer hw.Unlock()

	if hw.aq == nil {
		return nil
	}

	if hw.tx != nil {
		_ = hw.destroyTxQueue(uint32(hw.Index))
	}
	if hw.rx != nil {
		_ = hw.destroyRxQueue(uint32(hw.Index))
	}
	if hw.tx != nil {
		_ = hw.unregisterPageList(txQPLID)
	}
	if hw.rx != nil {
		_ = hw.unregisterPageList(rxQPLID)
	}
	_ = hw.deconfigureDeviceResources()

	hw.set(DRIVER_STATUS, uint32(DRIVER_STATUS_RESET))

	hw.aq = nil
	hw.tx = nil
	hw.rx = nil

	return nil
}
