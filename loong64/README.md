TamaGo - bare metal Go - LoongArch 64-bit support
=================================================

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

The [loong64](https://github.com/usbarmory/tamago/tree/master/loong64) package
provides support for LoongArch 64-bit CPUs.

The CPU runs in Direct Address translation mode (`CRMD.DA`), therefore no page
tables are required; the Go LoongArch assembler exposes no privileged CSR, TLB,
`ertn` or `idle` mnemonics, so these are emitted as hand-encoded `WORD`
directives (verified with `go tool objdump`).

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU     | Related platform packages                                                 | Core drivers                |
|---------|---------------------------------------------------------------------------|-----------------------------|
| la464   | [qemu/virt](https://github.com/usbarmory/tamago/blob/master/board/qemu/virt) | stable timer, CPUCFG, exceptions |

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkcpuinit`: exclude `cpuinit` from `init.s`

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
