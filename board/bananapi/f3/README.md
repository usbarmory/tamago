TamaGo - bare metal Go - Banana Pi BPI-F3 support
=================================================

tamago | https://github.com/usbarmory/tamago

Copyright (c) The TamaGo Authors. All Rights Reserved.

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal processors.

The [f3](https://github.com/usbarmory/tamago/tree/master/board/bananapi/f3)
package provides support for the [Banana Pi BPI-F3](https://docs.banana-pi.org/en/BPI-F3/BananaPi_BPI-F3)
board, an industrial grade single board computer powered by the SpacemiT K1 SoC
(eight RV64GCVB SpacemiT X60 cores, 2/4/8/16 GB LPDDR4x DRAM).

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[k1](https://github.com/usbarmory/tamago/tree/master/soc/spacemit/k1).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC         | Board                                                                     | SoC package                                                             | Board package                                                                     |
|-------------|-----------------------------------------------------------------------------|-------------------------------------------------------------------------|-------------------------------------------------------------------------------------|
| SpacemiT K1 | [Banana Pi BPI-F3](https://docs.banana-pi.org/en/BPI-F3/BananaPi_BPI-F3) | [k1](https://github.com/usbarmory/tamago/tree/master/soc/spacemit/k1)  | [bananapi/f3](https://github.com/usbarmory/tamago/tree/master/board/bananapi/f3)  |

> [!WARNING]
> This package is in early development stages and its API might still change,
> see the [k1](https://github.com/usbarmory/tamago/tree/master/soc/spacemit/k1)
> package for the list of supported SoC peripherals.

Compiling
=========

Go distribution supporting `GOOS=tamago`
---------------------------------------

The [tamago](https://github.com/usbarmory/tamago/tree/latest/cmd/tamago)
command downloads, compiles, and runs the `go` command from the
[TamaGo distribution](https://github.com/usbarmory/tamago-go) matching the
tamago module version from the application `go.mod`.

Applications can add `github.com/usbarmory/tamago` to `go.mod`, and then
replace the `go` command with:

```sh
go run github.com/usbarmory/tamago/cmd/tamago
```

Alternatively the
[latest TamaGo distribution](https://github.com/usbarmory/tamago-go/tree/latest)
can be manually built:

```sh
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Building applications
---------------------

Go applications are required to set `GOOSPKG` to the desired
[runtime/goos](https://github.com/usbarmory/tamago-go/tree/latest/src/runtime/goos)
overlay and import the relevant board package:

```golang
import (
	_ "github.com/usbarmory/tamago/board/bananapi/f3"
)
```

The image is linked to run from DRAM, with the text segment placed 64 KB into
the runtime managed memory region (`k1.ramStart`, 0x02000000):

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=riscv64 \
	${TAMAGO} build -ldflags "-T 0x02010000 -R 0x1000" main.go
```

The ELF entry point differs from the `-T` text base; extract it for the loader:

```sh
riscv64-linux-gnu-readelf -h main | awk '/Entry point/{print $4}'
```

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkcpuinit`: include the K1 boot vector (`soc/spacemit/k1/init.s`) which
  disables interrupts and enables the FPU before the runtime starts, it is
  recommended for a hand-off from a first stage boot loader
* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing
=========

The K1 boot ROM loads a first stage from SPI NAND flash, SPI NOR flash, eMMC or
SD/TF card according to the QSPI_DATA[1:0] straps (Section 2.12, K1 Datasheet).
DDR initialization is performed by the vendor first stage, a TamaGo image is
therefore linked to run from DRAM and must be loaded and entered by an earlier
boot stage.

> [!IMPORTANT]
> The `k1` package initializes the core in machine mode (it configures `mtvec`),
> the image must therefore be entered at machine level, taking the place of the
> SBI implementation in the U-Boot SPL hand-off. Entering the image at
> supervisor level, for instance with `bootelf` from a running U-Boot, requires
> supervisor mode support which is not yet implemented.

The [example](https://github.com/usbarmory/tamago/tree/master/board/bananapi/f3/example)
application provides a console smoke test along with the U-Boot SPL FIT image
integration details.

Standard output
---------------

The standard output is exposed on UART0 (115200 8N1), routed to the pads
controlled by GPIO_68 (TX) and GPIO_69 (RX) which are wired to the board 40-pin
GPIO header debug console:

```sh
picocom -b 115200 -eb /dev/ttyUSB0 --imap lfcrlf
```

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
