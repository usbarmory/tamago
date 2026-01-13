// Google Compute Engine support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// https://www.kernel.org/doc/html/latest/networking/device_drivers/ethernet/google/gve.html
package gvnic

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

// gVNIC configuration registers (Bar0)
const (
	GVE_DEVICE_STATUS   = 0x00
	DEVICE_STATUS_LINK  = 1
	DEVICE_STATUS_RESET = 0

	GVE_DRIVER_STATUS        = 0x04
	GVE_MAX_TX_QUEUES        = 0x08
	GVE_MAX_RX_QUEUES        = 0x0c
	GVE_ADMINQ_PFN           = 0x10
	GVE_ADMINQ_DOORBELL      = 0x14
	GVE_ADMINQ_EVENT_COUNTER = 0x18
)

// Admin queue commands
const (
	GVE_ADMINQ_DESCRIBE_DEVICE = 0x1
)

const (
	commandSize    = 64
	adminQueueSize = 4096
	txQueueSize    = 256
	rxQueueSize    = 256
)

type adminCommand struct {
	Opcode uint32
	Status uint32
	Cmd    [56]byte
}

type adminQueue struct {
	Doorbell uint32
	Counter  uint32

	// internal counter
	cnt int

	// DMA buffer
	addr uint
	buf  []byte
}

func (aq *adminQueue) Push(cmd *adminCommand) (err error) {
	low := aq.cnt * commandSize
	high := low + commandSize

	if _, err = binary.Encode(aq.buf[low:high], binary.BigEndian, cmd); err != nil {
		return
	}

	aq.cnt = (aq.cnt + 1) % (adminQueueSize / commandSize)
	reg.Write(aq.Doorbell, uint32(aq.cnt))

	if !reg.WaitFor(1*time.Second, aq.Counter, 0, 0xff, uint32(aq.cnt)) {
		return errors.New("admin queue timeout")
	}

	return
}

func (hw *GVE) initAdminQueue() (err error) {
	hw.aq = &adminQueue{
		Doorbell: hw.registers + GVE_ADMINQ_DOORBELL,
		Counter:  hw.registers + GVE_ADMINQ_EVENT_COUNTER,
	}

	hw.aq.addr, hw.aq.buf = dma.Reserve(adminQueueSize, 0)

	// set admin queue based address to region page frame number
	reg.Write(hw.Base+GVE_ADMINQ_PFN, uint32(hw.aq.addr>>12))

	return
}

// NIC represents a Google Virtual NIC instance.
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
	reg.Write(hw.Base+GVE_DEVICE_STATUS, DEVICE_STATUS_RESET)

	if err := hw.initAdminQueue(); err != nil {
		return fmt.Errorf("failed to initialize admin queue, %v", err)
	}

	// TODO: WiP

	// Query device capabilities
	//info, err := dev.describeDevice()

	//if err != nil {
	//	return nil, fmt.Errorf("failed to query device, %v", err)
	//}

	return
}
