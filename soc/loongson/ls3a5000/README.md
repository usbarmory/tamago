TamaGo - bare metal Go - Loongson 3A5000/LS7A support
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

The `ls3a5000` package provides support for Loongson 3A5000/3A6000 processors
paired with the LS7A bridge, as emulated by the QEMU LoongArch `virt` machine.

The package exposes the LoongArch CPU instance along with the NS16550 legacy
serial console; the constant-frequency stable timer feeds the runtime system
clock.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

For the CPU driver support see package
[loong64](https://github.com/usbarmory/tamago/tree/master/loong64).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC             | Related board packages                                                        | Peripheral drivers |
|-----------------|-------------------------------------------------------------------------------|--------------------|
| Loongson 3A5000 | [qemu/virt](https://github.com/usbarmory/tamago/blob/master/board/qemu/virt)  | UART               |

Secure random number generation
===============================

The Loongson SoC does not currently wire an entropy source, therefore at
runtime initialization a DRBG is seeded with the CPU timer. This is unsuitable
for secure random number generation and must therefore be overridden through
`SetRNG` to ensure secure operation of Go crypto.

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
