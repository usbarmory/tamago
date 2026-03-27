TamaGo - bare metal Go - ET-SoC-1 support
=========================================

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

The [etsoc1](https://github.com/usbarmory/tamago/tree/master/soc/aifoundry/etsoc1)
package provides support for AI Foundry [ET-SoC-1](https://github.com/aifoundry-org/et-man)
processor.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| Processor                    | Related board packages                                                             | Processor package                                                                            | Peripheral drivers                                                    |
|------------------------------|------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|
| AI Foundry Minion            | [sys_emu](https://github.com/usbarmory/tamago/tree/master/board/aifoundry/sys_emu) | [etsoc1/minion](https://github.com/usbarmory/tamago/tree/master/soc/aifoundry/etsoc1/minion) | [UART](https://github.com/usbarmory/tamago/tree/master/soc/aifoundry) |

> [!TIP]
> The ET-Soc-1 mixes RISC-V cores with different ISA and access levels
> (`rv64gc` cores: 4x ET-Maxion, `rv64imfc` cores: 1088x Minion), for this
> reason the SoC package has sub-packages for importing processor specific
> support.

Soft float requirement
======================

This target requires a specific `GOOS=tamago` compiler branch to support the
following:

  * `GOSOFT=1`: compiler build time variable to enable soft float for `riscv64`, removing
    requirement for `ad` extensions and forcing single-threaded operation.

  * `tiny`: build tag to support considerable reduction of RAM allocation requirements.

The [kotama repository](https://github.com/usbarmory/kotama) provides
instructions and a reference implementation for this target.

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramstart`: exclude `ramStart` from `mem.go`
* `linkcpuinit`: override `cpuinit` imported from `riscv64/init.s` to park
                 additional harts (required on multi-hart instances).
* `tiny`: reduce heap allocation requirements

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
