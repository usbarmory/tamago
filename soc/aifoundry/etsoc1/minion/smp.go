// AI Foundry Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package minion

import (
	"runtime"
	"runtime/goos"
	"time"
	"unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/aifoundry/etsoc1"
)

// MaxTasks is used to derive the stack size allocated for each [Task], using
// the following formula: `ramStackOffset / MaxTasks`.
var MaxTasks uint64 = 64

var (
	// hart initialization counter
	ncpu int

	// AP task address base, accounting for [alignExceptionHandler]
	taskBase      = goos.RamStart + 4*2
	taskStackSize = ramStackOffset / MaxTasks
)

// Workload represents a function for execution on a target hardware thread
// (hart) by [TaskWorkload].
//
// When tasked on a different hartm the function is called with no cache
// coherency and no Go scheduler M allocation, for this reason extra caution
// must be taken in defining the Workload function.
//
// A pure Go assembly Workload is recommended, though with extra care (and no
// guarantees of backward/future compatiblity) simple freestanding Go code can
// be executed.
type Workload func()

func (fn Workload) vector() uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

// task represents a CPU task (used in smp.s)
type task struct {
	sp uint64 // stack pointer
	gp uint64 // G
	pc uint64 // fn
}

func schedule(hart int, gp uint64, pc uint64, wait bool) {
	// use [ramEnd-ramStackOffset:ramEnd] for tasks stack
	stk := goos.RamStart + goos.RamSize
	sp := uint64(stk) - uint64(hart)*(ramStackOffset/MaxTasks)

	// write directly to memory to avoid &task allocation
	taskAddress := uint64(taskBase) + uint64(hart*24)
	reg.Write64(taskAddress+0, sp)
	reg.Write64(taskAddress+8, gp)
	reg.Write64(taskAddress+16, pc)

	// signal task through IPI
	etsoc1.IPI(hart)

	if !wait {
		return
	}

	for reg.Get64(etsoc1.IPI_TRIGGER, hart) {
		// yield to scheduler without holding locks
		time.Sleep(1)
	}
}

// TaskWorkload schedules a [Workload] on a previously initialized hardware
// thread (hart), optionally waiting for its completion.
//
// On the ET-Minion platform `GOOS=tamago` implementation the Go runtime lives
// only on hart 0, see [Workload] for implications on its definition.
func TaskWorkload(hart int, fn Workload, wait bool) {
	if fn == nil {
		return
	}

	if uint64(hart) == RV64.ID() {
		fn()
		return
	}

	gp := uint64(uintptr(runtime.GetG()))
	pc := fn.vector()

	schedule(hart, gp, pc, wait)
}

// Task schedules execution of an arbitrary program counter on a previously
// initialized hardware thread (hart), optionally waiting for its completion.
func Task(hart int, pc uint64, wait bool) {
	schedule(hart, 0, pc, wait)
}

// NumCPU returns the number of logical CPUs initialized on the platform.
func NumCPU() (n int) {
	return ncpu
}
