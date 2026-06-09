TamaGo - bare metal Go - MilkV Vega support
===========================================

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

The [vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega)
package provides support for the [MilkV Vega](https://milkv.io/vega) board, a
RISC-V switch board powered by the Fisilink FSL91030 SoC (Nuclei UX600 core,
240 MB DRAM, 64 MB NOR flash).

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[fsl91030](https://github.com/usbarmory/tamago/tree/master/soc/fisilink/fsl91030).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC               | Board                               | SoC package                                                                       | Board package                                                            |
|-------------------|-------------------------------------|-----------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| Fisilink FSL91030 | [MilkV Vega](https://milkv.io/vega) | [fsl91030](https://github.com/usbarmory/tamago/tree/master/soc/fisilink/fsl91030) | [vega](https://github.com/usbarmory/tamago/tree/master/board/milkv/vega) |

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
	_ "github.com/usbarmory/tamago/board/milkv/vega"
)
```

The image is linked to run from DRAM (text segment at DRAM base + 64 KB):

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=riscv64 \
	${TAMAGO} build -ldflags "-T 0x41010000 -R 0x1000" main.go
```

The ELF entry point differs from the `-T` text base; extract it for the loader:

```sh
riscv64-linux-gnu-readelf -h main | awk '/Entry point/{print $4}'
```

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkcpuinit`: include the board reset vector that initializes DDR, cache and
  the FPU (for loading directly into DRAM, e.g. via JTAG)
* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing
=========

The board boots from NOR flash. A TamaGo image is linked to run from DRAM, so a
first stage must initialize DDR and relocate the image before jumping to it.
Two options are available:

* load the image into DRAM with the on-board vendor bootloader (e.g. U-Boot
  `bootelf` against the ELF `e_entry`), or
* flash a small relocating stage ahead of the image; an example `flashboot`
  stub for this is provided with the TamaGo examples.

For emulation under the Nuclei QEMU fork see the
[eval_soc](https://github.com/usbarmory/tamago/tree/master/board/nuclei/eval_soc)
board package.

Standard output
---------------

The standard output is exposed on UART0 (115200 8N1):

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
