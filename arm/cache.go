// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// defined in cache.s
func cache_enable()
func cache_disable()
func cache_flush_data()
func cache_flush_instruction()

// CacheEnable activates the ARM MMU instruction and data caches.
func (cpu *CPU) CacheEnable() {
	cache_enable()
}

// CacheDisable disables the ARM MMU instruction and data caches.
func (cpu *CPU) CacheDisable() {
	cache_disable()
}

// CacheFlushData flushes the ARM MMU data cache.
func (cpu *CPU) CacheFlushData() {
	cache_flush_data()
}

// CacheFlushData flushes the ARM MMU instruction cache.
func (cpu *CPU) CacheFlushInstruction() {
	cache_flush_instruction()
}
