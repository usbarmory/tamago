// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

import (
	"sync/atomic"
	"unsafe"
)

func IsSet64(addr uint64, pos int) bool {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	r := atomic.LoadUint64(reg)

	return (int(r)>>pos)&1 == 1
}

func Read64(addr uint64) uint64 {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint64(reg)
}

// defined in reg_*.s
func Write64(addr uint64, val uint64)
