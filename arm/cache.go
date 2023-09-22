// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// ARM cache register constants
const (
	ACTLR_SMP = 6
)

// defined in cache.s
func read_actlr() uint32
func write_actlr(aux uint32)
func cache_enable()
func cache_disable()
func cache_flush_data()
func cache_flush_instruction()

// EnableSMP sets the SMP bit in Cortex-A7 Auxiliary Control Register, to
// enable coherent requests to the processor. This must be ensured before
// caches and MMU are enabled or any cache and TLB maintenance operations are
// performed (p115, Cortex™-A7 MPCore® Technical Reference Manual r0p5).
func (cpu *CPU) EnableSMP() {
	aux := read_actlr()
	aux |= 1 << ACTLR_SMP
	write_actlr(aux)
}

// EnableCache activates the ARM instruction and data caches.
func (cpu *CPU) EnableCache() {
	cache_enable()
}

// DisableCache disables the ARM instruction and data caches.
func (cpu *CPU) DisableCache() {
	cache_disable()
}

// FlushDataCache flushes the ARM data cache.
func (cpu *CPU) FlushDataCache() {
	cache_flush_data()
}

// FlushInstructionCache flushes the ARM instruction cache.
func (cpu *CPU) FlushInstructionCache() {
	cache_flush_instruction()
}

// FlushTLBs flushes the ARM Translation Lookaside Buffers.
func (cpu *CPU) FlushTLBs() {
	flush_tlb()
}
