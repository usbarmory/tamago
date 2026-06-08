TamaGo - bare metal Go - NuMaker-IIoT-NUC980G2 support
======================================================

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

The [nuc980iiot](https://github.com/usbarmory/tamago/tree/master/board/nuvoton/nuc980iiot)
package provides support for the
[NuMaker-IIoT-NUC980G2](https://www.nuvoton.com/products/iot-solution/iot-platform/numaker-iiot-nuc980g2/)
development board (NUC980DK71YC, 128 MB DDR2).

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[nuc980](https://github.com/usbarmory/tamago/tree/master/soc/nuvoton/nuc980).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC            | Board                                                                                                     | SoC package                                                                    | Board package                                                                            |
|----------------|-----------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------|------------------------------------------------------------------------------------------|
| Nuvoton NUC980 | [NuMaker-IIoT-NUC980G2](https://www.nuvoton.com/products/iot-solution/iot-platform/numaker-iiot-nuc980g2) | [nuc980](https://github.com/usbarmory/tamago/tree/master/soc/nuvoton/nuc980)   | [nuc980iiot](https://github.com/usbarmory/tamago/tree/master/board/nuvoton/nuc980iiot)   |

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

or add the following line to `go.mod` to use `go tool tamago` as go command:

```
tool github.com/usbarmory/tamago/cmd/tamago
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
	_ "github.com/usbarmory/tamago/board/nuvoton/nuc980iiot"
)
```

Applications are compiled with `GOARM=5` and the `linkcpuinit` tag, which
installs the board reset vector required to boot from the Nuvoton boot ROM:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARM=5 GOARCH=arm \
	${TAMAGO} build -tags linkcpuinit -ldflags "-T 0x00010000 -R 0x1000" main.go
```

> [!NOTE]
> The ARM926EJ-S is an ARMv5TE core without LDREX/STREX; a single-processor
> compare-and-swap is selected which `GOARM=5`.

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkcpuinit`: include the board reset vector (`cpuinit.s`)
* `linkramstart`: exclude `ramStart` from the `nuc980` `mem.go`
* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing
=========

The boot ROM loads images that embed the DDR initialization parameters for the
on-board SDRAM. Convert the compiled ELF to a flat binary and wrap it in a boot
image for the desired medium (USB, microSD or NAND) using
[NuWriter](https://github.com/OpenNuvoton/NUC980_NuWriter) or an equivalent
image tool, then boot the board with the matching boot-select straps.

Standard output
---------------

The standard output is exposed on UART0 (115200 8N1) via the board debug
connector:

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
