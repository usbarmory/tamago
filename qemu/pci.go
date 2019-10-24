// Basic PCI driver to support QEMU VirtIORNG
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,amd64

package qemu

import (
	"fmt"
	"runtime"
)

const CONFIG_ADDRESS = 0x0cf8
const CONFIG_DATA = 0x0cfc

var PCICache []*PCICacheEntry

type PCICacheEntry struct {
	bus    uint32
	vendor uint16
	device uint16
	bar0   uint32
	found  bool
}

func PCIRead(bus uint32, slot uint32, fn uint32, offset uint32) uint16 {
	address := (bus << 16) | (slot << 11) | (fn << 8) | (offset & 0xfc) | 0x80000000
	runtime.Outl(CONFIG_ADDRESS, int(address))

	return (uint16)((runtime.Inl(CONFIG_DATA) >> ((offset & 2) * 8)) & 0xffff)
}

func PCIProbe(bus uint32, vendor uint16, device uint16) (bar0 uint32, found bool) {
	for _, entry := range PCICache {
		if entry.bus == bus && entry.vendor == vendor && entry.device == device {
			return entry.bar0, entry.found
		}
	}

	for slot := uint32(0); slot <= 31; slot++ {
		probed_vendor := PCIRead(bus, slot, 0, 0)

		if probed_vendor == 0xffff {
			continue
		}

		probed_device := PCIRead(0, slot, 0, 0x02)

		if vendor == probed_vendor && device == probed_device {
			bar0_0 := PCIRead(bus, slot, 0, 0x10)
			bar0_1 := PCIRead(bus, slot, 0, 0x12)
			bar0 = (uint32(bar0_1) << 16) + uint32(bar0_0)
			bar0 = bar0 & 0xFFFFFFFC

			found = true

			if bootstrapped {
				fmt.Printf("PCI: found device %d:%d at slot %d, BAR0 %#x\n", vendor, device, slot, bar0)
			}
		}
	}

	entry := &PCICacheEntry{
		bus:    bus,
		vendor: vendor,
		device: device,
		bar0:   bar0,
		found:  found,
	}

	PCICache = append(PCICache, entry)

	return
}
