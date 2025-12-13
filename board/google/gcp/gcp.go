// Google Compute Engine support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gcp provides hardware initialization, automatically on import, for a
// Google Compute Engine machine configured with one or more x86_64 cores.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package gcp

import (
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/kvm/pvclock"
	"github.com/usbarmory/tamago/soc/intel/ioapic"
	"github.com/usbarmory/tamago/soc/intel/rtc"
	"github.com/usbarmory/tamago/soc/intel/uart"
)

const (
	dmaStart = 0x50000000
	dmaSize  = 0x10000000 // 256MB
)

// Peripheral registers
const (
	// Communication port
	COM1 = 0x3f8

	// Intel I/O Programmable Interrupt Controllers
	IOAPIC0_BASE = 0xfec00000

	// VirtIO Networking
	VIRTIO_NET_PCI_VENDOR = 0x1af4 // Red Hat, Inc.
	VIRTIO_NET_PCI_DEVICE = 0x1000 // Virtio 1.0 network device
)

// Peripheral instances
var (
	// CPU instance(s)
	AMD64 = &amd64.CPU{
		// required before Init()
		TimerMultiplier: 1,
	}

	// I/O APIC - GSI 0-23
	IOAPIC0 = &ioapic.IOAPIC{
		Index:   0,
		Base:    IOAPIC0_BASE,
		GSIBase: 0,
	}

	// Real-Time Clock
	RTC = &rtc.RTC{}

	// Serial port
	UART0 = &uart.UART{
		Index: 1,
		Base:  COM1,
		DTR:   true,
		RTS:   true,
	}
)

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return AMD64.GetTime()
}

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime.hwinit1
func Init() {
	// initialize BSP
	AMD64.Init()

	// initialize I/O APIC
	IOAPIC0.Init()

	// initialize serial console
	UART0.Init()

	runtime.Exit = func(_ int32) {
		amd64.Fault()
	}
}

func init() {
	// trap CPU exceptions
	AMD64.EnableExceptions()

	// initialize APs
	AMD64.InitSMP(-1)

	// allocate global DMA region
	dma.Init(dmaStart, dmaSize)

	// initialize KVM pvclock as needed
	pvclock.Init(AMD64)
}
