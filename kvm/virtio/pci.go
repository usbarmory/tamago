// VirtIO over PCI driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package virtio

import (
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/intel/apic"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

// VirtIO Common Configuration offsets
const (
	deviceFeatureSel = 0x00
	deviceFeature    = 0x04
	driverFeatureSel = 0x08
	driverFeature    = 0x0c
	configMSIXVector = 0x10
	numQueues        = 0x12
	deviceStatus     = 0x14
	configGeneration = 0x15
	queueSel         = 0x16
	queueSize        = 0x18
	queueMSIXVector  = 0x1a
	queueEnable      = 0x1c
	queueNotifyOff   = 0x1e
	queueDesc        = 0x20
	queueDriver      = 0x28
	queueDevice      = 0x30
)

const capabilityLength = 16

// VirtIO PCI Capabilities Configuration Types
const (
	capCommon = 1
	capNotify = 2
	capISR    = 3
	capDevice = 4
	capPCI    = 5
	capShmem  = 8
	capVendor = 9
)

// VirtIO PCI Capability
type capability struct {
	pci.CapabilityHeader

	CapLength uint8
	CfgType   uint8
	Bar       uint8
	ID        uint8
	_         uint16
	Offset    uint32
	Length    uint32
}

func (c *capability) Unmarshal(d *pci.Device, off uint32) (desc []byte, addr uint, err error) {
	buf := make([]byte, capabilityLength)

	for i := uint32(0); i < capabilityLength; i += 4 {
		binary.LittleEndian.PutUint32(buf[i:i+4], d.Read(0, off+i))
	}

	if _, err = binary.Decode(buf, binary.LittleEndian, c); err != nil {
		return nil, 0, errors.New("invalid capability format")
	}

	addr = d.BaseAddress(int(c.Bar)) + uint(c.Offset)
	size := int(c.Length)

	if addr == 0 || size == 0 {
		return
	}

	r, err := dma.NewRegion(addr, size, false)

	if err != nil {
		return nil, 0, errors.New("invalid capability region")
	}

	_, desc = r.Reserve(size, 0)

	return
}

// PCI represents a VirtIO over PCI device.
type PCI struct {
	// Device represents the probed PCI device.
	Device *pci.Device

	features uint64

	// notification structure layout
	queueNotifyOff   uint16
	notifyAddress    uint64
	notifyMultiplier uint32

	msix *pci.CapabilityMSIX

	// DMA buffers
	common []byte
	config []byte
}

func (io *PCI) init() error {
	off := io.Device.Read(0, pci.CapabilitiesOffset)

	// TODO: move to pci pkg as walk function
	for off != 0 {
		hdr := &pci.CapabilityHeader{}

		if err := hdr.Unmarshal(io.Device, off); err != nil {
			return errors.New("invalid capability header")
		}

		switch hdr.Vendor {
		case pci.VendorSpecific:
			c := &capability{}

			buf, addr, err := c.Unmarshal(io.Device, off)

			if err != nil {
				return err
			}

			switch c.CfgType {
			case capCommon:
				io.common = buf
			case capNotify:
				io.notifyAddress = uint64(addr)
				io.notifyMultiplier = io.Device.Read(0, off+capabilityLength)
			case capDevice:
				io.config = buf
			}
		case pci.MSIX:
			c := &pci.CapabilityMSIX{}

			if err := c.Unmarshal(io.Device, off); err != nil {
				return err
			}

			io.msix = c
		}

		off = uint32(hdr.Next)
	}

	if io.common == nil || io.config == nil {
		return errors.New("missing required capabilities")
	}

	return nil
}

func (io *PCI) negotiate(driverFeatures uint64) (err error) {
	io.features = negotiate(io.DeviceFeatures(), driverFeatures)
	io.SetDriverFeatures(io.features)

	io.common[deviceStatus] |= (1 << FeaturesOk)

	if io.common[deviceStatus]&(1<<FeaturesOk) != (1 << FeaturesOk) {
		return errors.New("could not set features")
	}

	return
}

// Init initializes a VirtIO over PCI device instance.
func (io *PCI) Init(features uint64) (err error) {
	if io.Device == nil {
		return errors.New("invalid VirtIO instance")
	}

	if rev := io.Device.Read(0, pci.RevisionID) & 0xff; rev == 0 {
		return errors.New("transitional devices are not supported")
	}

	if err = io.init(); err != nil {
		return
	}

	// reset
	io.common[deviceStatus] = 0

	// initialize driver
	io.common[deviceStatus] |= (1 << Acknowledge)
	io.common[deviceStatus] |= (1 << Driver)

	return io.negotiate(features)
}

// Config returns the device configuration layout.
func (io *PCI) Config(size int) (config []byte) {
	config = make([]byte, size)
	copy(config, io.config)
	return
}

// DeviceID returns the VirtIO subsystem device ID
func (io *PCI) DeviceID() uint32 {
	// The PCI Device ID is calculated by adding 0x1040 to the Virtio
	// Device ID (4.1.2 PCI Device Discovery.)
	return uint32(io.Device.Device - 0x1040)
}

// DeviceFeatures returns the device feature bits.
func (io *PCI) DeviceFeatures() (features uint64) {
	for i := uint32(0); i <= 1; i++ {
		binary.LittleEndian.PutUint32(io.common[deviceFeatureSel:], i)
		features |= uint64(binary.LittleEndian.Uint32(io.common[deviceFeature:])) << (i * 32)
	}

	return
}

// DriverFeatures returns the driver feature bits.
func (io *PCI) DriverFeatures() (features uint64) {
	for i := uint32(0); i <= 1; i++ {
		binary.LittleEndian.PutUint32(io.common[driverFeatureSel:], i)
		features |= uint64(binary.LittleEndian.Uint32(io.common[driverFeature:])) << (i * 32)
	}

	return
}

// SetDriverFeatures sets the driver feature bits.
func (io *PCI) SetDriverFeatures(features uint64) {
	for i := uint32(0); i <= 1; i++ {
		binary.LittleEndian.PutUint32(io.common[driverFeatureSel:], i)
		binary.LittleEndian.PutUint32(io.common[driverFeature:], uint32(features>>(i*32)))
	}

	return
}

// NegotiatedFeatures returns the set of negotiated feature bits.
func (io *PCI) NegotiatedFeatures() (features uint64) {
	return io.features
}

// QueueReady returns whether a queue is ready for use.
func (io *PCI) QueueReady(index int) (ready bool) {
	binary.LittleEndian.PutUint16(io.common[queueSel:], uint16(index))
	ready = binary.LittleEndian.Uint16(io.common[queueEnable:]) != 0
	return
}

// MaxQueueSize returns the maximum virtual queue size.
func (io *PCI) MaxQueueSize(index int) int {
	binary.LittleEndian.PutUint16(io.common[queueSel:], uint16(index))
	return int(binary.LittleEndian.Uint16(io.common[queueSize:]))
}

// SetQueueSize sets the virtual queue size.
func (io *PCI) SetQueueSize(index int, n int) {
	binary.LittleEndian.PutUint16(io.common[queueSel:], uint16(index))
	binary.LittleEndian.PutUint16(io.common[queueSize:], uint16(n))
}

// EnableInterrupt enables MSI-X interrupt vector routing to a LAPIC instance
// for the indexed virtual queue.
func (io *PCI) EnableInterrupt(id int, lapic *apic.LAPIC, index int) {
	if io.msix == nil {
		return
	}

	entry := 0
	addr := uint64(lapic.Base)
	data := uint32(id)

	io.msix.EnableInterrupt(entry, addr, data)

	binary.LittleEndian.PutUint16(io.common[queueSel:], uint16(index))
	binary.LittleEndian.PutUint16(io.common[queueMSIXVector:], uint16(entry))
}

// Status returns the device status.
func (io *PCI) Status() uint32 {
	return uint32(io.common[deviceStatus])
}

// SetQueue registers the indexed virtual queue for device access.
func (io *PCI) SetQueue(index int, queue *VirtualQueue) {
	desc, driver, device := queue.Address()

	binary.LittleEndian.PutUint16(io.common[queueSel:], uint16(index))
	binary.LittleEndian.PutUint64(io.common[queueDesc:], uint64(desc))
	binary.LittleEndian.PutUint64(io.common[queueDriver:], uint64(driver))
	binary.LittleEndian.PutUint64(io.common[queueDevice:], uint64(device))
	binary.LittleEndian.PutUint16(io.common[queueEnable:], 1)
}

// SetReady indicates that the driver is set up and ready to drive the device.
func (io *PCI) SetReady() {
	io.queueNotifyOff = binary.LittleEndian.Uint16(io.common[queueNotifyOff:])
	io.common[deviceStatus] |= (1 << DriverOk)
}

// QueueNotify notifies the device that a queue can be processed.
func (io *PCI) QueueNotify(index int) {
	addr := io.notifyAddress
	addr += uint64(index) * uint64(io.queueNotifyOff) * uint64(io.notifyMultiplier)

	reg.Write64(addr, uint64(index))
}

// ConfigVersion returns the device configuration (see Config field) version.
func (io *PCI) ConfigVersion() uint32 {
	return uint32(io.common[configGeneration])
}
