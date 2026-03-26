// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv64

import (
	"strings"

	"github.com/usbarmory/tamago/bits"
)

const extensions = "abcdefghijklmnopqrstuvwxyz"

// defined in features.s
func read_mhartid() uint64
func read_misa() uint64

type Extensions uint64

func (e Extensions) String() string {
	var s []string

	val := uint64(e)

	for i, ext := range extensions {
		if bits.Get64(&val, i) {
			s = append(s, string(ext))
		}
	}

	return strings.Join(s, "")
}

// Features represents the processor capabilities.
type Features struct {
	// Extensions fields as reported by the Machine ISA Register.
	Extensions Extensions
}

// Features returns the processor capabilities.
func (cpu *CPU) Features() (features Features) {
	features.Extensions = Extensions(read_misa())
	return
}
