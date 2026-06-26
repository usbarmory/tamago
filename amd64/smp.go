// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"runtime/goos"
	"time"
	"unsafe"

	"github.com/usbarmory/tamago/amd64/lapic"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// ·apinit 16-bit relocation address
	apinitAddress = 0x4000
	// ·apstart pointer address
	apstartAddress = 0x5000

	// AP Global Descriptor Table (GDT) 16-bit address
	gdtAddress = 0x6000
	// AP GDT Descriptor (GDTR) 16-bit address
	gdtrAddress = 0x6018

	// AP task address
	taskAddress = 0x6020

	// CR3 value address
	cr3Address = 0x6028
)

// defined in smp.s
func apinit_reloc(init uintptr, start uintptr)

// task represents a CPU task
type task struct {
	sp uint64 // stack pointer
	gp uint64 // G
	pc uint64 // fn
}

// Write writes the task structure to memory
func (t *task) Write(addr uint) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, t)

	n := buf.Len()
	r, err := dma.NewRegion(addr, n, false)

	if err != nil {
		panic(err)
	}

	_, p := r.Reserve(n, 0)
	defer r.Release(addr)

	copy(p, buf.Bytes())
}

// Task schedules a goroutine on a previously initialized Application Processor
// (see [CPU.InitSMP]).
//
// On `GOOS=tamago` Go scheduler M's are never dropped, therefore the function
// is invoked only once per AP (i.e. GOMAXPROCS-1).
func (cpu *CPU) Task(sp, _, gp, fn unsafe.Pointer) {
	t := &task{
		sp: uint64(uintptr(sp)),
		gp: uint64(uintptr(gp)),
		pc: uint64(uintptr(fn)),
	}

	if t.sp == 0 || t.gp == 0 {
		return
	}

	if cpu.init+1 >= runtime.GOMAXPROCS(-1) {
		panic("Task exceeds available resources")
	}

	t.Write(taskAddress)

	// set last initialized CPU and signal task through NMI
	cpu.init += 1
	cpu.LAPIC.IPI(cpu.init, 0, lapic.ICR_DLV_NMI)
}

// ID returns the processor identifier.
func (cpu *CPU) ID() uint64 {
	return uint64(cpu.LAPIC.ID())
}

// GOMAXPROCS sets the maximum number of CPUs that can be executing
// simultaneously and returns the previous setting, it must be used instead of
// [runtime.GOMAXPROCS] when using this package.
//
// The function checks if n-1 APs are ready (within 1s) at their idle state, if
// the condition fails the current state is not changed.
//
// The function is not required when using [CPU.InitSMP], which invokes it, as
// it is exported for applications that initialize APs by other means.
func (cpu *CPU) GOMAXPROCS(n int) int {
	if cpu.init > 0 {
		return runtime.GOMAXPROCS(n)
	}

	// wait for all APs to reach ·apstart idle state
	if reg.WaitFor(1*time.Second, taskAddress, 0, 0xffffffff, uint32(n-1)) {
		goos.ProcID = cpu.ID
		goos.Task = cpu.Task
		goos.Wake = cpu.Wake
	} else {
		n = 1
	}

	return runtime.GOMAXPROCS(n)
}

// InitSMP enables Secure Multiprocessor (SMP) operation by initializing
// available Application Processors.
//
// A positive argument caps the total (BSP+APs) number of cores, a negative
// argument initializes all available APs, an agument of 0 or 1 disables SMP.
//
// After initialization [runtime.NumCPU] or [runtime.GOMAXPROCS] can be used to
// verify SMP use by the runtime.
func (cpu *CPU) InitSMP(n int) {
	var i int

	if n == 0 || n == 1 {
		goos.Task = nil
		runtime.GOMAXPROCS(1)
		return
	}

	// copy ·apinit to a 16-bit address reachable in real mode
	// copy ·apstart pointer to avoid RIP/EIP-relative addressing
	apinit_reloc(apinitAddress, apstartAddress)

	// create AP Global Descriptor Table (GDT)
	reg.Write64(gdtAddress+0x00, 0x0000000000000000) // null descriptor
	reg.Write64(gdtAddress+0x08, 0x00209a00000fffff) // code descriptor (x/r)
	reg.Write64(gdtAddress+0x10, 0x00009200000fffff) // data descriptor (r/w)

	// create AP GDT Descriptor (GDTR)
	reg.Write16(gdtrAddress+0x00, 3*8-1)      // GTD Limit
	reg.Write32(gdtrAddress+0x02, gdtAddress) // GDT Base Address

	// read fresh CR3 as a cpuinit override might have overridden ours
	reg.Write(cr3Address, uint32(read_cr3()))

	for i = 1; i < NumCPU(); i++ {
		if i == n {
			break
		}

		// AMD64 Architecture Programmer’s Manual
		// Volume 2 - 15.27.8 Secure Multiprocessor Initialization
		//
		// AP Startup Sequence:
		// The vector provides the upper 8 bits of a 20-bit physical address.
		vector := apinitAddress >> 12

		cpu.LAPIC.IPI(i, vector, 1<<lapic.ICR_INIT|lapic.ICR_DLV_INIT)
		time.Sleep(10 * time.Millisecond)

		cpu.LAPIC.IPI(i, vector, 1<<lapic.ICR_INIT|lapic.ICR_DLV_SIPI)
	}

	cpu.GOMAXPROCS(i)
}

// Wake issues a wake interrupt (see [IRQ_WAKEUP]) to the target processor.
func (cpu *CPU) Wake(procid uint64) {
	cpu.LAPIC.IPI(int(procid), IRQ_WAKEUP, lapic.ICR_DLV_IRQ)
}
