// VirtIO driver support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package virtio implements a driver for Virtual I/O devices (VirtIO)
// following reference specifications:
//   - Virtual I/O Device (VIRTIO) - Version 1.2
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package virtio

import (
	"errors"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// VirtIO MMIO Device Registers
const (
	Magic             = 0x000
	Version           = 0x004
	DeviceID          = 0x008
	VendorID          = 0x00c
	DeviceFeatures    = 0x010
	DeviceFeaturesSel = 0x014
	DriverFeatures    = 0x020
	DriverFeaturesSel = 0x024
	QueueSel          = 0x030
	QueueNumMax       = 0x034
	QueueNum          = 0x038
	QueueReady        = 0x044
	QueueNotify       = 0x050
	InterruptStatus   = 0x060
	InterruptACK      = 0x064
	Status            = 0x070
	QueueDesc         = 0x080
	QueueDriver       = 0x090
	QueueDevice       = 0x0a0
	ConfigGeneration  = 0x0fc
	Config            = 0x100
)

// Reserved Feature bits
const (
	Packed           = 34
	NotificationData = 38
)

// Device Status bits
const (
	Acknowledge      = 0
	Driver           = 1
	DriverOk         = 2
	FeaturesOk       = 3
	DeviceneedsReset = 6
	Failed           = 7
)

const (
	MAGIC   = 0x74726976 // "virt"
	VERSION = 0x02

	// bits 0 to 23, and 50 to 63
	deviceSpecificFeatureMask = 0xfffc000000ffffff
	// bits 24 to 49
	deviceReservedFeatureMask = 0x0003ffffff000000
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

	features uint64
}

// Init initializes a VirtIO over MMIO device instance.
func (io *VirtIO) Init(driverFeatures uint64) (err error) {
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

	// get offered features
	io.features = io.DeviceFeatures()

	// clear unsupported features
	bits.Clear64(&io.features, Packed)
	bits.Clear64(&io.features, NotificationData)

	// keep all remaining reserved features, clear device type ones
	io.features &= deviceReservedFeatureMask

	// apply device type features from the driver
	io.features &= driverFeatures

	// negotiate features
	io.SetDriverFeatures(io.features)
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

// SetQueue registers the indexed virtual queue for device access.
func (io *VirtIO) SetQueue(index int, queue *VirtualQueue) {
	desc, driver, device := queue.Address()

	reg.Write(io.Base+QueueSel, uint32(index))
	reg.Write(io.Base+QueueDesc, uint32(desc))
	reg.Write(io.Base+QueueDriver, uint32(driver))
	reg.Write(io.Base+QueueDevice, uint32(device))
	reg.Write(io.Base+QueueReady, 1)
}

// QueueNotify notifies the device that a queue can be processed.
func (io *VirtIO) QueueNotify(index int) {
	reg.Write(io.Base+QueueNotify, uint32(index))
}

// ConfigVersion returns the device configuration (see Config field) version.
func (io *VirtIO) ConfigVersion() uint32 {
	return reg.Read(io.Base + ConfigGeneration)
}
