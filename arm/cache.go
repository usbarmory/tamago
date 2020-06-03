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

func (c *CPU) CacheEnable() {
	cache_enable()
}

func (c *CPU) CacheDisable() {
	cache_disable()
}

func (c *CPU) CacheFlushData() {
	cache_flush_data()
}

func (c *CPU) CacheFlushInstruction() {
	cache_flush_instruction()
}
