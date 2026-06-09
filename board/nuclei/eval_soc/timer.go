// Nuclei EvalSoC emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// The Nuclei QEMU nuclei_evalsoc machine does not model the FSL91030 hardware
// CLINT timer, so the FSL91030 SoC nanotime (which reads it) must be excluded
// with the `linknanotime` build tag and replaced by the implementation below,
// which reads the standard RISC-V time CSR. Real UX600 hardware does not
// implement the time CSR, hence this lives in the emulator board package.

package eval_soc

import (
	_ "unsafe"
)

// TIMER_FREQ is the rate of the RISC-V time CSR as modeled by the Nuclei QEMU
// nuclei_evalsoc machine (the `timer_freq` field of its SoC configuration).
// It is independent of the real FSL91030 CLINT rate (fsl91030.RTCCLK).
const TIMER_FREQ = 32768

// defined in timer_riscv64.s
func rdtime() uint64

func mulDiv(x, m, d uint64) uint64 {
	divx := x / d
	modx := x - divx*d
	divm := m / d
	modm := m - divm*d
	return divx*m + modx*divm + modx*modm/d
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return int64(mulDiv(rdtime(), 1e9, TIMER_FREQ))
}
