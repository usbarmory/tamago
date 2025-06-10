// QEMU microvm support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package microvm provides hardware initialization, automatically on import,
// for a QEMU microvm machine configured with a single x86_64 core.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package microvm

import (
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/amd64/lapic"
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
	LAPIC_BASE   = 0xfee00000
	IOAPIC0_BASE = 0xfec00000
	IOAPIC1_BASE = 0xfec10000

	// VirtIO Memory-mapped I/O
	VIRTIO_MMIO_BASE = 0xfeb00000

	// VirtIO Networking
	VIRTIO_NET0_BASE = VIRTIO_MMIO_BASE + 0x2e00
	VIRTIO_NET0_IRQ  = 47
)

// Peripheral instances
var (
	// CPU instance(s)
	AMD64 = &amd64.CPU{
		// required before Init()
		TimerMultiplier: 1,
		// Local APIC
		LAPIC: &lapic.LAPIC{
			Base: LAPIC_BASE,
		},
	}

	// I/O APIC - GSI 0-23
	IOAPIC0 = &ioapic.IOAPIC{
		Index:   0,
		Base:    IOAPIC0_BASE,
		GSIBase: 0,
	}

	// I/O APIC - GSI 24-47
	IOAPIC1 = &ioapic.IOAPIC{
		Index:   1,
		Base:    IOAPIC1_BASE,
		GSIBase: 24,
	}

	// Real-Time Clock
	RTC = &rtc.RTC{}

	// Serial port
	UART0 = &uart.UART{
		Index: 1,
		Base:  COM1,
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

	// initialize I/O APICs
	IOAPIC0.Init()
	IOAPIC1.Init()

	// initialize serial console
	UART0.Init()

	runtime.Exit = func(_ int32) {
		// On microvm the recommended way to trigger a guest-initiated
		// shut down is by generating a triple-fault.
		amd64.Fault()
	}
}

func init() {
	// trap CPU exceptions
	AMD64.EnableExceptions()

	// initialize AP
	AMD64.InitSMP(2)

	// allocate global DMA region
	dma.Init(dmaStart, dmaSize)

	// initialize KVM pvclock as needed
	pvclock.Init(AMD64)
}
