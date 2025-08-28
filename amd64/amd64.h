// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#define MSR_EFER 0xc0000080

#define PML4T 0x9000	// Page Map Level 4 Table       (512GB entries)
#define PDPT  0xa000	// Page Directory Pointer Table   (1GB entries)
#define PDT   0xb000	// Page Directory Table           (2MB entries)
#define PT    0xc000	// Page Table                     (4kB entries)

// These legacy prefixes are used in 16-bit Real Mode to ensure valid Go
// assembly interpretation.

#define CSADDR BYTE $0x2e	//          CS segment override prefix
#define DATA32 BYTE $0x66	// 32-bit operand size override prefix
#define ADDR32 BYTE $0x67	// 32-bit address size override prefix
