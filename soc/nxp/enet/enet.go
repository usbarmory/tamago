// NXP 10/100-Mbps Ethernet MAC (ENET)
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package enet implements a driver for NXP Ethernet controllers adopting the
// following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package enet

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// ENET registers
const (
	// p879, 22.5 Memory map/register definition, IMX6ULLRM

	ENETx_EIR  = 0x0004
	ENETx_EIMR = 0x0008

	ENETx_RDAR  = 0x0010
	RDAR_ACTIVE = 24

	ENETx_TDAR  = 0x0014
	TDAR_ACTIVE = 24

	ENETx_ECR   = 0x0024
	ECR_DBSWP   = 8
	ECR_EN1588  = 5
	ECR_ETHEREN = 1
	ECR_RESET   = 0

	ENETx_MMFR = 0x0040
	MMFR_ST    = 30
	MMFR_OP    = 28
	MMFR_PA    = 23
	MMFR_RA    = 18
	MMFR_TA    = 16
	MMFR_DATA  = 0

	ENETx_MSCR     = 0x0044
	MSCR_HOLDTIME  = 8
	MSCR_MII_SPEED = 1

	ENETx_MIB = 0x0064
	MIB_DIS   = 31

	ENETx_RCR     = 0x0084
	RCR_MAX_FL    = 16
	RCR_RMII_MODE = 8
	RCR_FCE       = 5
	RCR_MII_MODE  = 2
	RCR_LOOP      = 0

	ENETx_TCR = 0x00c4
	TCR_FDEN  = 2

	ENETx_PALR = 0x00e4
	ENETx_PAUR = 0x00e8
	ENETx_RDSR = 0x0180
	ENETx_TDSR = 0x0184
	ENETx_MRBR = 0x0188
	ENETx_FTRL = 0x01b0

	ENETx_RACC   = 0x01c4
	RACC_LINEDIS = 6
)

// ENET interrupt events
const (
	// p889, 22.5.1 Interrupt Event Register (ENETx_EIR),  IMX6ULLRM
	// p891, 22.5.2 Interrupt Mask  Register (ENETx_EIMR), IMX6ULLRM

	IRQ_BABR     = 30
	IRQ_BABT     = 29
	IRQ_GRA      = 28
	IRQ_TXF      = 27
	IRQ_TXB      = 26
	IRQ_RXF      = 25
	IRQ_RXB      = 24
	IRQ_MII      = 23
	IRQ_EBERR    = 22
	IRQ_LC       = 21
	IRQ_RL       = 20
	IRQ_UN       = 19
	IRQ_PLR      = 18
	IRQ_WAKEUP   = 17
	IRQ_TS_AVAIL = 16
	IRQ_TS_TIMER = 15
)

type Stats struct {
	// The following MAC receive errors are tracked only if
	// ENET.DiscardErrors is false at its initialization.
	FrameLengthViolation uint32
	NonOctetAlignedFrame uint32
	CRCOrFrameError      uint32
	Overrun              uint32
	FrameTooSmall        uint32
	FrameTooLarge        uint32
}

// ENET represents an Ethernet MAC instance.
type ENET struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Clock retrieval function
	Clock func() uint32
	// Interrupt ID
	IRQ int
	// PLL enable function
	EnablePLL func(index int) error
	// PHY enable function
	EnablePHY func(eth *ENET) error
	// RMII mode
	RMII bool
	// MAC address (use SetMAC() for post Init() changes)
	MAC net.HardwareAddr
	// Incoming packet handler
	RxHandler func([]byte)
	// Descriptor ring size
	RingSize int

	// Discard MAC layer errors
	DiscardErrors bool
	// Statistics about the MAC
	Stats Stats

	// control registers
	eir  uint32
	eimr uint32
	rdar uint32
	tdar uint32
	ecr  uint32
	mmfr uint32
	mscr uint32
	mib  uint32
	rcr  uint32
	tcr  uint32
	palr uint32
	paur uint32
	rdsr uint32
	tdsr uint32
	mrbr uint32
	ftrl uint32
	racc uint32

	// receive data buffers
	rx bufferDescriptorRing
	// transmit data buffers
	tx bufferDescriptorRing
}

// Init initializes and enables the Ethernet MAC controller for 100 Mbps full
// duplex operation with Auto-Negotiation enabled.
func (hw *ENET) Init() {
	hw.Lock()

	if hw.Base == 0 || hw.Clock == nil || hw.EnablePLL == nil || hw.EnablePHY == nil {
		panic("invalid ENET controller instance")
	}

	if hw.MAC == nil {
		hw.MAC = make([]byte, 6)
		rand.Read(hw.MAC)
		// flag address as unicast and locally administered
		hw.MAC[0] &= 0xfe
		hw.MAC[0] |= 0x02
	} else if len(hw.MAC) != 6 {
		panic("invalid ENET hardware address")
	}

	if hw.RingSize == 0 {
		hw.RingSize = defaultRingSize
	}

	hw.eir = hw.Base + ENETx_EIR
	hw.eimr = hw.Base + ENETx_EIMR
	hw.rdar = hw.Base + ENETx_RDAR
	hw.tdar = hw.Base + ENETx_TDAR
	hw.ecr = hw.Base + ENETx_ECR
	hw.mmfr = hw.Base + ENETx_MMFR
	hw.mscr = hw.Base + ENETx_MSCR
	hw.mib = hw.Base + ENETx_MIB
	hw.rcr = hw.Base + ENETx_RCR
	hw.tcr = hw.Base + ENETx_TCR
	hw.palr = hw.Base + ENETx_PALR
	hw.paur = hw.Base + ENETx_PAUR
	hw.rdsr = hw.Base + ENETx_RDSR
	hw.tdsr = hw.Base + ENETx_TDSR
	hw.mrbr = hw.Base + ENETx_MRBR
	hw.ftrl = hw.Base + ENETx_FTRL
	hw.racc = hw.Base + ENETx_RACC

	hw.setup()

	hw.Unlock()
}

func (hw *ENET) setup() {
	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
	hw.EnablePLL(hw.Index)

	// soft reset
	reg.Set(hw.ecr, ECR_RESET)
	reg.Wait(hw.ecr, ECR_RESET, 1, 0)
	reg.Set(hw.ecr, ECR_DBSWP)

	// clear all interrupts
	reg.Write(hw.eir, 0xffffffff)
	// mask all interrupts
	reg.Write(hw.eimr, 0)

	// enable Full-Duplex
	reg.Set(hw.tcr, TCR_FDEN)
	// disable Management Information Database
	reg.Set(hw.mib, MIB_DIS)

	// use legacy descriptors
	reg.Clear(hw.ecr, ECR_EN1588)

	// set receive buffer size and maximum frame length
	size := MTU + (bufferAlign - (MTU % bufferAlign))
	reg.Write(hw.mrbr, uint32(size))
	reg.Write(hw.ftrl, MTU)
	reg.SetN(hw.rcr, RCR_MAX_FL, 0x3fff, MTU)

	if hw.DiscardErrors {
		reg.Set(hw.racc, RACC_LINEDIS)
	}

	// set physical address
	hw.SetMAC(hw.MAC)

	// set Media Independent Interface Mode
	reg.Set(hw.rcr, RCR_MII_MODE)
	reg.SetTo(hw.rcr, RCR_RMII_MODE, hw.RMII)
	// enable Flow Control
	reg.Set(hw.rcr, RCR_FCE)
	// disable loopback
	reg.Clear(hw.rcr, RCR_LOOP)

	// set MII clock
	reg.SetN(hw.mscr, MSCR_HOLDTIME, 0b111, 1)
	reg.SetN(hw.mscr, MSCR_MII_SPEED, 0x3f, hw.Clock()/(2*2500000))

	// enable Ethernet MAC
	reg.Set(hw.ecr, ECR_ETHEREN)

	// enable Ethernet PHY
	hw.EnablePHY(hw)
}

// SetMAC allows to change the controller physical address register after
// initialization.
func (hw *ENET) SetMAC(mac net.HardwareAddr) {
	hw.MAC = mac

	lower := binary.BigEndian.Uint32(hw.MAC[0:4])
	upper := binary.BigEndian.Uint16(hw.MAC[4:6])

	reg.Write(hw.palr, lower)
	reg.Write(hw.paur, uint32(upper)<<16)
}

// Start begins processing of incoming packets. When the argument is true the
// function waits and handles received packets (see Rx()) through RxHandler()
// (when set), it should never return.
func (hw *ENET) Start(rx bool) {
	// set receive and transmit descriptors
	reg.Write(hw.rdsr, hw.rx.init(true, hw.RingSize, &hw.Stats))
	reg.Write(hw.tdsr, hw.tx.init(false, hw.RingSize, &hw.Stats))

	reg.Set(hw.rdar, RDAR_ACTIVE)

	if !rx || hw.RxHandler == nil {
		return
	}

	var buf []byte

	for {
		runtime.Gosched()

		if buf = hw.Rx(); buf != nil {
			hw.RxHandler(buf)
		}
	}
}

// EnableInterrupt enables interrupt generation for a specific event.
func (hw *ENET) EnableInterrupt(event int) {
	reg.Set(hw.eimr, event)
}

// ClearInterrupt clears the interrupt corresponding to a specific event.
func (hw *ENET) ClearInterrupt(event int) {
	reg.Set(hw.eir, event)
}
