TamaGo - bare metal Go - i.MX 8M Plus support
=============================================

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

The [imx8mp](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx8mp)
package provides support for the NXP i.MX 8M Plus series of System-on-Chip
(SoCs) parts.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC              | Related board packages                                                               | Peripheral drivers                                                                                                                                                                                                                                                                 |
|------------------|--------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------|
| NXP i.MX 8M Plus | [nxp/imx8mpevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/imx8mpevk) | [CAAM, ENET, OCOTP, UART, USDHC, WDOG](https://github.com/usbarmory/tamago/tree/master/soc/nxp) |

> [!WARNING]
> This package is in early development stages, only emulated runs (qemu) have been tested.

Build tags
==========

The following build tags allow application to override the package own definition of
[external functions required by the runtime](https://pkg.go.dev/github.com/usbarmory/tamago/doc):

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
