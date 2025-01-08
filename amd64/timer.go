// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"github.com/usbarmory/tamago/bits"
)

const (
	// nanoseconds
	refFreq uint32 = 1000000000
)

// defined in timer.s
func read_tsc() int64

func (cpu *CPU) initTimers() {
	var timerFreq uint32

	_, _, _, apmFeatures := cpuid(CPUID_APM, 0x0)

	if !bits.IsSet(&apmFeatures, APM_TSC_INVARIANT) {
		panic("TSC is not invariant")
	}

	if denominator, numerator, nominalFreq, _ := cpuid(CPUID_TSC_CCC, 0x00); nominalFreq != 0 {
		timerFreq = (numerator * nominalFreq) / denominator
	} else if khz, _, _, _ := cpuid(CPUID_HYP_TSC_KHZ, 0x00); khz != 0 {
		timerFreq = khz * 1000
	} else {
		panic("TSC frequency is unavailable")
	}

	cpu.TimerMultiplier = float64(refFreq) / float64(timerFreq)
	cpu.TimerFn = read_tsc
}

// SetTimer sets the timer to the argument nanoseconds value.
func (cpu *CPU) SetTimer(t int64) {
	if cpu.TimerFn == nil || cpu.TimerMultiplier == 0 {
		return
	}

	cpu.TimerOffset = t - int64(float64(cpu.TimerFn())*cpu.TimerMultiplier)
}
