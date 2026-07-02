TamaGo - bare metal Go - QEMU LoongArch virt support
=====================================================

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

The `virt` package provides support for the
[QEMU LoongArch virt](https://www.qemu.org/docs/master/system/loongarch/virt.html)
emulated machine, configured with a single core.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[ls3a5000](https://github.com/usbarmory/tamago/tree/master/soc/loongson/ls3a5000).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC               | Board                                                                          | SoC package                                                                        | Board package                                                                |
|-------------------|--------------------------------------------------------------------------------|------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| Loongson 3A5000   | [QEMU virt](https://www.qemu.org/docs/master/system/loongarch/virt.html)       | [ls3a5000](https://github.com/usbarmory/tamago/tree/master/soc/loongson/ls3a5000)  | [qemu/virt](https://github.com/usbarmory/tamago/tree/master/board/qemu/virt) |

Compiling
=========

Go applications are required to set `GOOSPKG` to the desired
[runtime/goos](https://github.com/usbarmory/tamago-go/tree/latest/src/runtime/goos)
overlay and import the relevant board package to ensure that hardware
initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/qemu/virt"
)
```

Go applications can be compiled as usual, using the compiler built for the
TamaGo framework, but with the addition of the following flags/variables:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=loong64 \
	${TAMAGO} build -ldflags "-T 0x1000000 -R 0x1000" main.go
```

The QEMU `virt` machine reserves the low 2 MiB of RAM for boot information and
the device tree, therefore the text segment must be linked above that region
(the example above links at 16 MiB).

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing and debugging
=======================

The target can be executed under emulation as follows:

```sh
qemu-system-loongarch64 \
	-machine virt -m 256M \
	-nographic -monitor none -serial stdio -net none \
	-kernel main
```

Unlike the riscv64 `qemu` targets no BIOS is required, as the QEMU LoongArch
`virt` machine loads the ELF image and jumps to its entry point directly.

The emulated target can be debugged with GDB by adding the `-S -s` flags to the
previous execution command, this will make qemu wait for a GDB connection that
can be launched as follows:

```sh
loongarch64-unknown-linux-gnu-gdb -ex "target remote 127.0.0.1:1234" main
```

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
