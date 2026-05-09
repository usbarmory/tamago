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
	"unsafe"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

// Google Virtual NIC PCI Device
const (
	PCI_VENDOR = 0x1ae0 // Google, Inc.
	PCI_DEVICE = 0x0042 // Compute Engine Virtual Ethernet [gVNIC]
)

const (
	registersBAR = 0
	msixTableBAR = 1
	doorbellsBAR = 2
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
	DRIVER_VERSION       = 0x1f

	// PCI Revision 00
	ADMINQ_PFN = 0x10

	// PCI Revision 01
	ADMINQ_BASE_ADDRESS_HIGH = 0x20
	ADMINQ_BASE_ADDRESS_LOW  = 0x24
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

	// Region represents the memory region for shared DMA buffers, it is
	// initialized to the global DMA region if unset during [GVE.Init].
	//
	// It can be used to override the global DMA region as needed, for
	// example to specify an unencrypted memory region when running in a
	// Confidential VMs.
	Region *dma.Region

	// PCI memory BARS
	registers uint32
	msixTable uint32
	doorbells uint32

	aq *adminQueue
	rx *rxQueue
	tx *txQueue

	state driverState
}

func (hw *GVE) writeDriverVersion(v string) {
	for i := 0; i < len(v); i++ {
		*(*uint8)(unsafe.Pointer(uintptr(hw.registers + DRIVER_VERSION))) = v[i]
	}
	*(*uint8)(unsafe.Pointer(uintptr(hw.registers + DRIVER_VERSION))) = '\n'
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

	if hw.Region == nil {
		hw.Region = dma.Default()
	}

	hw.registers = uint32(hw.Device.BaseAddress(registersBAR))
	hw.msixTable = uint32(hw.Device.BaseAddress(msixTableBAR))
	hw.doorbells = uint32(hw.Device.BaseAddress(doorbellsBAR))

	if hw.registers&1 != 0 || hw.msixTable&1 != 0 || hw.doorbells&1 != 0 {
		return errors.New("unexpected PCI BAR type, expected memory")
	}

	hw.writeDriverVersion(driverVersion)
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

	// Create TX queue before RX, matching Linux gve_main.c.
	if err = hw.initTxQueue(int(txQPLID)); err != nil {
		return fmt.Errorf("failed to initialize tx queue, %v", err)
	}

	if err = hw.initRxQueue(int(rxQPLID)); err != nil {
		return fmt.Errorf("failed to initialize rx queue, %v", err)
	}

	hw.setupState()
	hw.fillRxRing()

	// Unmask all notification-block doorbells. Linux gve_turnup writes
	// BE32 0 to each ntfy_block's doorbell in BAR2 after queue creation;
	// without this the firmware holds inbound traffic in polling mode.
	hw.unmaskAllIRQs()

	return
}
