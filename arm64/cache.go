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
func cache_line_size() uint64
func cache_invalidate_range(start uint64, end uint64, line uint64)

// EnableCache activates the ARM instruction and data caches.
func (cpu *CPU) EnableCache() {
	cache_enable()
}

// DisableCache disables the ARM instruction and data caches.
func (cpu *CPU) DisableCache() {
	cache_disable()
}

// DataCacheLineSize returns the smallest data cache line size in bytes.
func (cpu *CPU) DataCacheLineSize() int {
	return int(cache_line_size())
}

// InvalidateDataCacheRange discards data cache lines covering the address
// range, so that the CPU reads memory written by a device which is not cache
// coherent. Invalidation also discards dirty data sharing those lines, so the
// range must cover complete lines owned by the caller. The range
// must be invalidated after the device write completes, as the CPU may
// speculatively fetch lines while it is in progress.
func (cpu *CPU) InvalidateDataCacheRange(addr uint, size int) {
	if size <= 0 {
		return
	}

	line := cache_line_size()
	start := uint64(addr) &^ (line - 1)

	cache_invalidate_range(start, uint64(addr)+uint64(size), line)
}

// FlushTLBs flushes the ARM Translation Lookaside Buffers.
func (cpu *CPU) FlushTLBs() {
	flush_tlb()
}
