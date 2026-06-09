// Fisilink FSL91030 hardware boot initialization (linkcpuinit)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// This file overrides the default riscv64 cpuinit when built with the
// 'linkcpuinit' build tag, performing the hardware initialization (DDR,
// cache, QSPI clock, FPU, stack) that a first-stage boot loader would
// otherwise do. It suits loading the image directly into DRAM (e.g. via
// JTAG); for NOR boot a relocating stage is required instead.
//
// The register offsets, the cache-enable value and the DDR controller
// timing/configuration values are taken verbatim from the Nuclei freeloader.S
// (Copyright Nuclei System Technologies) shipped with the FSL91030 vendor SDK.
// The DDR timing values are opaque vendor-provided constants with no documented
// bit-level meaning and are therefore reproduced as-is; the offsets they are
// written to are named below for clarity.
//
// Sequence: disable cache, initialize DDR, configure the QSPI0 clock,
// enable cache, disable interrupts and enable the FPU, set the stack to the
// top of DRAM and jump to the runtime entry. No stack is available during
// DDR init; register usage:
//   X14 (A4): DDR controller base, constant
//   X15 (A5): working register, DDR config values
//   X12 (A2): 8192, used for multiple timing registers
//   X13 (A3): derived timing value, reused for REG3
//   X11 (A1): working register, more timing/config values

//go:build linkcpuinit

#include "textflag.h"

// Nuclei vendor-specific machine cache control CSR (mcache_ctl, 0x7ca) is
// unknown to the Go assembler, so accesses to it are WORD-encoded below:
//   csrrs x0, 0x7ca, t0 = 0x7ca2a073
//   csrrc x0, 0x7ca, t0 = 0x7ca2b073
#define MCACHE_CTL_IC_EN 0  // bit 0:  instruction cache enable
#define MCACHE_CTL_DC_EN 16 // bit 16: data cache enable

// DDR controller (offsets relative to DDR_BASE)
#define DDR_BASE          0x10019000
#define DDR_STATUS        20  // status register
#define DDR_STATUS_READY  1   // bit 1: controller ready
#define DDR_MODE          512
#define DDR_CTRL          516
#define DDR_CTRL_INIT     0   // bit 0: trigger initialization
#define DDR_CTRL_DONE     8   // bit 8: initialization done
#define DDR_TIMING0       520
#define DDR_TIMING1       524
#define DDR_TIMING2       528
#define DDR_TIMING3       532
#define DDR_TIMING4       536
#define DDR_TIMING5       540
#define DDR_TIMING6       544
#define DDR_TIMING7       548
#define DDR_TIMING8       552
#define DDR_TIMING9       556
#define DDR_TIMINGA       560
#define DDR_TIMINGB       564
#define DDR_TIMINGC       568
#define DDR_TIMINGD       572
#define DDR_REG0          608
#define DDR_REG1          624
#define DDR_REG2          628
#define DDR_REG3          632
#define DDR_REG4          680
#define DDR_REG5          684
#define DDR_REG6          816
#define DDR_REG7          824

// QSPI0 controller
#define QSPI0_BASE        0x10014000
#define QSPI0_SCKDIV      100 // serial clock divider register (SCKDIV = 0x130009)

// MSTATUS bits
#define MSTATUS_LOW_MASK  0x7fff // MSTATUS[14:0]: interrupt-enable and FS bits
#define MSTATUS_FS_INITIAL 13    // FS field low bit: 0b01 = Initial (FPU on)

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0

	// disable I/D cache before DDR init
	MOV	$(1<<MCACHE_CTL_IC_EN | 1<<MCACHE_CTL_DC_EN), T0
	WORD	$0x7ca2b073	// csrrc x0, CSR_MCACHE_CTL, t0

	// DDR controller initialization
	MOV	$DDR_BASE, X14

	// clear DDR_STATUS, then wait for the controller-ready bit
	MOVW	X0, DDR_STATUS(X14)
ddr_wait_ready:
	MOVW	DDR_STATUS(X14), X15
	AND	$(1<<DDR_STATUS_READY), X15, X15
	BEQ	X15, X0, ddr_wait_ready

	// DDR_MODE: 134258688 + (-765) = 0x08009d03
	MOV	$134258688, X15
	ADDW	$-765, X15, X15
	MOVW	X15, DDR_MODE(X14)

	// DDR_CTRL: 0  (clear before triggering init)
	MOVW	X0, DDR_CTRL(X14)

	// DDR_TIMING0: 270336 + (-912)
	MOV	$270336, X15
	ADDW	$-912, X15, X15
	MOVW	X15, DDR_TIMING0(X14)

	// DDR_TIMING1: 24
	MOV	$24, X15
	MOVW	X15, DDR_TIMING1(X14)

	// DDR_TIMING2: -2147483648 + 84
	MOV	$-2147483648, X15
	ADDW	$84, X15, X15
	MOVW	X15, DDR_TIMING2(X14)

	// DDR_TIMING3: 520556544 + (-1529)
	MOV	$520556544, X15
	ADDW	$-1529, X15, X15
	MOVW	X15, DDR_TIMING3(X14)

	// DDR_TIMING4: 289738752 + 530
	// Also init X12 = 8192 here (used for TIMING5 and REG2)
	MOV	$289738752, X15
	MOV	$8192, X12
	ADDW	$530, X15, X15
	MOVW	X15, DDR_TIMING4(X14)

	// DDR_TIMING5: X12 + 48 = 8240  -> stored in X13
	ADDW	$48, X12, X13
	MOVW	X13, DDR_TIMING5(X14)

	// DDR_TIMING6: 28672 + (-254)  -> reuse X13
	// X15 is overwritten here with 28672; X15 is later used for DDR_REG6
	MOV	$28672, X15
	ADDW	$-254, X15, X13
	MOVW	X13, DDR_TIMING6(X14)

	// DDR_TIMING7: 40960 + (-1639) -> X13 (saved for DDR_REG3)
	MOV	$40960, X13
	ADDW	$-1639, X13, X13
	MOVW	X13, DDR_TIMING7(X14)

	// DDR_TIMING8: 4096 + (-839)
	MOV	$4096, X11
	ADDW	$-839, X11, X11
	MOVW	X11, DDR_TIMING8(X14)

	// DDR_TIMING9: 0
	MOVW	X0, DDR_TIMING9(X14)

	// DDR_TIMINGA: -1879048192 + 4
	MOV	$-1879048192, X11
	ADDW	$4, X11, X11
	MOVW	X11, DDR_TIMINGA(X14)

	// DDR_TIMINGB: 50528256 + 771
	MOV	$50528256, X11
	ADDW	$771, X11, X11
	MOVW	X11, DDR_TIMINGB(X14)

	// DDR_TIMINGC: same value as TIMINGB
	MOVW	X11, DDR_TIMINGC(X14)

	// DDR_TIMINGD: 262144 + 19
	MOV	$262144, X11
	ADDW	$19, X11, X11
	MOVW	X11, DDR_TIMINGD(X14)

	// DDR_REG0: 8
	MOV	$8, X11
	MOVW	X11, DDR_REG0(X14)

	// DDR_REG1: 255
	MOV	$255, X11
	MOVW	X11, DDR_REG1(X14)

	// DDR_REG2: X12 + 546 = 8192 + 546 = 8738
	ADDW	$546, X12, X12
	MOVW	X12, DDR_REG2(X14)

	// DDR_REG3: X13 (= DDR_TIMING7 value = 39321)
	MOVW	X13, DDR_REG3(X14)

	// DDR_REG4: 16384 + (-384)
	MOV	$16384, X13
	ADDW	$-384, X13, X13
	MOVW	X13, DDR_REG4(X14)

	// DDR_REG5: 401408 + (-1408)
	MOV	$401408, X13
	ADDW	$-1408, X13, X13
	MOVW	X13, DDR_REG5(X14)

	// DDR_REG6: X15 (= 28672) + 1911 = 30583
	// X15 was set to 28672 at DDR_TIMING6 and not modified since
	ADDW	$1911, X15, X15
	MOVW	X15, DDR_REG6(X14)

	// DDR_REG7: 5
	MOV	$5, X15
	MOVW	X15, DDR_REG7(X14)

	// DDR_CTRL: 1 -> trigger DDR initialization
	MOV	$(1<<DDR_CTRL_INIT), X15
	MOVW	X15, DDR_CTRL(X14)

	// trigger DDR init, then wait for the init-done bit
	MOV	$(1<<DDR_CTRL_INIT), X15
	MOVW	X15, DDR_CTRL(X14)
	MOV	$DDR_BASE, X14
ddr_wait_init:
	MOVW	DDR_CTRL(X14), X15
	AND	$(1<<DDR_CTRL_DONE), X15, X15
	BEQ	X15, X0, ddr_wait_init

	// configure the QSPI0 clock divider (SCKDIV = 0x130009)
	MOV	$1245184, X15
	MOV	$QSPI0_BASE, X14
	ADDW	$9, X15, X15
	MOVW	X15, QSPI0_SCKDIV(X14)

	// enable I/D cache
	MOV	$(1<<MCACHE_CTL_IC_EN | 1<<MCACHE_CTL_DC_EN), T0
	WORD	$0x7ca2a073	// csrrs x0, CSR_MCACHE_CTL, t0

	// disable interrupts and enable the FPU
	MOV	$0, T0
	CSRRW	T0, SIE, ZERO
	CSRRW	T0, MIE, ZERO
	// clear MSTATUS[14:0] (interrupt-enable and FS bits) ...
	MOV	$MSTATUS_LOW_MASK, T0
	CSRRC	T0, MSTATUS, ZERO
	// ... then set FS = 0b01 (Initial) to enable the FPU
	MOV	$(1<<MSTATUS_FS_INITIAL), T0
	CSRRS	T0, MSTATUS, ZERO

	// set the stack to the top of DRAM and jump to the runtime entry
	MOV	runtime∕goos·RamStart(SB), X2
	MOV	runtime∕goos·RamSize(SB), T1
	MOV	runtime∕goos·RamStackOffset(SB), T2
	ADD	T1, X2
	SUB	T2, X2

	JMP	_rt0_tamago_start(SB)
