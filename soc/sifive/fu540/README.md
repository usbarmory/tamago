TamaGo - bare metal Go for RISC-V SoCs - FU540 support
======================================================

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

The [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) package
provides support for the SiFive FU540 System-on-Chip (SoCs).

Documentation
=============

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC          | Related board packages                                                               | Peripheral drivers                                        |
|--------------|--------------------------------------------------------------------------------------|-----------------------------------------------------------|
| SiFive FU540 | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u) | UART                                                      |

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
