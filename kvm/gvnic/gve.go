// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// https://www.kernel.org/doc/html/latest/networking/device_drivers/ethernet/google/gve.html
package gvnic

import (
	"errors"
	"fmt"
	"math/bits"
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

// Google Virtual NIC PCI Device
const (
	PCI_VENDOR = 0x1ae0 // Google, Inc.
	PCI_DEVICE = 0x0042 // Compute Engine Virtual Ethernet [gVNIC]
)

// gVNIC configuration registers (Bar0)
const (
	DEVICE_STATUS              = 0x00
	DEVICE_STATUS_LINK  uint32 = 1
	DEVICE_STATUS_RESET uint32 = 0

	DRIVER_STATUS              = 0x04
	DRIVER_STATUS_RESET uint32 = 0
	DRIVER_STATUS_RUN   uint32 = 1

	MAX_TX_QUEUES        = 0x08
	MAX_RX_QUEUES        = 0x0c
	ADMINQ_DOORBELL      = 0x14
	ADMINQ_EVENT_COUNTER = 0x18

	// PCI Revision 00
	ADMINQ_PFN = 0x10

	// PCI Revision 01
	ADMINQ_BASE_ADDRESS_LOW  = 0x20
	ADMINQ_BASE_ADDRESS_HIGH = 0x24
	ADMINQ_LENGTH            = 0x28
)

// GVE represents a Google Virtual NIC instance.
type GVE struct {
	sync.Mutex

	// Controller index
	Index int
	// Interrupt ID
	IRQ int

	// Device represents the probed PCI device.
	Device *pci.Device
	// Info represents the initialized device descriptor.
	Info *DeviceDescriptor

	// QueueRegion is an optional memory page for persistent allocation of
	// the shared admin queue. It must be set to override the global DMA
	// region with unencrypted memory when running in a Confidential VM.
	QueueRegion *dma.Region

	// CommandRegion is an optional memory page for on-demand allocation of
	// shared command buffers. It must be set to override the global DMA
	// region with unencrypted memory when running in a Confidential VM.
	CommandRegion *dma.Region

	// PCI memory BARS
	registers uint32
	msixTable uint32
	doorbells uint32

	// Admin Queue (AQ)
	aq *adminQueue
}

func (hw *GVE) set(off uint32, val any) {
	switch v := val.(type) {
	case uint32:
		reg.Write(hw.registers+off, bits.ReverseBytes32(v))
	case uint16:
		reg.Write16(hw.registers+off, bits.ReverseBytes16(v))
	}
}

// Init initializes a Google Virtual NIC instance.
func (hw *GVE) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Device == nil {
		return errors.New("invalid GVE instance")
	}

	if hw.QueueRegion == nil {
		hw.QueueRegion = dma.Default()
	}

	if hw.CommandRegion == nil {
		hw.CommandRegion = dma.Default()
	}

	hw.registers = uint32(hw.Device.BaseAddress(0))
	hw.doorbells = uint32(hw.Device.BaseAddress(1))

	if hw.registers&1 != 0 || hw.doorbells&1 != 0 {
		return errors.New("unexpected PCI BAR type, expected memory")
	}

	hw.set(DEVICE_STATUS, uint32(DEVICE_STATUS_RESET))

	if err := hw.initAdminQueue(); err != nil {
		return fmt.Errorf("failed to initialize admin queue, %v", err)
	}

	hw.set(DRIVER_STATUS, uint32(DRIVER_STATUS_RUN))

	if err = hw.describeDevice(); err != nil {
		return fmt.Errorf("failed to describe device, %v", err)
	}

	if err = hw.configureDeviceResources(); err != nil {
		return fmt.Errorf("failed to configure device resources, %v", err)
	}

	return
}
