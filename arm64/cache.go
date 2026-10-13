// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

// defined in cache.s
func cache_enable()
func cache_disable()
func cache_flush_data()
func cache_line_size() uint
func cache_clean_data(start, end, line uint)
func cache_invalidate_data(start, end, line uint)

func cache_range(addr uint, size int) (start, end, line uint) {
	line = cache_line_size()
	start = addr & ^(line - 1)
	end = addr + uint(size)
	return
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

// CleanDataCache cleans the data cache covering a memory range.
func (cpu *CPU) CleanDataCache(addr uint, size int) {
	if size <= 0 {
		return
	}

	start, end, line := cache_range(addr, size)
	cache_clean_data(start, end, line)
}

// InvalidateDataCache invalidates the data cache covering a memory range.
func (cpu *CPU) InvalidateDataCache(addr uint, size int) {
	if size <= 0 {
		return
	}

	start, end, line := cache_range(addr, size)
	cache_invalidate_data(start, end, line)
}

// FlushTLBs flushes the ARM Translation Lookaside Buffers.
func (cpu *CPU) FlushTLBs() {
	flush_tlb()
}
