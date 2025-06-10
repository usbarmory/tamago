// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"runtime"

	"github.com/usbarmory/tamago/amd64/lapic"
)

// InitSMP enables Secure Multiprocessor (SMP) operation by initializing the
// available Application Processors (see [CPU.APs]).
//
// A positive argument caps the total (BSP+APs) number of cores, a negative
// argument initializes all available APs, an agument of 0 or 1 disables SMP.
//
// After initialization [runtime.NumCPU()] can be used to verify SMP use by the
// runtime.
func (cpu *CPU) InitSMP(n int) (aps []*CPU) {
	cpu.APs = nil

	defer func() {
		// TODO: WiP
		//runtime.GOMAXPROCS(1+len(cpu.APs))
		runtime.GOMAXPROCS(1)
	}()

	if n == 0 || n == 1 {
		return
	}

	for i := 1; i < NumCPU(); i++ {
		if i == n {
			break
		}

		ap := &CPU{
			TimerMultiplier: cpu.TimerMultiplier,
			LAPIC: &lapic.LAPIC{
				Base: cpu.LAPIC.Base,
			},
		}

		cpu.APs = append(cpu.APs, ap)
	}

	return
}
