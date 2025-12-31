// AMD64 processor support
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

// ACPI PM Timer constants
const (
	ACPI_PM_TIMER_PORT = 0xb008
	ACPI_PM_FREQ       = 3579545
)

// defined in timer.s
func read_tsc() uint64
func write_tsc_deadline(cnt uint64)

// calibrate frequency based on KVM clock
func (cpu *CPU) calibrateByPairing() {
	_, nsecA, tscA := kvmclock.Pairing()
	_, nsecB, tscB := kvmclock.Pairing()

	if den := uint64(nsecB - nsecA); den != 0 {
		cpu.freq = uint32(((tscB - tscA) * uint64(refFreq)) / den)
	}
}

// calibrate frequency based on ACPI PM timer
func (cpu *CPU) calibrateByTimer() {
	var apmB uint32
	var tscB uint64

	loop := ACPI_PM_FREQ / uint32(100)
	mask := uint32(0xffffff)

	apmA := reg.In32(ACPI_PM_TIMER_PORT)
	tscA := read_tsc()

	if apmA & ^mask != 0 {
		// invalid I/O port
		return
	}

	for {
		apmB = reg.In32(ACPI_PM_TIMER_PORT)

		if (apmB-apmA)&mask > loop {
			tscB = read_tsc()
			break
		}
	}

	if den := (apmB - apmA) & mask; den != 0 {
		cpu.freq = uint32((tscB-tscA)/uint64(den)) * ACPI_PM_FREQ
	}

	return
}

func (cpu *CPU) detectCoreFrequency() {
	if den, num, nominalFreq, _ := cpuid(CPUID_TSC_CCC, 0); den != 0 {
		if nominalFreq == 0 {
			baseFreq, _, _, _ := cpuid(CPUID_CPU_FRQ, 0)
			nominalFreq = uint32(uint64(baseFreq) * 1e6 * uint64(den) / uint64(num))
		}

		cpu.freq = uint32((uint64(num) * uint64(nominalFreq)) / uint64(den))
	}

	if cpu.features.KVM {
		if khz, _, _, _ := cpuid(CPUID_KVM_TSC_KHZ, 0); khz != 0 {
			cpu.freq = khz * 1000
		} else {
			cpu.calibrateByPairing()

			if cpu.freq == 0 {
				cpu.calibrateByTimer()
			}
		}
	}

	if cpu.freq != 0 {
		return
	}

	if _, _, ecx, _ := cpuid(CPUID_VENDOR, 0); ecx == CPUID_VENDOR_ECX_AMD {
		if _, ebx, _, _ := cpuid(CPUID_AMD_PROC, 0); bits.IsSet(&ebx, AMD_PROC_CPPC) {
			// Open-Source Register Reference
			// For AMD Family 17h Processors Models 00h-2Fh
			// Rev 3.03 - July, 2018 - Core::X86::Msr::PStateDef
			pstate := uint32(reg.ReadMSR(MSR_AMD_PSTATE))

			num := float64(bits.Get(&pstate, 0, 0xff)) * 25
			den := float64(bits.Get(&pstate, 8, 0b111111)) / 8

			if num != 0 && den != 0 {
				cpu.freq = uint32(num/den) * 1e6
			}
		}
	}

	return
}

func (cpu *CPU) initTimers() {
	cpu.detectCoreFrequency()

	if cpu.freq == 0 {
		print("WARNING: TSC frequency is unavailable\n")
		cpu.freq = 1
	}

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
// nanoseconds value, an interrupt (see [IRQ_WAKEUP] is generated on
// expiration. The timer is enabled only on [CPU] instances supporting
// [Features.TSCDeadline].
func (cpu *CPU) SetAlarm(ns int64) {
	if cpu.TimerMultiplier == 0 || !cpu.features.TSCDeadline {
		return
	}

	// TODO: move to apinit ?
	cpu.LAPIC.SetTimer(IRQ_WAKEUP, lapic.TIMER_MODE_TSC_DEADLINE)

	if ns == 0 {
		cpu.LAPIC.IPI(0, IRQ_WAKEUP, lapic.ICR_DST_REST|lapic.ICR_DLV_IRQ)
		write_tsc_deadline(0)
		return
	}

	cnt := float64(ns-cpu.TimerOffset) / cpu.TimerMultiplier
	write_tsc_deadline(uint64(cnt))
}
