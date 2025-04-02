// VirtIO driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
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
	"github.com/usbarmory/tamago/bits"
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
	DeviceNeedsReset = 6
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
type VirtIO interface {
	// Init initializes a VirtIO device instance.
	Init(features uint64) (err error)
	// Config returns the device configuration layout.
	Config(size int) []byte
	// DeviceID returns the VirtIO subsystem device ID
	DeviceID() uint32
	// DeviceFeatures returns the device feature bits.
	DeviceFeatures() (features uint64)
	// DriverFeatures returns the driver feature bits.
	DriverFeatures() (features uint64)
	// SetDriverFeatures sets the driver feature bits.
	SetDriverFeatures(features uint64)
	// NegotiatedFeatures returns the set of negotiated feature bits.
	NegotiatedFeatures() (features uint64)
	// QueueReady returns whether a queue is ready for use.
	QueueReady(index int) (ready bool)
	// MaxQueueSize returns the maximum virtual queue size.
	MaxQueueSize(index int) int
	// SetQueueSize sets the virtual queue size.
	SetQueueSize(index int, n int)
	// InterruptStatus returns the interrupt status and reason.
	InterruptStatus() (buffer bool, config bool)
	// Status returns the device status.
	Status() uint32
	// SetQueue registers the indexed virtual queue for device access.
	SetQueue(index int, queue *VirtualQueue)
	// SetReady indicates that the driver is set up and ready to drive the device.
	SetReady()
	// QueueNotify notifies the device that a queue can be processed.
	QueueNotify(index int)
	// ConfigVersion returns the device configuration (see Config field) version.
	ConfigVersion() uint32
}

func negotiate(deviceFeatures, driverFeatures uint64) (features uint64) {
	features = deviceFeatures

	// clear unsupported features
	bits.Clear64(&features, Packed)
	bits.Clear64(&features, NotificationData)

	// keep all remaining reserved features, clear device type ones
	features &= deviceReservedFeatureMask

	// apply device type features from the driver
	features &= driverFeatures

	return
}
