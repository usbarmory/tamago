// SiFive FU540 clock control
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fu540

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// Clock registers
const (
	PRCI_BASE = 0x10000000

	PRCI_COREPLLCFG = PRCI_BASE + 0x4
	COREPLL_DIVR    = 0
	COREPLL_DIVF    = 6
	COREPLL_DIVQ    = 15

	PRCI_CORECLKSEL = PRCI_BASE + 0x24
)

// Oscillator frequencies
const (
	// p43, 7.1 Clocking, FU540C00RM
	RTCCLK  = 1000000
	COREPLL = 33330000
)

func init() {
	c := reg.Read(PRCI_COREPLLCFG)

	// set COREPLL for 1 GHz operation
	bits.Clear(&c, COREPLL_DIVR)
	bits.SetN(&c, COREPLL_DIVF, 0x1ff, 59)
	bits.SetN(&c, COREPLL_DIVQ, 0b111, 2)

	reg.Write(PRCI_COREPLLCFG, c)
	reg.Clear(PRCI_CORECLKSEL, 0)
}

// Freq returns the RISC-V core frequency.
func Freq() (hz uint32) {
	if reg.Get(PRCI_CORECLKSEL, 0, 1) == 1 {
		return COREPLL
	}

	// p47, 7.4.2 Setting coreclk frequency, FU540C00RM

	c := reg.Read(PRCI_COREPLLCFG)

	divr := bits.Get(&c, COREPLL_DIVR, 0x3f)
	divf := bits.Get(&c, COREPLL_DIVF, 0x1ff)
	divq := bits.Get(&c, COREPLL_DIVQ, 0b111)

	return (COREPLL * 2 * (divf + 1)) / ((divr + 1) * 1 << divq)
}
