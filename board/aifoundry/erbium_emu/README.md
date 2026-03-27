TamaGo - bare metal Go - AI Foundry erbium_emu support
======================================================

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

The [erbium_emu](https://github.com/usbarmory/tamago/tree/master/board/aifoundry/erbium_emu)
package provides support for AI Foundry [Erbium](https://github.com/aifoundry-org/erbium) processor, running on the
[erbium_emu emulator](https://github.com/aifoundry-org/et-platform) as a single ET-Minion (rv64imfc) core.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[erbium](https://github.com/usbarmory/tamago/tree/master/soc/aifoundry/erbium).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC               | Board                                                      | SoC package                                                                    | Board package                                                                            |
|-------------------|------------------------------------------------------------|--------------------------------------------------------------------------------|------------------------------------------------------------------------------------------|
| AI Foundry Erbium | [erbium_emu](https://github.com/aifoundry-org/et-platform) | [erbium](https://github.com/usbarmory/tamago/tree/master/soc/aifoundry/erbium) | [erbium_emu](https://github.com/usbarmory/tamago/tree/master/board/aifoundry/erbium_emu) |

Compiling
=========

Go distribution supporting `GOOS=tamago GOSOFT=1`
------------------------------------------------

This target requires a specific `GOOS=tamago` compiler branch to support the
following:

  * `GOSOFT=1`: compiler build time variable to enable soft float for `riscv64`, removing
    requirement for `ad` extensions and forcing single-threaded operation.

  * `tiny`: build tag to support considerable reduction of RAM allocation requirements.

The [kotama repository](https://github.com/usbarmory/kotama) provides
instructions and a reference implementation for this target.

Building applications
---------------------

Go applications are required to set `GOOSPKG` to the desired
[runtime/goos](https://github.com/usbarmory/tamago-go/tree/latest/src/runtime/goos)
overlay and import the relevant board package to ensure that hardware
initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/aifoundry/erbium_emu"
)
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=riscv64 GOSOFT=1 \
	${TAMAGO} build -ldflags "-T 0x40010000 -R 0x1000" main.go
```

An example application, targeting AI Foundry erbium_emu,
is [available](https://github.com/usbarmory/kotama).

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`
* `linkcpuinit`: override `cpuinit` imported from `riscv64/init.s` to park
                 additional harts (required on multi-hart instances).
* `tiny`: reduce heap allocation requirements

Executing and debugging
=======================

The [kotama repository](https://github.com/usbarmory/kotama) provides reference
usage and a script for automatic creation of an ELF image for emulated
execution.

erbium_emu
----------

The target can be executed under emulation as follows:

```sh
/opt/et/bin/erbium_emu \
	-reset_pc $(nm main|grep _rt0_riscv64_tamago | cut -d' ' -f1) \
	-max_cycles -1 -single_thread \
	-elf_load main
```
The emulated target can be debugged with GDB by adding the `-gdb` flag to the
previous execution command, the emulator will wait for a GDB connection that
can be launched as follows:

```sh
riscv64-elf-gdb -ex "target remote 127.0.0.1:1337" main
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
