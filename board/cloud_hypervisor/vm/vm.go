// Cloud Hypervisor support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package vm provides hardware initialization, automatically on import, for a
// Cloud Hypervisor virtual machine configured with a single x86_64 core.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package vm

import (
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/kvm/pvclock"
	"github.com/usbarmory/tamago/soc/intel/apic"
	"github.com/usbarmory/tamago/soc/intel/pci"
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

	// VirtIO Memory-mapped I/O
	VIRTIO_MMIO_BASE = 0xe8000000

	// VirtIO Networking
	VIRTIO_NET_PCI_VENDOR = 0x1af4 // Red Hat, Inc.
	VIRTIO_NET_PCI_DEVICE = 0x1041 // Virtio 1.0 network device
)

// Peripheral instances
var (
	// AMD64 core
	AMD64 = &amd64.CPU{}

	// Local APIC
	LAPIC = &apic.LAPIC{
		Base: LAPIC_BASE,
	}

	// I/O APIC - GSI 0-23
	IOAPIC0 = &apic.IOAPIC{
		Base: IOAPIC0_BASE,
	}

	// Serial port
	UART0 = &uart.UART{
		Index: 1,
		Base:  COM1,
	}
)

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(float64(AMD64.TimerFn())*AMD64.TimerMultiplier) + AMD64.TimerOffset
}

// Init takes care of the lower level initialization triggered early in runtime
// setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize CPU
	AMD64.Init()

	// initialize I/O APIC
	IOAPIC0.Init()
	// initialize serial console
	UART0.Init()

	runtime.Exit = func(_ int32) {
		// shutdown_pio_address
		reg.Out32(0x600, 0x34)
	}
}

func init() {
	// trap CPU exceptions
	AMD64.EnableExceptions()

	// allocate global DMA region
	dma.Init(dmaStart, dmaSize)

	// initialize KVM pvclock as needed
	pvclock.Init(AMD64)

	if dev := pci.Probe(0, VIRTIO_NET_PCI_VENDOR, VIRTIO_NET_PCI_DEVICE); dev != nil {
		// reconfigure BAR to mapped memory region
		dev.Write(0, pci.Bar0, 0xc0000000)
		dev.Write(0, pci.Bar0+4, 0)
	}
}
