TamaGo - bare metal Go for RISC-V SoCs - QEMU sifive_u support
==============================================================

tamago | https://github.com/usbarmory/tamago  

Copyright (c) WithSecure Corporation  
https://foundry.withsecure.com

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@withsecure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@withsecure.com   | andrej@inversepath.com  

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM/RISC-V System-on-Chip (SoC) components.

The [sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u)
package provides support for the [QEMU sifive_u](https://qemu.readthedocs.io/en/latest/system/riscv/sifive_u.html)
emulated machine configured with a single U54 core.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC          | Board                                                                        | SoC package                                                               | Board package                                                                        |
|--------------|------------------------------------------------------------------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| SiFive FU540 | [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html) | [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u) |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/qemu/sifive_u"
)
```

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables and
ensuring that the required SoC and board packages are available in `GOPATH`:

```
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARCH=riscv64 \
  ${TAMAGO} build -ldflags "-T 0x80010000 -E _rt0_riscv64_tamago -R 0x1000"
```

An example application, targeting the QEMU sifive_u platform,
is [available](https://github.com/usbarmory/tamago-example).

Executing and debugging
=======================

The [example application](https://github.com/usbarmory/tamago-example) provides
reference usage and a Makefile target for automatic creation of an ELF image as
well as emulated execution.

QEMU
----

The target can be executed under emulation as follows:

```
dtc -I dts -O dtb qemu-riscv64-sifive_u.dts -o qemu-riscv54-sifive_u.dtb

qemu-system-riscv64 \
	-machine sifive_u -m 512M \
	-nographic -monitor none -serial stdio -net none
	-dtb qemu-riscv64-sifive_u.dtb \
	-bios bios.bin
```

At this time a bios is required to jump to the correct entry point of the ELF
image, the [example application](https://github.com/usbarmory/tamago-example)
includes a minimal bios which is configured and compiled for all riscv64 `qemu`
targets.

The emulated target can be debugged with GDB by adding the `-S -s` flags to the
previous execution command, this will make qemu waiting for a GDB connection
that can be launched as follows:

```
riscv64-elf-gdb -ex "target remote 127.0.0.1:1234" example
```

Breakpoints can be set in the usual way:

```
b ecdsa.Verify
continue
```

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
