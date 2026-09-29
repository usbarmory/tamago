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

	"github.com/usbarmory/tamago/internal/reg"
)

func encodeLongJump(ptr, pc uint64) uint64 {
	off := uint64(ptr) - uint64(pc)
	hi := uint32(off+0x800) >> 12
	lo := uint32(off & 0xfff)

	// gp (x3) is used as scratch register as it is never used by Go, this
	// avoids clobbering registers which must be preserved on interrupts.
	r := uint32(3)

	// Volume I: RISC-V Unprivileged ISA V20191213
	// RV32I Base Instruction Set
	auipc := (hi << 12) | (r << 7) | uint32(0b0010111)
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
