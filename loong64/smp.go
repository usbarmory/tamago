// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

// ID returns the processor core identifier as reported by the CPUID CSR.
func (cpu *CPU) ID() uint64 {
	return read_cpuid() & 0x1ff
}
