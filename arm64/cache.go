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
func cache_line_size() uint64
func cache_clean_range(start uint64, end uint64, line uint64)
func cache_invalidate_range(start uint64, end uint64, line uint64)

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

// CleanDataCacheRange cleans the data cache lines covering the argument memory
// range.
func (cpu *CPU) CleanDataCacheRange(addr uint, size int) {
	if size <= 0 {
		return
	}

	line := cache_line_size()
	cache_clean_range(uint64(addr)&^(line-1), uint64(addr)+uint64(size), line)
}

// InvalidateDataCacheRange invalidates the data cache lines covering the
// argument memory range, including data outside it that shares those lines.
func (cpu *CPU) InvalidateDataCacheRange(addr uint, size int) {
	if size <= 0 {
		return
	}

	line := cache_line_size()
	cache_invalidate_range(uint64(addr)&^(line-1), uint64(addr)+uint64(size), line)
}

// FlushTLBs flushes the ARM Translation Lookaside Buffers.
func (cpu *CPU) FlushTLBs() {
	flush_tlb()
}
