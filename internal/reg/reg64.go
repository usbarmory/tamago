// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build arm64 || amd64 || loong64 || riscv64

package reg

import (
	"sync/atomic"
	"unsafe"
)

func Get64(addr uint64, pos int) bool {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	r := atomic.LoadUint64(reg)

	return (r>>pos)&1 == 1
}

func Set64(addr uint64, pos int) {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint64(reg)
	r |= (1 << pos)

	atomic.StoreUint64(reg, r)
}

func SetTo64(addr uint64, pos int, val bool) {
	if val {
		Set64(addr, pos)
	} else {
		Clear64(addr, pos)
	}
}

func Clear64(addr uint64, pos int) {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint64(reg)
	r &= ^(1 << pos)

	atomic.StoreUint64(reg, r)
}

func GetN64(addr uint64, pos int, mask int) uint64 {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	r := atomic.LoadUint64(reg)

	return (r >> pos) & uint64(mask)
}

func Read64(addr uint64) uint64 {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint64(reg)
}

// defined in reg_*.s
func Write64(addr uint64, val uint64)
