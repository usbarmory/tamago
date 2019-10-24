// QEMU VirtIORNG driver
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
	"time"
	"unsafe"
)

// https://wiki.osdev.org/Virtio

type VirtualQueue struct {
	Desc            [8]VirtualQueueDesc
	Avail_flags     uint16
	Avail_idx       uint16
	Available       [8]uint16
	Used_event_idx  uint16
	Pad             [3008]int8
	Used_flags      uint16
	Used_idx        uint16
	Pad_cgo_0       [2]byte
	Used            [8]VirtualQueueUsedElem
	Avail_event_idx uint16
	Pad_cgo_1       [2]byte
}

type VirtualQueueDesc struct {
	Addr  uint64
	Len   uint32
	Flags uint16
	Next  uint16
}

type VirtualQueueUsedElem struct {
	Id  uint32
	Len uint32
}

type VirtualQueueCache struct {
	vq_buf  []byte
	vq_addr uintptr
	vq      *VirtualQueue
}

var vqc VirtualQueueCache

//go:linkname getRandomDataFromVirtRngPCI runtime.getRandomData
func getRandomDataFromVirtRngPCI(b []byte) {
	// found by debugging src/runtime/os_linux.go:alloc_map
	// pmsg("va:"); _pnum(va); pmsg("pa:"); _pnum(p_pg); pmsg("diff:"); _pnum(va-p_pg); pmsg("\n")
	// warning this offset changes depending on imported modules
	va_pa_off := uintptr(0x000000bfff993000)

	// To allow global destination pointers (used pre-bootstrap) it is best
	// to allocate our own byte array to ensure that the pointer arithmetic
	// is always consistent.
	r := make([]byte, len(b))
	r_p := unsafe.Pointer(&r[0])
	r_addr := uintptr(r_p)

	bar0, found := PCIProbe(0, 0x1af4, 0x1005)

	if !found {
		if bootstrapped {
			fmt.Println("Virtio PCI RNG: error, device not found")
		} else {
			runtime.Pmsg("Virtio PCI RNG: error, device not found")
		}

		return
	}

	if vqc.vq_addr == 0 {
		if bootstrapped {
			fmt.Println("Virtio PCI RNG: initializing virtual queue")
		} else {
			runtime.Pmsg("Virtio PCI RNG: initializing virtual queue")
		}

		// vq_addr must be 4096 byte aligned
		buf, addr, err := alignedBuffer(unsafe.Sizeof(VirtualQueue{}), 4096)

		if err != nil {
			runtime.Pmsg("Virtio PCI RNG: alignement error")
			return
		}

		vqc.vq_buf = buf
		vqc.vq_addr = addr
		vqc.vq = (*VirtualQueue)(unsafe.Pointer(vqc.vq_addr))

		// Driver Status
		runtime.Outw(int(bar0)+0x12, 0x01|0x02)

		// Set Guest Features as Device Features
		features := runtime.Inl(int(bar0) + 0x00)
		runtime.Outw(int(bar0)+0x04, features)

		// Driver Status
		runtime.Outw(int(bar0)+0x12, 0x04)

		// Queue Select
		runtime.Outl(int(bar0)+0x0e, 0)

		// Queue Size (8)
		if runtime.Inl(int(bar0)+0x0c) != 8 {
			if bootstrapped {
				fmt.Println("Virtio PCI RNG: error, unexpected Queue Size!")
			} else {
				runtime.Pmsg("Virtio PCI RNG: error, unexpected Queue Size!\n")
			}

			return
		}

		if bootstrapped {
			fmt.Printf("Virtio PCI RNG: set queue address %x (%x)\n", vqc.vq_addr, int(vqc.vq_addr/4096))
		} else {
			runtime.Pmsg("Virtio PCI RNG: set queue address")
			runtime.Pnum(vqc.vq_addr)
			runtime.Pmsg("\n")
		}

		// Queue Address
		runtime.Outl(int(bar0)+0x08, int((vqc.vq_addr-va_pa_off)/4096))
	}

	vqc.vq.Desc[0].Addr = uint64(r_addr - va_pa_off)
	vqc.vq.Desc[0].Len = uint32(len(r))
	vqc.vq.Desc[0].Flags = 0x02 // Write-Only
	vqc.vq.Avail_idx += 1

	// Queue Notify
	runtime.Outl(int(bar0)+0x10, 0)

	if bootstrapped {
		// We are lazy and we do not implement IOAPIC to trap
		// interrupts, this means we have to wait a little for the
		// buffer to be filled. This is broken pre-bootstrap (as we
		// cannot sleep with standard library calls there).
		time.Sleep(1 * time.Millisecond)
	}

	copy(b, r)
}
