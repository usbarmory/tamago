TamaGo - bare metal Go - RISC-V 64-bit support
==============================================

tamago | https://github.com/usbarmory/tamago  

Copyright (c) WithSecure Corporation  

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
applications on bare metal AMD64/ARM/RISC-V processors.

The [riscv64](https://github.com/usbarmory/tamago/tree/master/riscv64) package
provides support for RISC-v 64-bit CPUs.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU      | Related platform packages                                                       | Core drivers |
|----------|---------------------------------------------------------------------------------|--------------|
| RV64IMAC | [sifive_u](https://github.com/usbarmory/tamago/blob/master/board/qemu/sifive_u) | PMP          |

Build tags
==========

The following build tags allow application to override the package own definition of
[external functions required by the runtime](https://github.com/usbarmory/tamago/wiki/Internals#go-runtime-changes):

* `linkramstart`: exclude `ramStart` from `mem.go`
* `linkcpuinit`: exclude `cpuinit` from `init.s`

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
