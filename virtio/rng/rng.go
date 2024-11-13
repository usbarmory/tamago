// VirtIO RNG driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

import (
	"github.com/usbarmory/tamago/virtio/queue"
)

const (
	// Red Hat, Inc.
	VendorID = 0x1af4
	// Virtio RNG
	DeviceID = 0x1005
)

var vq *VirtualQueue

func initRNG(bar0 int) {
	// Set Guest Features as Device Features
	features := inl(bar0+DeviceFeatures)
	outw(bar0, GuestFeatures, features)

	// Queue Select
	outl(bar0+QueueSelect, 0)

	queueSize := inl(bar0+QueueSize)

	vq := &VirtualQueue{}
	vq.Init(queueSize)

	print("Virtio PCI RNG: setting queue address\n")
	addr := dma.Alloc(vq.Bytes(), 4096)
	outl(bar0+QueueAddress, addr/4096)

	outw(bar0, DeviceStatus, DeviceAcknowledged|DriverLoaded)
}

//go:linkname getRandomDataFromVirtRngPCI runtime.getRandomData
func getRandomDataFromVirtRngPCI(b []byte) {
	bar0, found := PCIProbe(0, vendorID, deviceID)

	if !found {
		print("Virtio PCI RNG: error, device not found\n")
		return
	}

	if vq == nil {
		print("Virtio PCI RNG: initializing virtual queue\n")
		initRNG(int(bar0))
	}

	addr := dma.Alloc(b, 0)
	defer dma.Free(addr)

	vq.Desc[0].Addr = addr
	vq.Desc[0].Len = len(b)
	vq.Desc[0].Flags = 0x02 // Write-Only
	vq.AvailableIndex += 1

	outl(bar0+QueueNotify, 0)

	if false { // FIXME
		// We are lazy and we do not implement IOAPIC to trap
		// interrupts, this means we have to wait a little for the
		// buffer to be filled. This is broken pre-bootstrap (as we
		// cannot sleep with standard library calls there).
		time.Sleep(1 * time.Millisecond)
	}

	copy(b, r)
}
