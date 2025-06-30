// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"time"
	"unsafe"

	"github.com/usbarmory/tamago/amd64/lapic"
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// ·apinit 16-bit relocation address
	apinitAddress = 0x4000
	// AP Global Descriptor Table (GDT) 16-bit address
	gdtAddress = 0x5000
	// AP GDT Descriptor (GDTR) 16-bit address
	gdtrAddress = 0x5018

	procAddress = 0x5020
)

// Proc represents a CPU task
type Proc struct {
	sp uint64
	mp uint64
	gp uint64
	pc uint64
}

// Bytes converts the descriptor structure to byte array format.
func (d *Proc) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// defined in smp.s
func apinit_reloc(addr uintptr)

// NumCPU returns the number of logical CPUs.
func NumCPU() (n int) {
	_, _, ecx, _ := cpuid(CPUID_VENDOR, 0)

	switch ecx {
	case CPUID_VENDOR_ECX_AMD:
		// AMD64 Architecture Programmer’s Manual
		// Volume 3 - E4.7
		_, _, ecx, _ := cpuid(CPUID_AMD_PROC, 0)
		n = int(bits.Get(&ecx, AMD_PROC_NC, 0xff)) + 1
	case CPUID_VENDOR_ECX_INTEL:
		// Intel® Architecture Instruction Set Extensions and Future
		// Features Programming Reference
		// 5.1.12 x2APIC Features / Processor Topology (Function 0Bh)
		_, ebx, _, _ := cpuid(CPUID_INTEL_APIC, 1) // core sublevel
		n = int(bits.Get(&ebx, INTEL_APIC_LP, 0xffff))
	}

	if n == 0 {
		n = runtime.NumCPU()
	}

	return
}

func (bsp *CPU) task(sp, mp, gp, fn unsafe.Pointer) {
	proc := &Proc{
		sp: uint64(uintptr(sp)),
		mp: uint64(uintptr(mp)),
		gp: uint64(uintptr(gp)),
		pc: uint64(uintptr(fn)),
	}

	buf := proc.Bytes()
	r, err := dma.NewRegion(procAddress, len(buf), false)

	if err != nil {
		panic(err)
	}

	addr, p := r.Reserve(len(buf), 0)
	defer r.Release(addr)

	copy(p, buf)

	bsp.LAPIC.IPI(1, 255, lapic.ICR_NMI) // FIXME: ?
}

// InitSMP enables Secure Multiprocessor (SMP) operation by initializing the
// available Application Processors (see [CPU.APs]).
//
// A positive argument caps the total (BSP+APs) number of cores, a negative
// argument initializes all available APs, an agument of 0 or 1 disables SMP.
//
// After initialization [runtime.NumCPU()] can be used to verify SMP use by the
// runtime.
func (bsp *CPU) InitSMP(n int) (aps []*CPU) {
	bsp.aps = nil

	defer func() {
		runtime.Task = bsp.task
		runtime.GOMAXPROCS(1+len(bsp.aps))
	}()

	if n == 0 || n == 1 {
		return
	}

	// copy ·apinit to a 16-bit address reachable in real mode
	apinit_reloc(apinitAddress)

	// create AP Global Descriptor Table (GDT)
	reg.Write64(gdtAddress+0x00, 0x0000000000000000) // null descriptor
	reg.Write64(gdtAddress+0x08, 0x00209a00000fffff) // code descriptor (x/r)
	reg.Write64(gdtAddress+0x10, 0x00009200000fffff) // data descriptor (r/w)

	// create AP GDT Descriptor (GDTR)
	reg.Write16(gdtrAddress+0x00, 3*8-1)      // GTD Limit
	reg.Write32(gdtrAddress+0x02, gdtAddress) // GDT Base Address

	for i := 1; i < NumCPU(); i++ {
		if i == n {
			break
		}

		ap := &CPU{
			TimerMultiplier: bsp.TimerMultiplier,
			LAPIC: &lapic.LAPIC{
				Base: bsp.LAPIC.Base,
			},
		}

		// AMD64 Architecture Programmer’s Manual
		// Volume 2 - 15.27.8 Secure Multiprocessor Initialization
		//
		// AP Startup Sequence:
		// The vector provides the upper 8 bits of a 20-bit physical address.
		vector := apinitAddress >> 12

		bsp.LAPIC.IPI(i, vector, (1<<16)|(1<<14)|lapic.ICR_INIT)
		time.Sleep(10 * time.Millisecond)

		bsp.LAPIC.IPI(i, vector, (1<<14)|lapic.ICR_SIPI)
		bsp.aps = append(bsp.aps, ap)
	}

	return
}
