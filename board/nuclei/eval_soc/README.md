TamaGo - bare metal Go - Nuclei EvalSoC support
===============================================

tamago | https://github.com/usbarmory/tamago

Copyright (c) The TamaGo Authors. All Rights Reserved.

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Marvin Drees  
marvin.drees@9elements.com

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal processors.

The [eval_soc](https://github.com/usbarmory/tamago/tree/master/board/nuclei/eval_soc)
package provides support for the Nuclei EvalSoC machine emulated by the
[Nuclei QEMU](https://doc.nucleisys.com/nuclei_tools/qemu/index.html) fork
(`-M nuclei_evalsoc`).

The EvalSoC reuses the [Fisilink FSL91030](https://github.com/usbarmory/tamago/tree/master/soc/fisilink/fsl91030)
SoC package (Nuclei UX600 core) and is closely related to the
[MilkV Vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega)
board: both run the same UX600 core and SiFive-compatible peripherals. This
package adapts the SoC to the emulator, which does not model the GPIO block or
the hardware CLINT timer.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC               | Related board packages                                                         | Peripheral drivers                                                        |
|-------------------|--------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| Fisilink FSL91030 | [milkv/vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega) | [UART, CLINT](https://github.com/usbarmory/tamago/tree/master/soc/sifive) |

> [!WARNING]
> This package targets the Nuclei QEMU emulator only; it is not meant for
> physical hardware (see board/milkv/vega for the MilkV Vega board).

Compilation
===========

The `linknanotime` build tag is required: it excludes the FSL91030 SoC timer
(which reads the hardware CLINT, unmapped in the emulator) so this package can
supply a time source based on the RISC-V `time` CSR.

```bash
# Set up TamaGo compiler
export TAMAGO=/path/to/tamago-go/bin/go

GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -tags linknanotime \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o example \
    main.go
```

Always extract the ELF entry point and use it as the emulator jump target, as
it differs from the `-T` text segment base:

```bash
ENTRY=$(riscv64-linux-gnu-readelf -h example | awk '/Entry point/{print $4}')
```

QEMU emulation
==============

Use the [Nuclei QEMU](https://doc.nucleisys.com/nuclei_tools/qemu/index.html)
fork with the `nuclei_evalsoc` machine:

```bash
# Generate the machine configuration
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
    -M nuclei_evalsoc,download=ddr,soc-cfg=fsl91030.json \
    -cpu nuclei-ux600fd \
    -m 200M -smp 1 -nodefaults -nographic \
    -serial stdio \
    -bios example
```

RAM is limited to 200 MB to avoid overlap with the emulator internal address
decoding. Pass `startaddr` set to the ELF `e_entry`, not the `-T` base address.

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
