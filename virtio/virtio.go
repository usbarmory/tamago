// VirtIO driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// https://wiki.osdev.org/Virtio
package virtio

import (
	"errors"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	MAGIC = 0x74726976 // "virt"
)

// VirtIO represents a VirtIO device.
type VirtIO struct {
	// Base register
	Base uint32
	// Virtual Queue
	Queue *VirtualQueue
}

// Init initializies a VirtIO instance.
func (io *VirtIO) Init() (err error) {
	if io.Base == 0 || reg.Read(io.Base+Magic) != MAGIC {
		return errors.New("invalid VirtIO instance")
	}

	if reg.Read(io.Base+Version) != 0x02 {
		return errors.New("unsupported VirtIO interface")
	}

	io.Queue = &VirtualQueue{}

	return
}

// DeviceID returns the VirtIO subsystem device ID
func (io *VirtIO) DeviceID() uint32 {
	return reg.Read(io.Base + DeviceID)
}

// DeviceID returns the device feature bits.
func (io *VirtIO) DeviceFeatures() uint32 {
	return reg.Read(io.Base + DeviceFeatures)
}

// SelectQueue selects the virtual queue index.
func (io *VirtIO) SelectQueue(index uint32) {
	reg.Write(io.Base+QueueSel, index)
}

// MaxQueueSize returns the maximum virtual queue size for the indexed queue.
func (io *VirtIO) MaxQueueSize() uint32 {
	return reg.Read(io.Base + QueueNumMax)
}

// SetQueueSize sets the virtual queue size for the indexed queue.
func (io *VirtIO) SetQueueSize(n uint32) {
	reg.Write(io.Base+QueueNum, n)
}
