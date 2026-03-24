TamaGo - bare metal Go - FSL91030 support
==========================================

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

The [fsl91030](https://github.com/usbarmory/tamago/tree/master/soc/fisilink/fsl91030)
package provides support for the Fisilink FSL91030, a RISC-V 64-bit SoC based
on the [Nuclei UX600](https://www.nucleisys.com/product.php) core.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC      | Board packages                                                                          | Peripheral drivers                                    |
|----------|-----------------------------------------------------------------------------------------|-------------------------------------------------------|
| FSL91030 | [milkv/vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega)         | UART, GPIO, WDT, CLINT timer, cache control, DDR      |

> [!WARNING]
> This package is in early development stages, only emulated runs (QEMU) have been tested.

Build tags
==========

The following build tags allow applications to override package definitions
for the `runtime/goos` overlay:

* `linkramstart`: exclude `ramStart` from `mem.go`
* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude the default `printk`
* `linknanotime`: exclude the default `nanotime`; the board package can
  supply a custom one (e.g. `milkv-vega-qemu` for Nuclei QEMU timer)
* `linkcpuinit`: exclude the default `cpuinit`; `boot_riscv64.s` provides a
  full DDR/cache/QSPI init sequence for builds that bypass `flashboot.s`

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
