TamaGo - bare metal Go - Nuvoton NUC980 support
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

The [nuc980](https://github.com/usbarmory/tamago/tree/master/soc/nuvoton/nuc980)
package provides support for Nuvoton NUC980 microprocessors.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC            | Related board packages                                                                         | Peripheral drivers                                                                     |
|----------------|------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|
| Nuvoton NUC980 | [nuvoton/nuc980iiot](https://github.com/usbarmory/tamago/tree/master/board/nuvoton/nuc980iiot) | [AIC, ETIMER, PRNG, UART](https://github.com/usbarmory/tamago/tree/master/soc/nuvoton) |

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramstart`: exclude `ramStart` from `mem.go`
* `linkcpuinit`: exclude `cpuinit` imported from `arm/init.s`

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
