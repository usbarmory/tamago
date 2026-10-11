// Data cache maintenance by address range
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cache

// defined in cache_arm64.s
func cache_type() uint64
func cache_line_size() uint64
func cache_clean_data(start, end, line uint64)
func cache_clean_invalidate_data(start, end, line uint64)
func cache_invalidate_data(start, end, line uint64)

const maxWritebackGranule = 512 * 4

func cache_range(addr uint64, size int) (start, end, line uint64) {
	line = cache_line_size()
	start = addr & ^(line - 1)
	end = addr + uint64(size)
	return
}

// WritebackGranule returns the data cache writeback granule size.
func WritebackGranule() uint64 {
	if cwg := (cache_type() >> 24) & 0xf; cwg != 0 {
		return 4 << cwg
	}

	return maxWritebackGranule
}

// Clean cleans the data cache covering a memory range.
func Clean(addr uint64, size int) {
	if size <= 0 {
		return
	}

	start, end, line := cache_range(addr, size)
	cache_clean_data(start, end, line)
}

// Invalidate invalidates the data cache covering a memory range.
func Invalidate(addr uint64, size int) {
	if size <= 0 {
		return
	}

	start, end, line := cache_range(addr, size)

	if start != addr {
		cache_clean_invalidate_data(start, start+1, line)
		start += line
	}

	if tail := end &^ (line - 1); tail != end && tail >= start {
		cache_clean_invalidate_data(tail, tail+1, line)
		end = tail
	}

	if start < end {
		cache_invalidate_data(start, end, line)
	}
}
