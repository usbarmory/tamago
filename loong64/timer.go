// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

// nanoseconds per second
const refFreq int64 = 1e9

// Rdtime returns the constant frequency stable counter value.
//
// defined in timer.s
func Rdtime() uint64

// read_cpucfg returns the given CPUCFG configuration word.
//
// defined in timer.s
func read_cpucfg(sel uint64) uint64

// initTimers derives the stable counter frequency from CPUCFG and configures
// the timer scaling factor accordingly.
func (cpu *CPU) initTimers() {
	// CPUCFG word 0x4 holds the constant frequency base, word 0x5 its
	// multiplier (bits 15:0) and divider (bits 31:16).
	base := read_cpucfg(0x4)
	cfg5 := read_cpucfg(0x5)

	mul := cfg5 & 0xffff
	div := (cfg5 >> 16) & 0xffff

	if div == 0 {
		div = 1
	}

	freq := base * mul / div

	if freq == 0 {
		freq = 1
	}

	cpu.TimerMultiplier = float64(refFreq) / float64(freq)
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

	cpu.TimerOffset = ns - int64(float64(cpu.Counter())*cpu.TimerMultiplier)
}
