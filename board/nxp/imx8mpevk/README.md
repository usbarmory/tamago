TamaGo - bare metal Go - 8MPLUSLPD4-EVK support
===============================================

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

The [imx8mpevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/imx8mpevk)
package provides support for the [8MPLUSLPD4-EVK](https://www.nxp.com/design/design-center/development-boards-and-designs/8MPLUSLPD4-EVK) development board.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[imx8mp](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx8mp).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC             | Board                                                                                                    | SoC package                                                              | Board package                                                                    |
|-----------------|----------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|----------------------------------------------------------------------------------|
| NXP i.MX8M Plus | [8MPLUSLPD4-EVK](https://www.nxp.com/design/design-center/development-boards-and-designs/8MPLUSLPD4-EVK) | [imx8mp](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx8mp) | [imx8mpevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/imx8mpevk) |

> [!WARNING]
> This package is in early development stages, only emulated runs (qemu) have been tested.

Compiling
=========

Go applications are required to set `GOOSPKG` to the desired
[runtime/goos](https://github.com/usbarmory/tamago-go/tree/latest/src/runtime/goos)
overlay and import the relevant board package to ensure that hardware
initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/nxp/imx8mpevk"
)
```

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```sh
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=arm64 \
	${TAMAGO} build -ldflags "-T 0x40010000 -R 0x1000" main.go
```

An example application, targeting the 8MPLUSLPD4-EVK platform,
is [available](https://github.com/usbarmory/tamago-example).

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing and debugging
=======================

The [example application](https://github.com/usbarmory/tamago-example) provides
reference usage and a Makefile target for automatic creation of an ELF image
for emulated execution.

QEMU
----

The target can be executed under emulation as follows:

```sh
qemu-system-aarch64 \
	-machine imx8mp-evk -m 512M \
	-nographic -monitor none -serial stdio -serial null \
	-net nic,model=imx.enet,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
	-kernel example -semihosting
```

The emulated target can be debugged with GDB by adding the `-S -s` flags to the
previous execution command, this will make qemu waiting for a GDB connection
that can be launched as follows:

```sh
arm-none-eabi-gdb -ex "target remote 127.0.0.1:1234" example
```

Breakpoints can be set in the usual way:

```
b ecdsa.Verify
continue
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
