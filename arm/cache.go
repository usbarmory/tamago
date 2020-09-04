// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// ARM cache register constants
const (
	ACTLR_SMP = 6
)

// defined in cache.s
func read_actlr() int32
func write_actlr(aux int32)
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
	aux |= (1 << ACTLR_SMP)
	write_actlr(aux)
}

// CacheEnable activates the ARM instruction and data caches.
func (cpu *CPU) CacheEnable() {
	cache_enable()
}

// CacheDisable disables the ARM instruction and data caches.
func (cpu *CPU) CacheDisable() {
	cache_disable()
}

// CacheFlushData flushes the ARM data cache.
func (cpu *CPU) CacheFlushData() {
	cache_flush_data()
}

// CacheFlushInstruction flushes the ARM instruction cache.
func (cpu *CPU) CacheFlushInstruction() {
	cache_flush_instruction()
}
