TamaGo - bare metal Go for ARM SoCs - USB armory support
========================================================

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

The [usbarmory](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)
package provides support for the [USB armory](https://github.com/usbarmory/usbarmory/wiki)
Single Board Computer.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC           | Board                                                           | SoC package                                                              | Board package                                                                        |
|---------------|-----------------------------------------------------------------|--------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| NXP i.MX6ULZ  | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul) | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory/mk2) |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/usbarmory/mk2"
)
```

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables and
ensuring that the required SoC and board packages are available in `GOPATH`:

```
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x80010000 -E _rt0_arm_tamago -R 0x1000"
```

An example application, targeting the USB armory Mk II platform,
is [available](https://github.com/usbarmory/tamago-example).

Executing and debugging
=======================

The [example application](https://github.com/usbarmory/tamago-example)
provides reference usage and a Makefile target for automatic creation of an ELF
as well as `imx` image for flashing.

Native hardware: imx image
--------------------------

Follow [these instructions](https://github.com/usbarmory/usbarmory/wiki/Boot-Modes-(Mk-II)#flashing-bootable-images-on-externalinternal-media)
using the built `imx` image.

Native hardware: existing bootloader
------------------------------------

Copy the built ELF binary on an external microSD card (replace `$dev` with `0`)
or the internal eMMC (replace `$dev` with `1`), then launch it from the U-Boot
console as follows:

```
ext2load mmc $dev:1 0x90000000 example
bootelf -p 0x90000000
```

For non-interactive execution modify the U-Boot configuration accordingly.

Standard output
---------------

The standard output can be accessed through the
[debug accessory](https://github.com/usbarmory/usbarmory/tree/master/hardware/mark-two-debug-accessory)
and the following `picocom` configuration:

```
picocom -b 115200 -eb /dev/ttyUSB2 --imap lfcrlf
```

Debugging
---------

The application can be debugged with GDB over JTAG using `openocd` (version >
0.11.0) and the `gdbinit` debugging helper published
[here](https://github.com/usbarmory/tamago/tree/master/_dev).

```
# start openocd daemon
openocd -f interface/ftdi/jtagkey.cfg -f target/imx6ul.cfg -c "adapter speed 1000"

# connect to the OpenOCD command line
telnet localhost 4444

# debug with GDB
arm-none-eabi-gdb -x gdbinit example
```

Hardware breakpoints can be set in the usual way:

```
hb ecdsa.Verify
continue
```

QEMU
----

The target can be executed under emulation as follows:

```
qemu-system-arm \
	-machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
	-nographic -monitor none -serial null -serial stdio -net none \
	-kernel example -semihosting
```

The emulated target can be debugged with GDB by adding the `-S -s` flags to the
previous execution command, this will make qemu waiting for a GDB connection
that can be launched as follows:

```
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
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
