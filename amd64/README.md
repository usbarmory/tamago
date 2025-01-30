TamaGo - bare metal Go - AMD/Intel 64-bit support
=================================================

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

The [amd64](https://github.com/usbarmory/tamago/tree/master/amd64) package
provides support for AMD/Intel 64-bit CPUs.

Documentation
=============

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU              | Related board packages                                                                           | Peripheral drivers                                                                                                                                             |
|------------------|--------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AMD/Intel 64-bit | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm)               | [kvmclock, VirtIO](https://github.com/usbarmory/tamago/tree/master/kvm), [LAPIC, IOAPIC, RTC, UART](https://github.com/usbarmory/tamago/blob/master/soc/intel) |
| AMD/Intel 64-bit | [Firecracker microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm) | [kvmclock, VirtIO](https://github.com/usbarmory/tamago/tree/master/kvm), [LAPIC, IOAPIC, RTC, UART](https://github.com/usbarmory/tamago/blob/master/soc/intel) |

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
