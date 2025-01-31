// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

const (
	// nanoseconds
	refFreq uint32 = 1e9
)

// defined in timer.s
func read_tsc() uint64

func (cpu *CPU) initTimers() {
	var timerFreq uint32

	if denominator, numerator, nominalFreq, _ := cpuid(CPUID_TSC_CCC, 0); nominalFreq != 0 {
		timerFreq = (numerator * nominalFreq) / denominator
	}

	if cpu.kvm {
		if khz, _, _, _ := cpuid(KVM_CPUID_TSC_KHZ, 0); khz != 0 {
			timerFreq = khz * 1000
		}
	}

	if timerFreq == 0 {
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
