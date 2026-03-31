// AI Foundry ET-SoC-1 Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package minion

import (
	"runtime/goos"
	"unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/riscv64"
)

func vector(fn riscv64.ExceptionHandler) uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

func encodeLongJump(ptr, pc uint64) uint64 {
	off := uint64(ptr) - uint64(pc)
	hi := uint32(off+0x800) >> 12
	lo := uint32(off & 0xfff)

	// Volume I: RISC-V Unprivileged ISA V20191213
	// RV32I Base Instruction Set
	r := uint32(6)
	auipc := (hi << 12) | (r << 7) | uint32(0b10111)
	jalr := (lo << 20) | (r << 15) | uint32(0b1100111)

	return uint64(auipc) | uint64(jalr)<<32
}

func alignExceptionHandler() {
	// RamStart is naturally aligned to 0x1000
	src := uint64(goos.RamStart)
	dst := RV64.GetExceptionHandlerAddress()

	reg.Write64(src, encodeLongJump(dst, src))
	RV64.SetExceptionHandlerAddress(src)
}
