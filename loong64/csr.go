// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

// Control and Status Register accessors, defined in csr.s.
func read_crmd() uint64
func write_crmd(val uint64)
func read_ecfg() uint64
func write_ecfg(val uint64)
func read_estat() uint64
func read_era() uint64
func read_badv() uint64
func set_eentry(addr uint64)
func read_cpuid() uint64
