// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv64

// ID returns the processor hardware thread (hart) identifier.
func (cpu *CPU) ID() uint64 {
	return read_mhartid()
}

