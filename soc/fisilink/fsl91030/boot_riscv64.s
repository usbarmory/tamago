// Fisilink FSL91030 first-stage boot initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// This file overrides the default riscv64 cpuinit when built with
// the 'linkcpuinit' build tag. It performs full hardware initialization
// required when TamaGo runs as the first-stage boot loader directly from
// SPI flash:
//
//   1. Disable Nuclei UX600 I/D cache (CSR_MCACHE_CTL)
//   2. Initialize DDR SDRAM controller (ported from freeloader.S)
//   3. Configure QSPI0 clock divider
//   4. Enable I/D cache
//   5. Disable interrupts, enable FPU (standard riscv64 setup)
//   6. Set stack pointer to top of DRAM using runtime/goos variables
//   7. Jump to _rt0_tamago_start
//
// Register usage during DDR init (no stack available):
//   X14 (A4): DDR controller base address (0x10001000), constant
//   X15 (A5): working register, DDR config values
//   X12 (A2): 8192 = 0x2000, used for multiple timing registers
//   X13 (A3): derived timing value, reused for REG3
//   X11 (A1): working register, more timing/config values

//go:build linkcpuinit

#include "textflag.h"

// Register number for CSR encoding macros
#define t0 5

// CSR instruction encodings (Plan 9 Go assembler WORD directives)
// CSRS(RS, CSR): set bits in CSR using RS  (csrrs x0, csr, rs)
#define CSRS(RS,CSR) WORD $(0x2073 + RS<<15 + CSR<<20)
// CSRC(RS, CSR): clear bits in CSR using RS (csrrc x0, csr, rs)
#define CSRC(RS,CSR) WORD $(0x3073 + RS<<15 + CSR<<20)
// CSRW(RS, CSR): write CSR from RS          (csrrw x0, csr, rs)
#define CSRW(RS,CSR) WORD $(0x1073 + RS<<15 + CSR<<20)

// CSR addresses
#define sie      0x104
#define mstatus  0x300
#define mie      0x304

// Nuclei UX600 custom CSR for cache control
#define mcachectl 0x7CA
// Bit pattern to enable/disable I+D cache
#define CACHE_EN  0x10001

// TEXT cpuinit(SB): first code executed, called from runtime entry before
// any Go world setup. No stack is available; all state lives in registers.
TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0

	// ---------------------------------------------------------------
	// 1. Disable Nuclei I/D cache before DDR init (freeloader.S line 21-22)
	// ---------------------------------------------------------------
	MOV	$CACHE_EN, T0
	CSRC	(t0, mcachectl)

	// ---------------------------------------------------------------
	// 2. DDR SDRAM initialization (freeloader.S lines 29-101)
	//    Ported to Go assembler. No stack used; registers only.
	//    X14 = DDR base (constant), X15/X12/X13/X11 = working regs.
	// ---------------------------------------------------------------

	// Set DDR base address
	MOV	$0x10001000, X14

	// sw zero, 20(a4): clear DDR_STATUS before polling
	MOVW	X0, 20(X14)

	// .L2: wait until DDR_STATUS bit 1 is set (controller ready)
ddr_wait_ready:
	MOVW	20(X14), X15
	AND	$2, X15, X15
	BEQ	X15, X0, ddr_wait_ready

	// DDR_MODE (offset 512): 134258688 + (-765) = 0x80097D03 as 32-bit
	MOV	$134258688, X15
	ADDW	$-765, X15, X15
	MOVW	X15, 512(X14)

	// DDR_CTRL (offset 516): 0  (clear before triggering init)
	MOVW	X0, 516(X14)

	// DDR_TIMING0 (offset 520): 270336 + (-912)
	MOV	$270336, X15
	ADDW	$-912, X15, X15
	MOVW	X15, 520(X14)

	// DDR_TIMING1 (offset 524): 24
	MOV	$24, X15
	MOVW	X15, 524(X14)

	// DDR_TIMING2 (offset 528): -2147483648 + 84
	MOV	$-2147483648, X15
	ADDW	$84, X15, X15
	MOVW	X15, 528(X14)

	// DDR_TIMING3 (offset 532): 520556544 + (-1529)
	MOV	$520556544, X15
	ADDW	$-1529, X15, X15
	MOVW	X15, 532(X14)

	// DDR_TIMING4 (offset 536): 289738752 + 530
	// Also init X12 = 8192 here (used for TIMING5 and REG2)
	MOV	$289738752, X15
	MOV	$8192, X12
	ADDW	$530, X15, X15
	MOVW	X15, 536(X14)

	// DDR_TIMING5 (offset 540): X12 + 48 = 8240  -> stored in X13
	ADDW	$48, X12, X13
	MOVW	X13, 540(X14)

	// DDR_TIMING6 (offset 544): 28672 + (-254)  -> reuse X13
	// X15 is overwritten here with 28672; X15 is later used for DDR_REG6
	MOV	$28672, X15
	ADDW	$-254, X15, X13
	MOVW	X13, 544(X14)

	// DDR_TIMING7 (offset 548): 40960 + (-1639) -> X13 (saved for DDR_REG3)
	MOV	$40960, X13
	ADDW	$-1639, X13, X13
	MOVW	X13, 548(X14)

	// DDR_TIMING8 (offset 552): 4096 + (-839)
	MOV	$4096, X11
	ADDW	$-839, X11, X11
	MOVW	X11, 552(X14)

	// DDR_TIMING9 (offset 556): 0
	MOVW	X0, 556(X14)

	// DDR_TIMINGA (offset 560): -1879048192 + 4
	MOV	$-1879048192, X11
	ADDW	$4, X11, X11
	MOVW	X11, 560(X14)

	// DDR_TIMINGB (offset 564): 50528256 + 771
	MOV	$50528256, X11
	ADDW	$771, X11, X11
	MOVW	X11, 564(X14)

	// DDR_TIMINGC (offset 568): same value as TIMINGB
	MOVW	X11, 568(X14)

	// DDR_TIMINGD (offset 572): 262144 + 19
	MOV	$262144, X11
	ADDW	$19, X11, X11
	MOVW	X11, 572(X14)

	// DDR_REG0 (offset 608): 8
	MOV	$8, X11
	MOVW	X11, 608(X14)

	// DDR_REG1 (offset 624): 255
	MOV	$255, X11
	MOVW	X11, 624(X14)

	// DDR_REG2 (offset 628): X12 + 546 = 8192 + 546 = 8738
	ADDW	$546, X12, X12
	MOVW	X12, 628(X14)

	// DDR_REG3 (offset 632): X13 (= DDR_TIMING7 value = 39321)
	MOVW	X13, 632(X14)

	// DDR_REG4 (offset 680): 16384 + (-384)
	MOV	$16384, X13
	ADDW	$-384, X13, X13
	MOVW	X13, 680(X14)

	// DDR_REG5 (offset 684): 401408 + (-1408)
	MOV	$401408, X13
	ADDW	$-1408, X13, X13
	MOVW	X13, 684(X14)

	// DDR_REG6 (offset 816): X15 (= 28672) + 1911 = 30583
	// X15 was set to 28672 at DDR_TIMING6 and not modified since
	ADDW	$1911, X15, X15
	MOVW	X15, 816(X14)

	// DDR_REG7 (offset 824): 5
	MOV	$5, X15
	MOVW	X15, 824(X14)

	// DDR_CTRL (offset 516): 1 -> trigger DDR initialization
	MOV	$1, X15
	MOVW	X15, 516(X14)

	// .L3: wait until DDR_CTRL bit 8 (init done) is set
	MOV	$0x10001000, X14
ddr_wait_init:
	MOVW	516(X14), X15
	AND	$256, X15, X15
	BEQ	X15, X0, ddr_wait_init

	// ---------------------------------------------------------------
	// 3. Configure QSPI0 clock divider (freeloader.S lines 103-107)
	//    Sets QSPI0 SCKDIV register (offset 100 = 0x64) to 0x130009
	// ---------------------------------------------------------------
	MOV	$1245184, X15
	MOV	$0x10014000, X14
	ADDW	$9, X15, X15
	MOVW	X15, 100(X14)

	// ---------------------------------------------------------------
	// 4. Enable Nuclei I/D cache (freeloader.S lines 120-122)
	// ---------------------------------------------------------------
	MOV	$CACHE_EN, T0
	CSRS	(t0, mcachectl)

	// ---------------------------------------------------------------
	// 5. Standard riscv64 setup: disable interrupts, enable FPU
	// ---------------------------------------------------------------
	MOV	$0, T0
	CSRW	(t0, sie)
	CSRW	(t0, mie)
	MOV	$0x7FFF, T0
	CSRC	(t0, mstatus)
	MOV	$(1<<13), T0
	CSRS	(t0, mstatus)

	// ---------------------------------------------------------------
	// 6. Set stack pointer to top of DRAM
	//    SP = RamStart + RamSize - RamStackOffset
	// ---------------------------------------------------------------
	MOV	runtime∕goos·RamStart(SB), X2
	MOV	runtime∕goos·RamSize(SB), T1
	MOV	runtime∕goos·RamStackOffset(SB), T2
	ADD	T1, X2
	SUB	T2, X2

	// ---------------------------------------------------------------
	// 7. Jump into Go runtime
	// ---------------------------------------------------------------
	JMP	_rt0_tamago_start(SB)
