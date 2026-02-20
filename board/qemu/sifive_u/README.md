TamaGo - bare metal Go - QEMU sifive_u support
==============================================

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

The [sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u)
package provides support for the [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html)
emulated machine configured with a single U54 core.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

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

Go distribution supporting `GOOS=tamago`
---------------------------------------

The [tamago](https://github.com/usbarmory/tamago/tree/latest/cmd/tamago)
command downloads, compiles, and runs the `go` command from the
[TamaGo distribution](https://github.com/usbarmory/tamago-go) matching the
tamago module version from the application `go.mod`.

Applications can add `github.com/usbarmory/tamago` to `go.mod`, and then
replace the `go` command with:


```sh
go run github.com/usbarmory/tamago/cmd/tamago
```

or add the following line to `go.mod` go use `go tool tamago` as go command:

```
tool github.com/usbarmory/tamago/cmd/tamago
```

Alternatively the
[latest TamaGo distribution](https://github.com/usbarmory/tamago-go/tree/latest) can be
manually built or the
[latest binary release](https://github.com/usbarmory/tamago-go/releases/latest) can be used:

```sh
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Building applications
---------------------

Go applications are required to set `GOOSPKG` to the desired
[runtime/goos](https://github.com/usbarmory/tamago-go/tree/latest/src/runtime/goos)
overlay and import the relevant board package to ensure that hardware
initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/qemu/sifive_u"
)
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=riscv64 \
	${TAMAGO} build -ldflags "-T 0x80010000 -R 0x1000" main.go
```

An example application, targeting the QEMU sifive_u platform,
is [available](https://github.com/usbarmory/tamago-example).

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing and debugging
=======================

The [example application](https://github.com/usbarmory/tamago-example) provides
reference usage and a Makefile target for automatic creation of an ELF image as
well as emulated execution.

QEMU
----

The target can be executed under emulation as follows:

```sh
dtc -I dts -O dtb qemu-riscv64-sifive_u.dts -o qemu-riscv54-sifive_u.dtb

qemu-system-riscv64 \
	-machine sifive_u -m 512M \
	-nographic -monitor none -serial stdio -net none \
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

```sh
riscv64-elf-gdb -ex "target remote 127.0.0.1:1234" main
```

Breakpoints can be set in the usual way:

```
b ecdsa.Verify
continue
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
