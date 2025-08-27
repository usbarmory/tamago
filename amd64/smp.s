// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "amd64.h"
#include "go_asm.h"
#include "textflag.h"

// NOTE: this offset needs adjustment in case of any changes to ·apinit
#define doneOffset 0x68
#define doneMarker 0xcccccccccccccccc

// func apinit_reloc(init uintptr, start uintptr)
TEXT ·apinit_reloc(SB),$0-16
	MOVQ	$·apinit<>(SB), SI
	MOVL	init+0(FP), DI

	// end of function marker
	MOVQ	$doneMarker, BX
copy_8:
	MOVQ	(SI), AX
	ADDQ	$8, SI

	CMPQ	AX, BX
	JE	done

	MOVQ	AX, (DI)
	ADDQ	$8, DI

	JMP	copy_8
done:
	MOVQ	start+8(FP), DI
	MOVQ	$·apstart<>(SB), SI
	MOVQ	SI, (DI)

	RET

TEXT ·apinit<>(SB),NOSPLIT|NOFRAME,$0
	// disable interrupts
	CLI

	// This function runs in 16-Bit Real Mode, for this reason Go Assembly
	// must be treated differently.
	//
	// The DATA32 and ADDR32 macros (see amd64.h) are used to ensure
	// correct interpretation of 32-bit operands and/or addresses.
	//
	// The function is copied by `apinit_reloc` to a 16-bit address to be
	// called, for this reason RIP/EIP-relative addressing must be avoided.

	// we might not have a valid stack pointer for CALLs
	DATA32
	MOVL	$PML4T, SP

	// apply BSP paging setup (see init.s)
	MOVL	SP, CR3

	// adjust data segment
	BYTE	$0x8c; BYTE	$0xc8		// mov %cs,%ax
	BYTE	$0x8e; BYTE	$0xd8		// mov %eax,%ds

	// transition from 16-bit Real Mode to 64-bit Long Mode

	MOVL	CR4, AX
	DATA32
	MOVL	$(1<<7 | 1<<5), AX		// set CR4.(PGE|PAE)
	MOVL	AX, CR4

	DATA32
	MOVL	$MSR_EFER, CX
	RDMSR
	DATA32
	ORL	$(1<<8), AX			// set MSR_EFER.LME
	WRMSR

	MOVL	CR0, AX
	DATA32
	ORL	$(1<<31 | 1<<1 |1<<0), AX	// set CR0.(PG|MP|PE)
	MOVL	AX, CR0

	// set Global Descriptor Table
	DATA32
	MOVL	$(const_gdtrAddress), AX
	DATA32
	SUBL	$(const_apinitAddress), AX	// convert linear address to CS offset
	DATA32; ADDR32; CSADDR
	LGDT	(AX)

	// segment selector for GDT entry 2
	DATA32
	MOVL	$0x10, AX

	DATA32; BYTE	$0x8e; BYTE	$0xc0	// mov %eax,%es
	DATA32; BYTE	$0x8e; BYTE	$0xd0	// mov %eax,%ss
	DATA32; BYTE	$0x8e; BYTE	$0xd8	// mov %eax,%ds
	DATA32; BYTE	$0x8e; BYTE	$0xe0	// mov %eax,%fs
	DATA32; BYTE	$0x8e; BYTE	$0xe8	// mov %eax,%gs

	// set far return target (avoiding RIP/EIP relative addressing)
	DATA32
	MOVL	$(const_apinitAddress+doneOffset), AX

	// jump to target in Long Mode
	PUSHQ	$0x08
	PUSHQ	AX
	RETFQ
done:
	// 64-bit Long Mode

	// update GDT limits
	MOVQ	$(const_gdtAddress), AX
	ADDQ	$0x08, AX
	MOVW	$0x0000, (AX)			// code descriptor limit
	ADDQ	$0x08, AX
	MOVW	$0x0000, (AX)			// data descriptor limit

	// reload GDT
	SUBQ	$0x10, AX
	LGDT	(AX)

	// call ·apstart (avoiding RIP/EIP relative addressing)
	MOVQ	$(const_apstartAddress), AX
	CALL	(AX)
marker:
	WORD	$doneMarker

TEXT ·apstart<>(SB),NOSPLIT|NOFRAME,$0
	CALL	sse_enable(SB)

	// apply BSP GDT
	MOVQ	$gdtptr(SB), AX
	LGDT	(AX)

	// apply BSP IDT
	MOVQ	$idtptr(SB), AX
	LIDT	(AX)

	// restore GDT limits for next ·apinit
	MOVQ	$(const_gdtAddress), AX
	ADDQ	$0x08, AX
	MOVW	$0xffff, (AX)			// code descriptor limit
	ADDQ	$0x08, AX
	MOVW	$0xffff, (AX)			// data descriptor limit

	// use taskAddress as counting semaphore for SMP enabling
	MOVQ	$(const_taskAddress), BX
	MOVL	$1, AX
	LOCK
	XADDL	AX, 0(BX)
wait:
	// go to idle state
	HLT

	MOVQ	$(const_taskAddress), AX
	MOVQ	task_pc(AX), R12
	CMPQ	R12, $0
	JE	wait

	MOVQ	task_sp(AX), SP
	MOVQ	task_mp(AX), R13
	MOVQ	task_gp(AX), g

	// clear task
	MOVQ	$0, task_sp(AX)
	MOVQ	$0, task_mp(AX)
	MOVQ	$0, task_gp(AX)
	MOVQ	$0, task_pc(AX)

	MOVQ	g, DI
	CALL	runtime·settls(SB)
	MOVQ	g, (TLS)

	// enable LAPIC and interrupts to support Run()
	MOVL    $(const_LAPIC_BASE+0xf0), AX
	MOVL    $(1<<8), (AX)			// set SVR_ENABLE
	STI

	// call task target
	CALL	R12

	// go back to idle state in case we return
	JMP wait
