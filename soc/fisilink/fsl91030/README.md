TamaGo - bare metal Go - Fisilink FSL91030 support
===================================================

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

The [fsl91030](https://github.com/usbarmory/tamago/tree/master/soc/fisilink/fsl91030)
package provides support for the Fisilink FSL91030 SoC, a RV64IMAFDC processor
based on the Nuclei UX600 core (sv39 MMU, 400 MHz).

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC               | Related board packages                                                                                                                                                   | Peripheral drivers                                                        |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| Fisilink FSL91030 | [milkv/vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega), [nuclei/eval_soc](https://github.com/usbarmory/tamago/tree/master/board/nuclei/eval_soc) | [CLINT, UART](https://github.com/usbarmory/tamago/tree/master/soc/sifive) |

> [!WARNING]
> This package is in early development stages and its API might still change.

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkcpuinit`: include the `cpuinit` boot vector (`boot_riscv64.s`) which
  initializes DDR, cache and the FPU before the runtime starts
* `linknanotime`: exclude the default CLINT-based `nanotime`, allowing a board
  package to supply its own time source

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
