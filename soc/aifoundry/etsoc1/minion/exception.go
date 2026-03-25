// AI Foundry ET-SoC-1 Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package minion

import (
	"unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/aifoundry/etsoc1"
)

func vector(fn riscv64.ExceptionHandler) uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

func encodeLongJump(ptr, pc uint64) uint64 {
	off := uint64(ptr) - uint64(pc)

	hi := uint32(off + 0x800) >> 12
	lo := uint32(off & 0xfff)

	auipc := uint32(0x17) | (6 << 7) | (hi << 12)
	jalr := uint32(0x67) | (0 << 7) | (0 << 12) | (6 << 15) | (lo << 20)

	return uint64(auipc) | uint64(jalr) << 32
}

// SetAlignedExceptionHandler updates the CPU machine trap vector with the
// address of the argument function, honoring the ET-Minion requirement for a 4
// KB aligned handler.
func SetAlignedExceptionHandler(fn riscv64.ExceptionHandler) {
	src := uint64(etsoc1.DRAM_BASE)
	dst := vector(fn)

	reg.Write64(src, encodeLongJump(dst, src))
	RV64.SetExceptionHandlerAddress(src)
}
