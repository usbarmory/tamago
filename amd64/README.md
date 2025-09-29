TamaGo - bare metal Go - AMD/Intel 64-bit support
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

The [amd64](https://github.com/usbarmory/tamago/tree/master/amd64) package
provides support for AMD/Intel 64-bit CPUs.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU              | Related platform packages                                                                        | Core drivers                                                         | Peripheral drivers                                                                                                                                                                  |
|------------------|--------------------------------------------------------------------------------------------------|----------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AMD/Intel 64-bit | [cloud_hypervisor/vm](https://github.com/usbarmory/tamago/tree/master/board/cloud_hypervisor/vm) | [LAPIC](https://github.com/usbarmory/tamago/tree/master/amd64/lapic) | [KVM clock, VirtIO over PCI](https://github.com/usbarmory/tamago/tree/master/kvm), [IOAPIC, PCI, RTC, UART](https://github.com/usbarmory/tamago/blob/master/soc/intel)              |
| AMD/Intel 64-bit | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm)               | [LAPIC](https://github.com/usbarmory/tamago/tree/master/amd64/lapic) | [KVM clock, VirtIO over MMIO](https://github.com/usbarmory/tamago/tree/master/kvm), [IOAPIC, RTC, UART](https://github.com/usbarmory/tamago/blob/master/soc/intel)                  |
| AMD/Intel 64-bit | [firecracker/microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm) | [LAPIC](https://github.com/usbarmory/tamago/tree/master/amd64/lapic) | [KVM clock, VirtIO over MMIO](https://github.com/usbarmory/tamago/tree/master/kvm), [IOAPIC, UART](https://github.com/usbarmory/tamago/blob/master/soc/intel)                       |
| AMD/Intel 64-bit | [uefi/x64](https://github.com/usbarmory/go-boot/tree/main/uefi/x64)                              |                                                                      | [EFI Console I/O, Graphics, Boot and Runtime Services](https://github.com/usbarmory/go-boot/tree/main/uefi), [RTC, UART](https://github.com/usbarmory/tamago/blob/master/soc/intel) |

Build tags
==========

The following build tags allow application to override the package own definition of
[external functions required by the runtime](https://pkg.go.dev/github.com/usbarmory/tamago/doc):

* `linkramstart`: exclude `ramStart` from `mem.go`
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
