TamaGo - bare metal Go - SpacemiT K1 support
===================================================

tamago | https://github.com/usbarmory/tamago

Copyright (c) The TamaGo Authors. All Rights Reserved.

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Michal Gorlas
michal.gorlas@9elements.com

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal processors.

The [k1](https://github.com/usbarmory/tamago/tree/master/soc/spacemit/k1)
package provides support for the SpacemiT Key Stone® K1 SoC, a RV64GCVB processor with eight SpacemiT X60 cores.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC         | Related board                                                                    | Peripheral drivers                                                                                                                                   |
|-------------|------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| SpacemiT K1 | [bananapi/f3](https://github.com/usbarmory/tamago/tree/master/board/bananapi/f3) | [CLINT](https://github.com/usbarmory/tamago/tree/master/soc/sifive/clint), [MFPR](https://github.com/usbarmory/tamago/tree/master/soc/spacemit/pinctrl.go), [UART](https://github.com/usbarmory/tamago/tree/master/soc/spacemit/uart) |

> [!WARNING]
> This package is in early development stages and its API might still change.

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkcpuinit`: include the K1 boot vector (`init.s`), which disables
  interrupts and enables the FPU (`MSTATUS.FS`) before the runtime starts, as
  required by the Go compiler emitted F/D instructions when the first stage
  boot loader does not leave the FPU enabled
* `linkramstart`: exclude `ramStart` from `mem.go`
* `linknanotime`: exclude the default `time` CSR based `nanotime`, allowing a
  board package to supply its own time source

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
