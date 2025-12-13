// VirtIO over PCI driver (legacy)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package virtio

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

const (
	deviceMin = 0x1000
	deviceMax = 0x103f
)

const (
	pageSize            = 4096
	configurationLength = 20
)

// VirtIO Common Configuration offsets (legacy interface)
const (
	legacyDeviceFeatures   = 0x00
	legacyDriverFeatures   = 0x04
	legacyQueueAddress     = 0x08
	legacyQueueSize        = 0x0c
	legacyQueueSelect      = 0x0e
	legacyQueueNotify      = 0x10
	legacyDeviceStatus     = 0x12
	legacyISRStatus        = 0x13
	legacyDeviceConfig     = 0x14
	legacyConfigMSIXVector = 0x14
	legacyQueueMSIXVector  = 0x16
	legacyDeviceConfigMSI  = 0x18
)

// LegacyPCI represents a legacy VirtIO over PCI device.
type LegacyPCI struct {
	// Device represents the probed PCI device.
	Device *pci.Device

	// I/O port space
	config   uint16
	features uint64

	msix        *pci.CapabilityMSIX
	msixEnabled bool
}

func (io *LegacyPCI) addCapability(off uint32, hdr *pci.CapabilityHeader) error {
	switch hdr.Vendor {
	case pci.VendorSpecific:
		return fmt.Errorf("unexpected PCI capability %x", hdr.Vendor)
	case pci.MSIX:
		c := &pci.CapabilityMSIX{}

		if err := c.Unmarshal(io.Device, off); err != nil {
			return err
		}

		io.msix = c
	}

	return nil
}

func (io *LegacyPCI) negotiate(driverFeatures uint64) {
	io.features = negotiate(io.DeviceFeatures(), driverFeatures)
	io.SetDriverFeatures(io.features)
}

// Init initializes a legacy VirtIO over PCI device instance.
func (io *LegacyPCI) Init(features uint64) (err error) {
	if io.Device == nil {
		return errors.New("invalid VirtIO instance")
	}

	if rev := io.Device.Read(0, pci.RevisionID); rev&0xff != 0 {
		return errors.New("not a transitional device")
	}

	if io.Device.Device < deviceMin || io.Device.Device > deviceMax {
		return errors.New("not a transitional device")
	}

	bar0 := io.Device.BaseAddress(0)

	if bar0&1 != 1 {
		return errors.New("unexpected PCI BAR type, expected I/O port")
	}

	for off, hdr := range io.Device.Capabilities() {
		if err = io.addCapability(off, hdr); err != nil {
			return
		}
	}

	io.config = uint16(bar0) & 0xfff0

	// reset
	io.setStatus(0)

	// initialize driver
	s := io.Status()
	s |= (1 << Acknowledge)
	s |= (1 << Driver)
	io.setStatus(s)

	io.negotiate(features)

	return
}

// Config returns the device configuration.
func (io *LegacyPCI) Config(size int) (config []byte) {
	var off int

	if io.msixEnabled {
		off = legacyDeviceConfigMSI
	} else {
		off = legacyDeviceConfig
	}

	config = make([]byte, size)

	for i := 0; i < size; i += 1 {
		config[i] = reg.In8(io.config + uint16(off+i))
	}

	return
}

// DeviceID returns the VirtIO subsystem device ID
func (io *LegacyPCI) DeviceID() uint32 {
	return uint32(io.Device.Device - 0x1000 + 1)
}

// DeviceFeatures returns the device feature bits,
func (io *LegacyPCI) DeviceFeatures() (features uint64) {
	return uint64(reg.In32(io.config + legacyDeviceFeatures))
}

// DriverFeatures returns the driver feature bits.
func (io *LegacyPCI) DriverFeatures() (features uint64) {
	return uint64(reg.In32(io.config + legacyDriverFeatures))
}

// SetDriverFeatures sets the driver feature bits (note that only the first 32
// feature bits are accessible through the legacy interface).
func (io *LegacyPCI) SetDriverFeatures(features uint64) {
	reg.Out32(io.config+legacyDriverFeatures, uint32(features))
}

// NegotiatedFeatures returns the set of negotiated feature bits.
func (io *LegacyPCI) NegotiatedFeatures() (features uint64) {
	return io.features
}

// QueueReady returns whether a queue is ready for use.
func (io *LegacyPCI) QueueReady(index int) (ready bool) {
	reg.Out16(io.config+legacyQueueSelect, uint16(index))
	return reg.In32(io.config+legacyQueueAddress) != 0
}

// MaxQueueSize returns the maximum virtual queue size.
func (io *LegacyPCI) MaxQueueSize(index int) int {
	reg.Out16(io.config+legacyQueueSelect, uint16(index))
	return int(reg.In16(io.config + legacyQueueSize))
}

// SetQueueSize is unsupported on legacy devuces.
func (io *LegacyPCI) SetQueueSize(_ int, _ int) {}

// Status returns the device status.
func (io *LegacyPCI) Status() uint32 {
	return uint32(reg.In16(io.config + legacyDeviceStatus))
}

func (io *LegacyPCI) setStatus(s uint32) {
	reg.Out8(io.config+legacyDeviceStatus, uint8(s))
}

// SetQueue registers the indexed virtual queue for device access.
func (io *LegacyPCI) SetQueue(index int, queue *VirtualQueue) {
	desc, _, _ := queue.Address()
	reg.Out16(io.config+legacyQueueSelect, uint16(index))
	reg.Out32(io.config+legacyQueueAddress, uint32(desc/pageSize))
}

// SetReady indicates that the driver is set up and ready to drive the device.
func (io *LegacyPCI) SetReady() {
	s := io.Status()
	s |= (1 << DriverOk)
	io.setStatus(s)
}

// QueueNotify notifies the device that a queue can be processed.
func (io *LegacyPCI) QueueNotify(index int) {
	reg.Out16(io.config+legacyQueueNotify, uint16(index))
}

// ConfigVersion always returns 0 as legacy devices do not support a generation
// count for the configuration space.
func (io *LegacyPCI) ConfigVersion() uint32 {
	return 0
}

// EnableInterrupt enables MSI-X interrupt vector routing to a LAPIC instance
// for the indexed virtual queue.
func (io *LegacyPCI) EnableInterrupt(id int, index int) (err error) {
	if io.msix == nil {
		return errors.New("missing required capabilities")
	}

	entry := 0
	addr := uint64(amd64.LAPIC_BASE)
	data := uint32(id)

	if err = io.msix.EnableInterrupt(entry, addr, data); err != nil {
		return
	}

	io.msixEnabled = true
	reg.Out16(io.config+legacyQueueSelect, uint16(index))
	reg.Out16(io.config+legacyQueueMSIXVector, uint16(entry))

	return
}
