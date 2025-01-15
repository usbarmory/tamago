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

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	MAGIC   = 0x74726976 // "virt"
	VERSION = 0x02
)

// VirtIO represents a VirtIO device.
type VirtIO struct {
	// MMIO base address
	Base uint32
	// ConfigSize is the device configuration size
	ConfigSize int
	// Config is a reserved DMA buffer for device configuration access and
	// modification.
	Config []byte
}

// Init initializies a VirtIO over MMIO device instance.
func (io *VirtIO) Init() (err error) {
	if io.Base == 0 || reg.Read(io.Base+Magic) != MAGIC {
		return errors.New("invalid VirtIO instance")
	}

	if reg.Read(io.Base+Version) != VERSION {
		return errors.New("unsupported VirtIO interface")
	}

	// reset
	reg.Write(io.Base+Status, 0x0)

	// initialize driver
	reg.Set(io.Base+Status, Driver|Acknowledge)

	// set all features except packed virtual queues
	features := io.DeviceFeatures()
	bits.Clear64(&features, Packed)

	io.SetDriverFeatures(features)
	reg.Set(io.Base+Status, FeaturesOk)

	if !reg.IsSet(io.Base+Status, FeaturesOk) {
		return errors.New("could not set features")
	}

	// finalize driver
	reg.Set(io.Base+Status, DriverOk)

	// initialize Config DMA buffers
	r, err := dma.NewRegion(uint(io.Base+Config), io.ConfigSize, false)

	if err != nil {
		return
	}

	_, io.Config = r.Reserve(io.ConfigSize, 0)

	return
}

// DeviceID returns the VirtIO subsystem device ID
func (io *VirtIO) DeviceID() uint32 {
	return reg.Read(io.Base + DeviceID)
}

// DeviceFeatures returns the device feature bits.
func (io *VirtIO) DeviceFeatures() (features uint64) {
	for i := uint32(0); i <= 1; i++ {
		reg.Write(io.Base+DeviceFeaturesSel, i)
		features |= uint64(reg.Read(io.Base+DeviceFeatures)) << (i * 32)
	}

	return
}

// DriverFeatures returns the driver feature bits.
func (io *VirtIO) DriverFeatures() (features uint64) {
	for i := uint32(0); i <= 1; i++ {
		reg.Write(io.Base+DriverFeaturesSel, i)
		features |= uint64(reg.Read(io.Base+DriverFeatures)) << (i * 32)
	}

	return
}

// SetDriverFeatures sets the driver feature bits.
func (io *VirtIO) SetDriverFeatures(features uint64) {
	for i := uint32(0); i <= 1; i++ {
		reg.Write(io.Base+DriverFeaturesSel, i)
		reg.Write(io.Base+DriverFeatures, uint32(features>>(i*32)))
	}

	return
}

// QueueReady returns whether a queue is ready for use.
func (io *VirtIO) QueueReady(index int) (ready bool) {
	reg.Write(io.Base+QueueSel, uint32(index))
	ready = reg.Read(io.Base+QueueReady) != 0
	return
}

// MaxQueueSize returns the maximum virtual queue size.
func (io *VirtIO) MaxQueueSize(index int) int {
	reg.Write(io.Base+QueueSel, uint32(index))
	return int(reg.Read(io.Base + QueueNumMax))
}

// SetQueueSize sets the virtual queue size.
func (io *VirtIO) SetQueueSize(index int, n int) {
	reg.Write(io.Base+QueueSel, uint32(index))
	reg.Write(io.Base+QueueNum, uint32(n))
}

// InterruptStatus returns the interrupt status and reason.
func (io *VirtIO) InterruptStatus() (buffer bool, config bool) {
	s := reg.Read(io.Base + InterruptStatus)

	buffer = bits.IsSet(&s, 0)
	config = bits.IsSet(&s, 1)

	return
}

// Status returns the device status.
func (io *VirtIO) Status() uint32 {
	return reg.Read(io.Base + Status)
}

// Notify notifies the device about the location of the indexed virtual queue.
func (io *VirtIO) Notify(index int, queue *VirtualQueue) {
	desc, driver, device := queue.Address()

	reg.Write(io.Base+QueueSel, uint32(index))
	reg.Write(io.Base+QueueDesc, uint32(desc))
	reg.Write(io.Base+QueueDriver, uint32(driver))
	reg.Write(io.Base+QueueDriver, uint32(device))
}

// ConfigVersion returns the device configuration (see Config field) version.
func (io *VirtIO) ConfigVersion() uint32 {
	return reg.Read(io.Base + ConfigGeneration)
}
