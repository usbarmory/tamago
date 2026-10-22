// NVM Express driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package nvme implements a driver for NVM Express (NVMe) controllers attached
// over PCI Express, adopting the following reference specifications:
//   - NVM Express Base Specification, revision 2.0
//   - NVM Express NVM Command Set Specification, revision 1.0
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package nvme

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

const (
	PCI_CLASS    = 0x01 // Mass Storage
	PCI_SUBCLASS = 0x08 // Non-Volatile Memory
	PCI_PROG_IF  = 0x02 // NVMe programming interface
)

// Controller registers (BAR0)
// (NVM Express Base Specification 2.0 - 3.1.3 Controller Registers).
const (
	CAP        = 0x00 // Controller Capabilities
	CAP_MPSMIN = 48
	CAP_CSS    = 37
	CAP_DSTRD  = 32
	CAP_TO     = 24
	CAP_MQES   = 0

	VS     = 0x08 // Version
	VS_MJR = 16
	VS_MNR = 8
	VS_TER = 0

	INTMS = 0x0c // Interrupt Mask Set
	INTMC = 0x10 // Interrupt Mask Clear

	CC        = 0x14 // Controller Configuration
	CC_IOCQES = 20
	CC_IOSQES = 16
	CC_SHN    = 14
	CC_AMS    = 11
	CC_MPS    = 7
	CC_CSS    = 4
	CC_EN     = 0

	CSTS      = 0x1c // Controller Status
	CSTS_SHST = 2
	CSTS_CFS  = 1
	CSTS_RDY  = 0

	AQA      = 0x24 // Admin Queue Attributes
	AQA_ACQS = 16
	AQA_ASQS = 0

	ASQ = 0x28 // Admin Submission Queue Base Address
	ACQ = 0x30 // Admin Completion Queue Base Address

	DOORBELL_BASE = 0x1000
)

// Register field values
const (
	// CAP.CSS: NVM Command Set support (bit position within the field)
	CSS_NVM = 0
	// CAP.TO: Timeout unit
	TO_UNIT = 500 * time.Millisecond

	// CC.CSS: NVM Command Set
	CC_CSS_NVM = 0b000
	// CC.MPS: 4 KiB memory page size (2 ^ (12 + MPS))
	CC_MPS_4K = 0
	// CC.AMS: Round Robin arbitration
	CC_AMS_RR = 0b000
	// CC.SHN: Normal shutdown notification
	CC_SHN_NORMAL = 0b01
	// CC.IOSQES: 64 byte submission queue entries (2 ^ 6)
	CC_IOSQES_64 = 6
	// CC.IOCQES: 16 byte completion queue entries (2 ^ 4)
	CC_IOCQES_16 = 4

	// CSTS.SHST: Shutdown processing complete
	CSTS_SHST_COMPLETE = 0b10
)

const (
	adminQueueDepth   = 32
	defaultQueueDepth = 256
	defaultQueues     = 1

	pageSize         = 4096
	maxTransferPages = pageSize / 8
)

// Timeout is the default command completion timeout.
const Timeout = 5 * time.Second

// NVMe represents an NVM Express controller instance.
type NVMe struct {
	sync.Mutex

	// Controller index (placeholder for caller use)
	Index int

	// Vector is the first LAPIC interrupt vector used by the controller
	// (required): the admin queue is signaled on Vector and I/O queue n
	// on Vector+n, see [NVMe.Vectors] for the number of vectors in use.
	//
	// Interrupts are delivered to the bootstrap processor and must be
	// forwarded by the application interrupt service routine to
	// [NVMe.ServiceInterrupt], which must be running before [NVMe.Init] is
	// invoked. Vectors must lie between [MinVector] and [MaxVector].
	Vector int

	// Device represents the probed PCI device.
	Device *pci.Device

	// Info holds the Identify Controller data, set on [NVMe.Init].
	Info *ControllerInfo

	// Region represents the memory region for queues and DMA buffers, it is
	// initialized to the global DMA region if unset during [NVMe.Init].
	//
	// It can be used to override the global DMA region as needed, for
	// example to specify an unencrypted memory region when running in
	// Confidential VMs.
	Region *dma.Region

	// CPU is the processor instance used to map the controller registers,
	// wherever firmware placed them, as device memory (required, see
	// [pci.Device.MapBAR]).
	CPU *amd64.CPU

	// Queues is the number of I/O queue pairs to request (default 1), the
	// controller or interrupt vector range may allow fewer, see
	// [NVMe.IOQueues] for the number in use.
	Queues int

	// QueueDepth is the number of entries for each I/O queue (default
	// 256), it is capped to the controller maximum.
	QueueDepth int

	// Timeout is the command completion timeout (default [Timeout]).
	Timeout time.Duration

	// Namespaces holds the active namespaces, set on [NVMe.Init].
	Namespaces []*Namespace

	// controller state
	bar     uint64
	cap     uint64
	dstrd   uint64
	maxXfer int

	mqes  int
	nq    int
	admin *queue
	io    []*queue
	next  atomic.Uint32

	msix *pci.CapabilityMSIX
	irqs atomic.Pointer[[]chan struct{}]
}

// doorbell returns the submission queue tail (cq == false) or completion
// queue head (cq == true) doorbell address for a queue identifier.
func (hw *NVMe) doorbell(qid uint16, cq bool) uint64 {
	n := 2 * uint64(qid)

	if cq {
		n += 1
	}

	return hw.bar + DOORBELL_BASE + n*(4<<hw.dstrd)
}

func (hw *NVMe) waitStatus(pos int, mask int, val uint32, timeout time.Duration) error {
	start := time.Now()

	for {
		csts := reg.Read64As32(hw.bar + CSTS)

		if csts == 0xffffffff {
			return errors.New("controller not responding")
		}

		if bits.GetN(&csts, pos, mask) == val {
			return nil
		}

		if bits.Get(&csts, CSTS_CFS) {
			return errors.New("controller fatal status")
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for CSTS (%#x)", csts)
		}

		time.Sleep(time.Millisecond)
	}
}

func (hw *NVMe) initPCI() (err error) {
	// map the registers and MSI-X table where firmware placed them

	if hw.bar, _, err = hw.Device.MapBAR(hw.CPU, 0); err != nil {
		return
	}

	if hw.msix, err = hw.Device.MapMSIX(hw.CPU); err != nil {
		return
	}

	// enable memory decoding and bus mastering (DMA)
	hw.Device.SetCommand(pci.COMMAND_MEMORY_SPACE, true)
	hw.Device.SetCommand(pci.COMMAND_BUS_MASTER, true)

	// disable legacy INTx as MSI-X is used
	hw.Device.SetCommand(pci.COMMAND_INTERRUPT_DISABLE, true)

	return
}

func (hw *NVMe) initState() (err error) {
	hw.cap = reg.Read64(hw.bar + CAP)

	if hw.cap == (1<<64)-1 {
		return errors.New("controller not responding")
	}

	hw.dstrd = bits.GetN64(&hw.cap, CAP_DSTRD, 0xf)
	css := bits.GetN64(&hw.cap, CAP_CSS, 0xff)

	if !bits.Get64(&css, CSS_NVM) {
		return errors.New("NVM command set not supported")
	}

	if bits.GetN64(&hw.cap, CAP_MPSMIN, 0xf) > CC_MPS_4K {
		return errors.New("4 KiB memory page size not supported")
	}

	hw.mqes = int(bits.GetN64(&hw.cap, CAP_MQES, 0xffff)) + 1
	hw.QueueDepth = min(hw.QueueDepth, hw.mqes)

	return
}

func (hw *NVMe) initAdminQueue() (err error) {
	var aqa uint32

	if hw.admin, err = hw.newQueue(0, min(adminQueueDepth, hw.mqes)); err != nil {
		return
	}

	bits.SetN(&aqa, AQA_ACQS, 0xfff, uint32(hw.admin.depth-1))
	bits.SetN(&aqa, AQA_ASQS, 0xfff, uint32(hw.admin.depth-1))
	reg.Write64As32(hw.bar+AQA, aqa)
	reg.Write64(hw.bar+ASQ, uint64(hw.admin.sqAddr))
	reg.Write64(hw.bar+ACQ, uint64(hw.admin.cqAddr))

	return
}

func (hw *NVMe) initControllerConfiguration() (err error) {
	var cc uint32

	bits.SetN(&cc, CC_IOCQES, 0xf, CC_IOCQES_16)
	bits.SetN(&cc, CC_IOSQES, 0xf, CC_IOSQES_64)
	bits.SetN(&cc, CC_AMS, 0b111, CC_AMS_RR)
	bits.SetN(&cc, CC_MPS, 0xf, CC_MPS_4K)
	bits.SetN(&cc, CC_CSS, 0b111, CC_CSS_NVM)
	bits.Set(&cc, CC_EN)
	reg.Write64As32(hw.bar+CC, cc)

	return
}

// Init initializes an NVMe controller instance: the controller is reset and
// enabled, I/O queues are created and active namespaces identified.
func (hw *NVMe) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Device == nil || hw.CPU == nil {
		return errors.New("invalid NVMe instance")
	}

	if hw.Region == nil {
		hw.Region = dma.Default()
	}

	if hw.nq = hw.Queues; hw.nq <= 0 {
		hw.nq = defaultQueues
	}

	if hw.QueueDepth <= 0 {
		hw.QueueDepth = defaultQueueDepth
	}

	// discard any previous initialization
	stale := hw.discard()

	if err = hw.initPCI(); err != nil {
		return
	}

	if err = hw.initState(); err != nil {
		return
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	readyTimeout := time.Duration(bits.GetN64(&hw.cap, CAP_TO, 0xff)) * TO_UNIT

	if readyTimeout == 0 {
		readyTimeout = hw.Timeout
	}

	// reset controller
	if cc := reg.Read64As32(hw.bar + CC); bits.Get(&cc, CC_EN) {
		reg.Write64As32(hw.bar+CC, 0)
	}

	if err = hw.waitStatus(CSTS_RDY, 1, 0, readyTimeout); err != nil {
		return fmt.Errorf("controller reset failed, %v", err)
	}

	for _, q := range stale {
		q.release()
	}

	if err = hw.initMSIX(); err != nil {
		return fmt.Errorf("interrupt configuration failed, %v", err)
	}

	if err = hw.initAdminQueue(); err != nil {
		return
	}

	if err = hw.initControllerConfiguration(); err != nil {
		return
	}

	if err = hw.waitStatus(CSTS_RDY, 1, 1, readyTimeout); err != nil {
		return fmt.Errorf("controller enable failed, %v", err)
	}

	if err = hw.identifyController(); err != nil {
		return fmt.Errorf("identify controller failed, %v", err)
	}

	if err = hw.initIOQueues(); err != nil {
		return fmt.Errorf("I/O queue creation failed, %v", err)
	}

	if err = hw.identifyNamespaces(); err != nil {
		return fmt.Errorf("identify namespaces failed, %v", err)
	}

	return
}

// Version returns the controller NVMe specification version.
func (hw *NVMe) Version() (major, minor, tertiary int) {
	if hw.bar == 0 {
		return
	}

	vs := reg.Read64As32(hw.bar + VS)

	major = int(bits.GetN(&vs, VS_MJR, 0xffff))
	minor = int(bits.GetN(&vs, VS_MNR, 0xff))
	tertiary = int(bits.GetN(&vs, VS_TER, 0xff))

	return
}

// MaxTransfer returns the maximum number of bytes transferred by a single
// command, larger requests are automatically split.
func (hw *NVMe) MaxTransfer() int {
	return hw.maxXfer
}

// IOQueues returns the number of I/O queue pairs in use.
func (hw *NVMe) IOQueues() int {
	return len(hw.io)
}

// Shutdown performs a controller shutdown, flushing volatile caches to media,
// the instance must be re-initialized with [NVMe.Init] before further use.
func (hw *NVMe) Shutdown() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.bar == 0 {
		return errors.New("controller not initialized")
	}

	// take all I/O queues to prevent concurrent use
	for _, q := range hw.io {
		q.Lock()
		defer q.Unlock()

		q.failed = errors.New("controller shut down")
	}

	cc := reg.Read64As32(hw.bar + CC)
	bits.SetN(&cc, CC_SHN, 0b11, CC_SHN_NORMAL)
	reg.Write64As32(hw.bar+CC, cc)

	return hw.waitStatus(CSTS_SHST, 0b11, CSTS_SHST_COMPLETE, hw.Timeout)
}
