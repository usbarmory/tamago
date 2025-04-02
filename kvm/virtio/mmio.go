// VirtIO over MMIO driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

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

// MMIO represents a VirtIO over MMIO device.
type MMIO struct {
	// Base address
	Base uint32

	features uint64

	// DMA buffer
	config []byte
}

func (io *MMIO) negotiate(driverFeatures uint64) (err error) {
	io.features = negotiate(io.DeviceFeatures(), driverFeatures)
	io.SetDriverFeatures(io.features)

	reg.Set(io.Base+Status, FeaturesOk)

	if !reg.IsSet(io.Base+Status, FeaturesOk) {
		return errors.New("could not set features")
	}

	return
}

// Init initializes a VirtIO over MMIO device instance.
func (io *MMIO) Init(features uint64) (err error) {
	if io.Base == 0 || reg.Read(io.Base+Magic) != MAGIC {
		return errors.New("invalid VirtIO instance")
	}

	if reg.Read(io.Base+Version) != VERSION {
		return errors.New("unsupported VirtIO interface")
	}

	// reset
	reg.Write(io.Base+Status, 0x0)

	// initialize driver
	reg.Set(io.Base+Status, Acknowledge)
	reg.Set(io.Base+Status, Driver)

	return io.negotiate(features)
}

// Config returns the device configuration layout.
func (io *MMIO) Config(size int) (config []byte) {
	if io.config == nil {
		r, err := dma.NewRegion(uint(io.Base+Config), size, false)

		if err != nil {
			return
		}

		_, io.config = r.Reserve(size, 0)
	}

	config = make([]byte, size)
	copy(config, io.config)

	return
}

// DeviceID returns the VirtIO subsystem device ID
func (io *MMIO) DeviceID() uint32 {
	return reg.Read(io.Base + DeviceID)
}

// DeviceFeatures returns the device feature bits.
func (io *MMIO) DeviceFeatures() (features uint64) {
	for i := uint32(0); i <= 1; i++ {
		reg.Write(io.Base+DeviceFeaturesSel, i)
		features |= uint64(reg.Read(io.Base+DeviceFeatures)) << (i * 32)
	}

	return
}

// DriverFeatures returns the driver feature bits.
func (io *MMIO) DriverFeatures() (features uint64) {
	for i := uint32(0); i <= 1; i++ {
		reg.Write(io.Base+DriverFeaturesSel, i)
		features |= uint64(reg.Read(io.Base+DriverFeatures)) << (i * 32)
	}

	return
}

// SetDriverFeatures sets the driver feature bits.
func (io *MMIO) SetDriverFeatures(features uint64) {
	for i := uint32(0); i <= 1; i++ {
		reg.Write(io.Base+DriverFeaturesSel, i)
		reg.Write(io.Base+DriverFeatures, uint32(features>>(i*32)))
	}

	return
}

// NegotiatedFeatures returns the set of negotiated feature bits.
func (io *MMIO) NegotiatedFeatures() (features uint64) {
	return io.features
}

// QueueReady returns whether a queue is ready for use.
func (io *MMIO) QueueReady(index int) (ready bool) {
	reg.Write(io.Base+QueueSel, uint32(index))
	ready = reg.Read(io.Base+QueueReady) != 0
	return
}

// MaxQueueSize returns the maximum virtual queue size.
func (io *MMIO) MaxQueueSize(index int) int {
	reg.Write(io.Base+QueueSel, uint32(index))
	return int(reg.Read(io.Base + QueueNumMax))
}

// SetQueueSize sets the virtual queue size.
func (io *MMIO) SetQueueSize(index int, n int) {
	reg.Write(io.Base+QueueSel, uint32(index))
	reg.Write(io.Base+QueueNum, uint32(n))
}

// InterruptStatus returns the interrupt status and reason.
func (io *MMIO) InterruptStatus() (buffer bool, config bool) {
	s := reg.Read(io.Base + InterruptStatus)

	buffer = bits.IsSet(&s, 0)
	config = bits.IsSet(&s, 1)

	return
}

// Status returns the device status.
func (io *MMIO) Status() uint32 {
	return reg.Read(io.Base + Status)
}

// SetQueue registers the indexed virtual queue for device access.
func (io *MMIO) SetQueue(index int, queue *VirtualQueue) {
	desc, driver, device := queue.Address()

	reg.Write(io.Base+QueueSel, uint32(index))
	reg.Write(io.Base+QueueDesc, uint32(desc))
	reg.Write(io.Base+QueueDriver, uint32(driver))
	reg.Write(io.Base+QueueDevice, uint32(device))
	reg.Write(io.Base+QueueReady, 1)
}

// SetReady indicates that the driver is set up and ready to drive the device.
func (io *MMIO) SetReady() {
	reg.Set(io.Base+Status, DriverOk)
}

// QueueNotify notifies the device that a queue can be processed.
func (io *MMIO) QueueNotify(index int) {
	reg.Write(io.Base+QueueNotify, uint32(index))
}

// ConfigVersion returns the device configuration (see Config field) version.
func (io *MMIO) ConfigVersion() uint32 {
	return reg.Read(io.Base + ConfigGeneration)
}
