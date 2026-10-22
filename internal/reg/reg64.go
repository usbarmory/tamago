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
	r := Read64(addr)

	return (r>>pos)&1 == 1
}

func Set64(addr uint64, pos int) {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))

	r := Read64(addr)
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

	r := Read64(addr)
	r &= ^(1 << pos)

	atomic.StoreUint64(reg, r)
}

func GetN64(addr uint64, pos int, mask int) uint64 {
	r := Read64(addr)

	return (r >> pos) & uint64(mask)
}

func SetN64(addr uint64, pos int, mask int, val uint64) {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))

	r := Read64(addr)
	r = (r & (^(uint64(mask) << pos))) | ((val & uint64(mask)) << pos)

	atomic.StoreUint64(reg, r)
}

func Read64(addr uint64) uint64 {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint64(reg)
}

func Read64As32(addr uint64) uint32 {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint32(reg)
}

// defined in reg_*.s to avoid register read
func Write64(addr uint64, val uint64)

// Write64As32 uses a plain store, as sync/atomic stores might read the
// register (see Write), therefore, as Write, it does not enforce memory
// ordering. It must not be inlined to ensure that the store is issued exactly
// once and with 32-bit width.
//
//go:noinline
func Write64As32(addr uint64, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	*reg = val
}
