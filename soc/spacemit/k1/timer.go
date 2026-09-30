// SpacemiT K1 timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package k1

import (
	"github.com/usbarmory/tamago/riscv64"
	_ "unsafe"
)

// mulDiv computes x * m / d without overflowing on the intermediate product.
func mulDiv(x, m, d uint64) uint64 {
	divx := x / d
	modx := x - divx*d
	divm := m / d
	modm := m - divm*d

	return divx*m + modx*divm + modx*modm/d
}

// Counter returns the number of nanoseconds counted by the RISC-V `time` CSR,
// which is driven by the 24 MHz oscillator (see OSCCLK) as reflected by the
// timebase-frequency device tree property of the vendor boot stages.
//
// The X60 core implements Zicntr (p9, Section 2.1.2.2 Features, K1 Datasheet),
// making the CSR readable at machine level. The memory mapped CLINT mtime
// register is deliberately not used: it does not advance on this SoC, which
// stalls any runtime deadline loop. The vendor U-Boot equally reads the CSR
// rather than the CLINT on 64-bit (sifive_clint_get_count,
// drivers/timer/sifive_clint_timer.c), so it is the time source validated by
// the boot chain.
func Counter() uint64 {
	return mulDiv(riscv64.Rdtime(), 1e9, OSCCLK)
}
