TamaGo - bare metal Go - MilkV Vega support
===========================================

tamago | https://github.com/usbarmory/tamago

Copyright (c) The TamaGo Authors. All Rights Reserved.

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani
andrea@inversepath.com

Andrej Rosano
andrej@inversepath.com

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal processors.

The [vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega)
package provides support for the [MilkV Vega](https://milkv.io/vega) board,
a RISC-V network switch board powered by the Fisilink FSL91030 SoC.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Board Overview
==============

| Peripheral | Details                                  |
|------------|------------------------------------------|
| SoC        | Fisilink FSL91030 (Nuclei UX600 RV64IMAFDC @ 400 MHz) |
| DRAM       | 240 MB at 0x41000000                     |
| NOR Flash  | 64 MB XIP at 0x20000000 (Infineon S25HL512T) |
| UART0      | Console, SiFive-compatible at 0x10013000 |
| UART1      | Secondary, SiFive-compatible at 0x10023000 |
| CLINT      | Timer at 0x2001000, 32768 Hz timebase    |

Supported hardware
==================

| Board    | SoC        | RAM    | Flash  | Status       |
|----------|------------|--------|--------|--------------|
| MilkV Vega | FSL91030 | 240 MB | 64 MB  | Experimental |

Compilation
===========

```bash
# Set up TamaGo compiler
export TAMAGO=/path/to/tamago-go/bin/go

# Build for MilkV Vega (hardware)
GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o example \
    main.go

# Build for QEMU (200 MB RAM, QEMU-safe memory map)
GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -tags qemu \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o example \
    main.go
```

The ELF `e_entry` (entry point of `_rt0_tamago_start`) is not equal to the
`-T` text segment base. Always extract `e_entry` from the ELF and use that as
the CPU jump target:

```bash
riscv64-linux-gnu-readelf -h example | grep "Entry point"
```

### Linker Flags

* `-T 0x41010000`: Text segment at DRAM base + 64 KB
* `-R 0x1000`: Read-only segment alignment (4 KB)

QEMU Testing
============

Use the Nuclei QEMU (nuclei/9.0 fork) with the `nuclei_evalsoc` machine and
the `qemu` build tag:

```bash
# Extract actual entry point
ENTRY=$(riscv64-linux-gnu-readelf -h example | awk '/Entry point/{print $4}')

# Generate machine configuration
cat > fsl91030.json << EOF
{
    "general_config": {
        "ddr":      { "base": "0x41000000", "size": "200M" },
        "norflash": { "base": "0x20000000", "size": "64M" },
        "uart0":    { "base": "0x10013000", "irq": "33" },
        "uart1":    { "base": "0x10023000", "irq": "34" },
        "iregion":  { "base": "0x4000000" },
        "cpu_freq": "400000000",
        "timer_freq": "32768",
        "irqmax": "64"
    },
    "download": { "ddr": { "startaddr": "$ENTRY" } }
}
EOF

# Run
qemu-system-riscv64 \
    -M nuclei_evalsoc,download=ddr,soc-cfg=/abs/path/fsl91030.json \
    -cpu nuclei-ux600fd \
    -m 200M -smp 1 -nodefaults -nographic \
    -serial stdio \
    -bios example
```

Use `-bios` (not `-kernel`) to select PLIC+CLINT interrupt mode. RAM is
limited to 200 MB with the `qemu` build tag to avoid overlap with the PLIC
region. Pass `startaddr` set to the ELF `e_entry`, not the `-T` base address.

Build Tags
==========

The following build tags are supported:

* `linkramsize`: Exclude `ramSize`; provide your own via `go:linkname`
* `linkprintk`: Exclude `printk`; provide your own via `go:linkname`
* `qemu`: Use QEMU-safe memory map (200 MB ramSize, avoids PLIC region overlap)

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
