// VirtIO Queue Descriptor support
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
	"fmt"
	"runtime"
	"time"

	"github.com/usbarmory/tamago/dma"
)

// VirtIO device types
const (
	NetworkCard     = 0x01
	BlockDevice     = 0x02
	Console         = 0x03
	EntropySource   = 0x04
	MemoryBalloning = 0x05
	IOMemory        = 0x06
	RPMSG           = 0x07
	SCSIHost        = 0x08
	P9Transport     = 0x09
	MAC80211WLAN    = 0x10
)

// VirtIO I/O Registers
const (
	DeviceFeatures = 0x00
	GuestFeatures  = 0x04
	QueueAddress   = 0x08
	QueueSize      = 0x0c
	QueueSelect    = 0x0e
	QueueNotify    = 0x10
	DeviceStatus   = 0x12
	ISRStatus      = 0x13
)

// VirtIO Device Status
const (
	DeviceAcknowledged = 0x01
	DriverLoaded       = 0x02
	DriverReady        = 0x03
	DeviceError        = 0x40
	DriverFailed       = 0x80
)

type Buffer struct {
	Address uint64
	Length  uint32
	Flags   uint16
	Next    uint16
}

type Available struct {
	Flags      uint16
	Index      uint16
	Ring       [8]uint16
	EventIndex uint16
}

type Ring struct {
	Index  uint32
	Length uint32
}

type Used struct {
	Flags  uint16
	Index  uint16
	_ [2]byte

	Ring []Ring
	AvailEvent uint16
	_ [2]byte
}

type VirtualQueue struct {
	Buffers   []Buffer
	Available Available
	Used      Used
}
