// RISC-V 64-bit interrupt support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"
#include "csr.h"

// CSR addresses
#define mcause_csr  0x342
#define mstatus_csr 0x300
#define mie_csr     0x304

// Bit masks
#define MSTATUS_MIE 8     // mstatus bit 3: global machine-mode interrupt enable
#define MIE_MEIE    2048  // mie bit 11: machine external interrupt enable

// MRET: return from machine-mode trap (opcode 0x30200073)
#define MRET WORD $0x30200073

// func irq_enable()
// Enables machine-mode external interrupts:
//   mie.MEIE   (bit 11) = 1  — enables external interrupt from PLIC
//   mstatus.MIE (bit 3) = 1  — globally enables machine-mode interrupts
TEXT ·irq_enable(SB),NOSPLIT,$0-0
	MOV	$MIE_MEIE, T0
	CSRS(t0, mie_csr)     // mie |= MEIE
	MOV	$MSTATUS_MIE, T0
	CSRS(t0, mstatus_csr) // mstatus.MIE = 1
	RET

// func irq_disable()
// Disables machine-mode external interrupts by clearing mie.MEIE.
TEXT ·irq_disable(SB),NOSPLIT,$0-0
	MOV	$MIE_MEIE, T0
	CSRC(t0, mie_csr)     // mie &= ~MEIE
	RET

// func wfi()
TEXT ·wfi(SB),NOSPLIT,$0-0
	WORD	$0x10500073 // wfi
	RET

// trapHandler is the unified machine-mode trap handler installed at mtvec.
//
// Synchronous exceptions (mcause bit 63 = 0): restores registers and
// tail-calls DefaultExceptionHandler (panics).
// Machine external interrupt (cause 11): calls os/signal.Relay to wake
// the goroutine waiting in ServiceInterrupts, then returns via MRET.
// Other asynchronous interrupts: returns via MRET unchanged.
//
// Saves and restores all caller-saved integer registers (RA, T0-T6, A0-A7)
// in a 144-byte frame. For TEE/world-switch contexts additional callee-saved
// registers would need to be saved (see GoTEE monitor/exec_riscv64.s).
//
//go:nosplit
TEXT ·trapHandler(SB),NOSPLIT|NOFRAME,$0-0
	// Save RA and all caller-saved registers.
	MOV	X1,  -144(X2)
	SUB	$144, X2
	MOV	X5,  8(X2)
	MOV	X6,  16(X2)
	MOV	X7,  24(X2)
	MOV	X28, 32(X2)
	MOV	X29, 40(X2)
	MOV	X30, 48(X2)
	MOV	X31, 56(X2)
	MOV	X10, 64(X2)
	MOV	X11, 72(X2)
	MOV	X12, 80(X2)
	MOV	X13, 88(X2)
	MOV	X14, 96(X2)
	MOV	X15, 104(X2)
	MOV	X16, 112(X2)
	MOV	X17, 120(X2)

	// Read mcause. BLT treats it as signed: bit 63 set = asynchronous interrupt.
	CSRR(mcause_csr, t0)
	BLT	T0, ZERO, interrupt_path

exception_path:
	// Restore all registers and tail-call DefaultExceptionHandler (no return).
	MOV	0(X2),   X1
	MOV	8(X2),   X5
	MOV	16(X2),  X6
	MOV	24(X2),  X7
	MOV	32(X2),  X28
	MOV	40(X2),  X29
	MOV	48(X2),  X30
	MOV	56(X2),  X31
	MOV	64(X2),  X10
	MOV	72(X2),  X11
	MOV	80(X2),  X12
	MOV	88(X2),  X13
	MOV	96(X2),  X14
	MOV	104(X2), X15
	MOV	112(X2), X16
	MOV	120(X2), X17
	ADD	$144, X2
	JMP	·DefaultExceptionHandler(SB)

interrupt_path:
	// Only signal for machine external interrupt (cause code 11 = PLIC).
	AND	$63, T0, T0
	MOV	$11, T1
	BNE	T0, T1, done

	// Deliver irqSignal to the goroutine waiting in ServiceInterrupts.
	// Allocate a mini-frame matching arm64's 17*16 for the Relay call, then
	// pass irqSignal as the first argument (X10 = A0 in RISC-V register ABI).
	SUB	$272, X2
	MOV	·irqSignal(SB), X10
	MOV	X10, 8(X2)
	CALL	os∕signal·Relay(SB)
	ADD	$272, X2

done:
	// Restore all caller-saved registers.
	MOV	0(X2),   X1
	MOV	8(X2),   X5
	MOV	16(X2),  X6
	MOV	24(X2),  X7
	MOV	32(X2),  X28
	MOV	40(X2),  X29
	MOV	48(X2),  X30
	MOV	56(X2),  X31
	MOV	64(X2),  X10
	MOV	72(X2),  X11
	MOV	80(X2),  X12
	MOV	88(X2),  X13
	MOV	96(X2),  X14
	MOV	104(X2), X15
	MOV	112(X2), X16
	MOV	120(X2), X17
	ADD	$144, X2

	// Return from machine-mode trap.
	MRET
