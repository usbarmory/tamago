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
	"net"
	"sync"

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
	DEVICE_STATUS = 0x00
	STATUS_LINK   = 1
	STATUS_RESET  = 0

	DRIVER_STATUS        = 0x04
	MAX_TX_QUEUES        = 0x08
	MAX_RX_QUEUES        = 0x0c
	ADMINQ_PFN           = 0x10
	ADMINQ_DOORBELL      = 0x14
	ADMINQ_EVENT_COUNTER = 0x18
)

const (
	pageSize       = 4096
	commandSize    = 64
	adminQueueSize = 4096
	txQueueSize    = 256
	rxQueueSize    = 256
)

// GVE represents a Google Virtual NIC instance.
type GVE struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Interrupt ID
	IRQ int
	// MAC address
	MAC net.HardwareAddr

	// Device represents the probed PCI device.
	Device *pci.Device
	// Info represents the initialized device descriptor.
	Info *DeviceDescriptor

	// PCI memory BARS
	registers uint32
	msixTable uint32
	doorbells uint32

	// Admin Queue (AQ)
	aq *adminQueue
}

// Init initializes a Google Virtual NIC instance.
func (hw *GVE) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Device == nil {
		return errors.New("invalid PCI device")
	}

	hw.registers = uint32(hw.Device.BaseAddress(0))
	hw.doorbells = uint32(hw.Device.BaseAddress(1))

	if hw.registers&1 != 0 || hw.doorbells&1 != 0 {
		return errors.New("unexpected PCI BAR type, expected memory")
	}

	// soft reset
	reg.Write(hw.Base+DEVICE_STATUS, STATUS_RESET)

	if err := hw.initAdminQueue(); err != nil {
		return fmt.Errorf("failed to initialize admin queue, %v", err)
	}

	// query device capabilities
	if err = hw.describeDevice(); err != nil {
		return fmt.Errorf("failed to describe device, %v", err)
	}

	return
}
