// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"github.com/usbarmory/tamago/amd64/lapic"
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/kvm/clock"
)

// nanoseconds
const refFreq uint32 = 1e9

// defined in timer.s
func read_tsc() uint64
func write_tsc_deadline(ns uint64)

func (cpu *CPU) detectCoreFrequency() (freq uint32) {
	if denominator, numerator, nominalFreq, _ := cpuid(CPUID_TSC_CCC, 0); denominator != 0 {
		if nominalFreq == 0 {
			baseFreq, _, _, _ := cpuid(CPUID_CPU_FRQ, 0)
			nominalFreq = uint32(uint64(baseFreq) * 1e6 * uint64(denominator) / uint64(numerator))
		}

		cpu.freq = uint32((uint64(numerator) * uint64(nominalFreq)) / uint64(denominator))
	}

	if cpu.features.KVM {
		if khz, _, _, _ := cpuid(KVM_CPUID_TSC_KHZ, 0); khz != 0 {
			cpu.freq = khz * 1000
		} else {
			_, nsecA, tscA := kvmclock.Pairing()
			_, nsecB, tscB := kvmclock.Pairing()

			if denominator := uint64(nsecB - nsecA); denominator != 0 {
				cpu.freq = uint32(((tscB - tscA) * uint64(refFreq)) / denominator)
			}
		}
	}

	if cpu.freq != 0 {
		return
	}

	if _, _, ecx, _ := cpuid(CPUID_VENDOR, 0); ecx == CPUID_VENDOR_ECX_AMD {
		// Open-Source Register Reference
		// For AMD Family 17h Processors Models 00h-2Fh
		// Rev 3.03 - July, 2018 - Core::X86::Msr::PStateDef
		pstate := reg.Msr(MSR_AMD_PSTATE)

		numerator := float64(bits.Get(&pstate, 0, 0xff)) * 25
		denominator := float64(bits.Get(&pstate, 8, 0b111111)) / 8

		if numerator != 0 && denominator != 0 {
			cpu.freq = uint32(numerator/denominator) * 1e6
		}
	}

	if cpu.freq == 0 {
		print("WARNING: TSC frequency is unavailable\n")
		return 1
	}

	return
}

func (cpu *CPU) initTimers() {
	cpu.detectCoreFrequency()
	cpu.TimerMultiplier = float64(refFreq) / float64(cpu.freq)
}

// Freq() returns the AMD64 core frequency.
func (cpu *CPU) Freq() (hz uint32) {
	return cpu.freq
}

// Counter returns the CPU Time Stamp Counter (TSC).
func (cpu *CPU) Counter() uint64 {
	return read_tsc()
}

// GetTime returns the system time in nanoseconds.
func (cpu *CPU) GetTime() int64 {
	return int64(float64(cpu.Counter())*cpu.TimerMultiplier) + cpu.TimerOffset
}

// SetTime adjusts the system time to the argument nanoseconds value.
func (cpu *CPU) SetTime(ns int64) {
	if cpu.TimerMultiplier == 0 {
		return
	}

	cpu.TimerOffset = ns - int64(float64(read_tsc())*cpu.TimerMultiplier)
}

// SetAlarm sets a physical timer to the absolute time matching the argument
// nanoseconds value, an interrupt is generated on expiration. This function
// has effect only if the [CPU] supports [Features.TSCDeadline].
func (cpu *CPU) SetAlarm(ns int64) {
	if cpu.TimerMultiplier == 0 || !cpu.features.TSCDeadline {
		return
	}

	cpu.LAPIC.SetTimer(IRQ_WAKEUP, lapic.TIMER_MODE_TSC_DEADLINE)

	if ns == 0 {
		write_tsc_deadline(0)
		return
	}

	set := float64(ns-cpu.TimerOffset) / cpu.TimerMultiplier
	write_tsc_deadline(uint64(set))
}
