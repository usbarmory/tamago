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

// Register number for T0 = X5 (used as scratch in CSR operations)
#define t0_reg 5

// MRET: return from machine-mode trap (opcode 0x30200073)
#define MRET WORD $0x30200073

// func irq_enable()
// Enables machine-mode external interrupts:
//   mie.MEIE   (bit 11) = 1  — enables external interrupt from PLIC
//   mstatus.MIE (bit 3) = 1  — globally enables machine-mode interrupts
TEXT ·irq_enable(SB),NOSPLIT,$0-0
	MOV	$MIE_MEIE, T0
	CSRS(t0_reg, mie_csr)     // mie |= MEIE
	MOV	$MSTATUS_MIE, T0
	CSRS(t0_reg, mstatus_csr) // mstatus.MIE = 1
	RET

// func irq_disable()
// Disables machine-mode external interrupts by clearing mie.MEIE.
// Does not affect the global mstatus.MIE or timer/software interrupts.
TEXT ·irq_disable(SB),NOSPLIT,$0-0
	MOV	$MIE_MEIE, T0
	CSRC(t0_reg, mie_csr)     // mie &= ~MEIE
	RET

// func wfi()
// Executes a single WFI (Wait For Interrupt) instruction and returns.
// The CPU is suspended until the next interrupt is received.
TEXT ·wfi(SB),NOSPLIT,$0-0
	WORD	$0x10500073 // wfi
	RET

// trapHandler is the unified machine-mode trap handler installed at mtvec by
// riscv64.CPU.ServiceInterrupts. It handles both synchronous exceptions and
// asynchronous interrupts from a single direct-mode entry point.
//
// Synchronous exceptions (mcause bit 63 = 0): restores registers and
// tail-calls DefaultExceptionHandler, which panics.
// Asynchronous machine external interrupt (cause code 11): calls
// runtime.WakeG to wake the IRQ handler goroutine, then returns via MRET.
// Other asynchronous interrupts (timer, software): returns via MRET unchanged.
//
// Saves and restores all caller-saved integer registers (RA, T0-T6, A0-A7)
// in a 144-byte 16-byte-aligned frame. WakeG on RISC-V receives the G
// pointer in T0 rather than A0.
//
//go:nosplit
TEXT ·trapHandler(SB),NOSPLIT|NOFRAME,$0-0
	// Save RA at the top of the incoming frame, then allocate the frame.
	// After SUB, 0(SP) holds the saved RA.
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

	// Read mcause CSR into T0 (X5).
	CSRR(mcause_csr, t0_reg)

	// Check bit 63 (interrupt flag). BLT treats registers as signed 64-bit
	// integers, so a negative T0 means bit 63 is set → asynchronous interrupt.
	BLT	T0, ZERO, interrupt_path

exception_path:
	// Synchronous exception. Restore all caller-saved registers and SP,
	// then tail-call DefaultExceptionHandler. That function re-reads mcause
	// from the CSR and panics — it never returns, so no MRET is needed.
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
	// Asynchronous interrupt. Mask T0 to the 6-bit cause code (bits 5:0).
	// Machine external interrupt = code 11 (M-mode PLIC delivery).
	AND	$63, T0, T0
	MOV	$11, T1
	BNE	T0, T1, done   // not M-mode external interrupt → skip WakeG

	// Machine external interrupt. Wake the IRQ handler goroutine.
	// WakeG on RISC-V requires the G pointer in T0 (not A0).
	MOV	·irqHandlerG(SB), T0
	BEQ	T0, ZERO, done  // irqHandlerG not yet set → skip
	CALL	runtime·WakeG(SB)

done:
	// Restore all caller-saved registers (including RA clobbered by CALL above).
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
	// MRET restores mstatus.MIE from mstatus.MPIE (re-enables interrupts
	// if they were enabled before the trap), and returns to mepc.
	MRET
