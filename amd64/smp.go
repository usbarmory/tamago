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
	"time"

	"github.com/usbarmory/tamago/amd64/lapic"
)

const initAP = 0x8000

// InitSMP enables Secure Multiprocessor (SMP) operation by initializing the
// available Application Processors (see [CPU.APs]).
//
// A positive argument caps the total (BSP+APs) number of cores, a negative
// argument initializes all available APs, an agument of 0 or 1 disables SMP.
//
// After initialization [runtime.NumCPU()] can be used to verify SMP use by the
// runtime.
func (cpu *CPU) InitSMP(n int) (aps []*CPU) {
	cpu.aps = nil

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

		// AMD64 Architecture Programmer’s Manual 
		// Volume 2 - 15.27.8 Secure Multiprocessor Initialization
		//
		// AP Startup Sequence:
		// The vector provides the upper 8 bits of a 20-bit physical address.
		vector := initAP >> 12

		// startup AP
		cpu.LAPIC.IPI(i, vector, (1 << 14) | lapic.ICR_INIT)
		time.Sleep(1 * time.Millisecond)
		cpu.LAPIC.IPI(i, vector, (1 << 14) | lapic.ICR_SIPI)

		cpu.aps = append(cpu.aps, ap)
	}

	return
}
