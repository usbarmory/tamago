// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv

import (
	"errors"
	"sync"

	"github.com/usbarmory/tamago/bits"
)

// PMP CSRs constants
// (3.7.1 Physical Memory Protection CSRs
// RISC-V Privileged Architectures V20211203).
const (
	PMP_CFG_L = 7 // lock
	PMP_CFG_A = 3 // address-matching mode
	PMP_CFG_X = 2 // execution access
	PMP_CFG_W = 1 // write access
	PMP_CFG_R = 0 // read access

	PMP_A_OFF   = 0 // Null region (disabled)
	PMP_A_TOR   = 1 // Top of range
	PMP_A_NA4   = 2 // Naturally aligned four-byte region
	PMP_A_NAPOT = 3 // Naturally aligned power-of-two region, ≥8 bytes
)

// PMP CSRs helpers for RV64, only 8 PMPs are supported for now. In the future,
// to support up to 64 PMPs, this will benefit from dynamic generation with
// go:generate.

// defined in pmp.s
func read_pmpcfg0() uint64
func read_pmpaddr0() uint64
func read_pmpaddr1() uint64
func read_pmpaddr2() uint64
func read_pmpaddr3() uint64
func read_pmpaddr4() uint64
func read_pmpaddr5() uint64
func read_pmpaddr6() uint64
func read_pmpaddr7() uint64
func write_pmpcfg0(uint64)
func write_pmpaddr0(uint64)
func write_pmpaddr1(uint64)
func write_pmpaddr2(uint64)
func write_pmpaddr3(uint64)
func write_pmpaddr4(uint64)
func write_pmpaddr5(uint64)
func write_pmpaddr6(uint64)
func write_pmpaddr7(uint64)

var mux sync.Mutex

// ReadPMP returns the Physical Memory Protection CSRs, configuration and
// address, for the relevant index (currently limited to PMPs from 0 to 7).
func (cpu *CPU) ReadPMP(i int) (addr uint64, r bool, w bool, x bool, a int, l bool, err error) {
	mux.Lock()
	defer mux.Unlock()

	switch i {
	case 0:
		addr = read_pmpaddr0()
	case 1:
		addr = read_pmpaddr1()
	case 2:
		addr = read_pmpaddr2()
	case 3:
		addr = read_pmpaddr3()
	case 4:
		addr = read_pmpaddr4()
	case 5:
		addr = read_pmpaddr5()
	case 6:
		addr = read_pmpaddr6()
	case 7:
		addr = read_pmpaddr7()
	default:
		err = errors.New("unsupported PMP index")
		return
	}

	// addr holds bits 55:2
	addr = addr << 2

	cfg := read_pmpcfg0()
	off := i * 8

	r = bits.Get64(&cfg, off+PMP_CFG_R, 1) == 1
	w = bits.Get64(&cfg, off+PMP_CFG_W, 1) == 1
	x = bits.Get64(&cfg, off+PMP_CFG_X, 1) == 1
	l = bits.Get64(&cfg, off+PMP_CFG_L, 1) == 1
	a = int(bits.Get64(&cfg, off+PMP_CFG_A, 0b11))

	return
}

// WritePMP sets the Physical Memory Protection CSRs, configuration and
// address, for the relevant index (currently limited to PMPs from 0 to 7).
func (cpu *CPU) WritePMP(i int, addr uint64, r bool, w bool, x bool, a int, l bool) (err error) {
	mux.Lock()
	defer mux.Unlock()

	// addr holds bits 55:2
	addr = addr >> 2

	switch i {
	case 0:
		write_pmpaddr0(addr)
	case 1:
		write_pmpaddr1(addr)
	case 2:
		write_pmpaddr2(addr)
	case 3:
		write_pmpaddr3(addr)
	case 4:
		write_pmpaddr4(addr)
	case 5:
		write_pmpaddr5(addr)
	case 6:
		write_pmpaddr6(addr)
	case 7:
		write_pmpaddr7(addr)
	default:
		err = errors.New("unsupported PMP index")
		return
	}

	cfg := read_pmpcfg0()
	off := i * 8

	bits.SetTo64(&cfg, off+PMP_CFG_R, r)
	bits.SetTo64(&cfg, off+PMP_CFG_W, w)
	bits.SetTo64(&cfg, off+PMP_CFG_X, x)
	bits.SetTo64(&cfg, off+PMP_CFG_L, l)
	bits.SetN64(&cfg, off+PMP_CFG_A, 0b11, uint64(a))

	write_pmpcfg0(cfg)

	return
}
