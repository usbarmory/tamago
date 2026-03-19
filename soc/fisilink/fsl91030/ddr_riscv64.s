// Fisilink FSL91030 DDR controller initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func initDDR()
//
// DDR initialization is performed by the flashboot assembly stub
// (tools/flashboot.s) before the TamaGo runtime starts. By the time
// Go code runs, DDR is already operational.
//
// This Go-callable function provides the same DDR init sequence for
// use after the Go runtime is running (e.g., DDR reconfiguration or
// builds that do not use flashboot.s). The implementation is the same
// register sequence used by flashboot.s and boot_riscv64.s (linkcpuinit),
// derived from the vendor's freeloader.S.
//
// Callers are responsible for calling DisableCache() before and
// EnableCache() after this function when called at runtime.
//
// Register usage (all caller-saved; no stack frame needed):
//   X14 (A4): DDR base (0x10001000), constant throughout
//   X15 (A5): working register, DDR config values
//   X12 (A2): 8192 = 0x2000, DDR_TIMING5 base and REG2 accumulator
//   X13 (A3): derived timing value, saved for DDR_REG3
//   X11 (A1): miscellaneous timing/config values
TEXT ·initDDR(SB),NOSPLIT|NOFRAME,$0

	// Set DDR base address: 0x10001000
	MOV	$0x10001000, X14

	// Clear DDR_STATUS (offset 20) before polling
	MOVW	X0, 20(X14)

	// Wait for DDR_STATUS bit 1 (controller ready)
ddr_wait_ready:
	MOVW	20(X14), X15
	AND	$2, X15, X15
	BEQ	X15, X0, ddr_wait_ready

	// DDR_MODE (offset 512 = 0x200): 134258688 + (-765) = 0x08009D03
	MOV	$134258688, X15
	ADDW	$-765, X15, X15
	MOVW	X15, 512(X14)

	// DDR_CTRL (offset 516 = 0x204): 0 (clear before triggering init)
	MOVW	X0, 516(X14)

	// DDR_TIMING0 (offset 520 = 0x208): 270336 + (-912) = 0x041C70
	MOV	$270336, X15
	ADDW	$-912, X15, X15
	MOVW	X15, 520(X14)

	// DDR_TIMING1 (offset 524 = 0x20C): 24
	MOV	$24, X15
	MOVW	X15, 524(X14)

	// DDR_TIMING2 (offset 528 = 0x210): -2147483648 + 84 = 0x80000054
	MOV	$-2147483648, X15
	ADDW	$84, X15, X15
	MOVW	X15, 528(X14)

	// DDR_TIMING3 (offset 532 = 0x214): 520556544 + (-1529) = 0x1F070A07
	MOV	$520556544, X15
	ADDW	$-1529, X15, X15
	MOVW	X15, 532(X14)

	// DDR_TIMING4 (offset 536 = 0x218): 289738752 + 530 = 0x11451212
	// Also initialise X12 = 8192 (reused for TIMING5 and REG2)
	MOV	$289738752, X15
	MOV	$8192, X12
	ADDW	$530, X15, X15
	MOVW	X15, 536(X14)

	// DDR_TIMING5 (offset 540 = 0x21C): X12 + 48 = 8240 → saved in X13
	ADDW	$48, X12, X13
	MOVW	X13, 540(X14)

	// DDR_TIMING6 (offset 544 = 0x220): 28672 + (-254) = 0x6F02
	// X15 retains 28672 after this instruction, used later for DDR_REG6
	MOV	$28672, X15
	ADDW	$-254, X15, X13
	MOVW	X13, 544(X14)

	// DDR_TIMING7 (offset 548 = 0x224): 40960 + (-1639) = 0x9999
	// X13 is saved for DDR_REG3
	MOV	$40960, X13
	ADDW	$-1639, X13, X13
	MOVW	X13, 548(X14)

	// DDR_TIMING8 (offset 552 = 0x228): 4096 + (-839) = 0xCB9
	MOV	$4096, X11
	ADDW	$-839, X11, X11
	MOVW	X11, 552(X14)

	// DDR_TIMING9 (offset 556 = 0x22C): 0
	MOVW	X0, 556(X14)

	// DDR_TIMINGA (offset 560 = 0x230): -1879048192 + 4 = 0x90000004
	MOV	$-1879048192, X11
	ADDW	$4, X11, X11
	MOVW	X11, 560(X14)

	// DDR_TIMINGB (offset 564 = 0x234): 50528256 + 771 = 0x03030303
	MOV	$50528256, X11
	ADDW	$771, X11, X11
	MOVW	X11, 564(X14)

	// DDR_TIMINGC (offset 568 = 0x238): same as TIMINGB
	MOVW	X11, 568(X14)

	// DDR_TIMINGD (offset 572 = 0x23C): 262144 + 19 = 0x40013
	MOV	$262144, X11
	ADDW	$19, X11, X11
	MOVW	X11, 572(X14)

	// DDR_REG0 (offset 608 = 0x260): 8
	MOV	$8, X11
	MOVW	X11, 608(X14)

	// DDR_REG1 (offset 624 = 0x270): 255
	MOV	$255, X11
	MOVW	X11, 624(X14)

	// DDR_REG2 (offset 628 = 0x274): X12 + 546 = 8192 + 546 = 8738
	ADDW	$546, X12, X12
	MOVW	X12, 628(X14)

	// DDR_REG3 (offset 632 = 0x278): X13 (= DDR_TIMING7 value = 0x9999)
	MOVW	X13, 632(X14)

	// DDR_REG4 (offset 680 = 0x2A8): 16384 + (-384) = 16000 = 0x3E80
	MOV	$16384, X13
	ADDW	$-384, X13, X13
	MOVW	X13, 680(X14)

	// DDR_REG5 (offset 684 = 0x2AC): 401408 + (-1408) = 400000 = 0x61A80
	MOV	$401408, X13
	ADDW	$-1408, X13, X13
	MOVW	X13, 684(X14)

	// DDR_REG6 (offset 816 = 0x330): X15 (= 28672) + 1911 = 30583 = 0x7777
	// X15 was set to 28672 at DDR_TIMING6 and not modified since
	ADDW	$1911, X15, X15
	MOVW	X15, 816(X14)

	// DDR_REG7 (offset 824 = 0x338): 5
	MOV	$5, X15
	MOVW	X15, 824(X14)

	// Trigger DDR initialization: DDR_CTRL = 1
	MOV	$1, X15
	MOVW	X15, 516(X14)

	// Wait for DDR_CTRL bit 8 (init done)
ddr_wait_init:
	MOVW	516(X14), X15
	AND	$256, X15, X15
	BEQ	X15, X0, ddr_wait_init

	RET
