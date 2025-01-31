// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"errors"
	"time"
)

const (
	// nanoseconds
	refFreq uint32 = 1e9
)

// defined in timer.s
func kvmclock_pairing() (sec int64, nsec int64, tsc uint64)
func read_tsc() uint64

func (cpu *CPU) initTimers() {
	var timerFreq uint32

	if denominator, numerator, nominalFreq, _ := cpuid(CPUID_TSC_CCC, 0); nominalFreq != 0 {
		timerFreq = (numerator * nominalFreq) / denominator
	}

	if cpu.kvm {
		if khz, _, _, _ := cpuid(KVM_CPUID_TSC_KHZ, 0); khz != 0 {
			timerFreq = khz * 1000
		} else {
			_, nsecA, tscA := kvmclock_pairing()
			_, nsecB, tscB := kvmclock_pairing()
			timerFreq = uint32(((tscB - tscA) * uint64(refFreq)) / uint64(nsecB-nsecA))
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

// Now() returns the KVM host clock.
func (cpu *CPU) Now() (t time.Time, err error) {
	if !cpu.kvm {
		err = errors.New("KVM clock pairing is unavailable")
		return
	}

	sec, nsec, _ := kvmclock_pairing()

	return time.Unix(sec, nsec), nil
}
