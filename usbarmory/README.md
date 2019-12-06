TamaGo - bare metal Go for ARM SoCs - USB armory support
========================================================

tamago | https://github.com/inversepath/tamago  

Copyright (c) F-Secure Corporation  
https://foundry.f-secure.com

![TamaGo gopher](https://github.com/inversepath/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Introduction
============

TamaGo is a project that aims to provide compilation and execution of
unencumbered Go applications for bare metal ARM System-on-Chip (SoC)
components.

The [usbarmory](https://github.com/inversepath/tamago/tree/master/usbarmory)
package provide support for the [USB
armory](https://github.com/inversepath/usbarmory/wiki) Single Board Computer.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/inversepath/tamago) and
[project wiki](https://github.com/inversepath/tamago/wiki).

For the underlying driver support for this board see package
[imx6](https://github.com/inversepath/tamago/tree/master/imx6).

Supported hardware
==================

| SoC           | Board                                                             | SoC package                                                    | Board package                                                                              |
|---------------|-------------------------------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------------------|
| NXP i.MX6ULL | [USB armory Mk II](https://github.com/inversepath/usbarmory/wiki) | [imx6](https://github.com/inversepath/tamago/tree/master/imx6) | [usbarmory/mark-two](https://github.com/inversepath/tamago/tree/master/usbarmory/mark-two) |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
	_ "usbarmory/mark-two"
)
```

Build the [TamaGo compiler](https://github.com/inversepath/tamago-go):

```
git clone https://github.com/inversepath/tamago-go -b tamago1.13.5
cd tamago-go/src && ./make.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables and
ensuring that the required SoC and board packages are available in `GOPATH`:

```
# USB armory Mk II example
git clone https://github.com/inversepath/tamago && cd tamago
cd example &&
  GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x80010000  -E _rt0_arm_tamago -R 0x1000"
```

Executing and debugging
=======================

Native hardware
---------------

Copy the compiled application on an external microSD card (replace `$dev` with
`0`) or the internal eMMC (replace `$dev` with `1`), then launch it from the
U-Boot console as follows:

```
ext2load mmc $dev:1 0x90000000 example
bootelf -p 0x90000000
```

For non-interactive execution modify the U-Boot configuration accordingly.

The standard output can be accessed through the
[debug accessory](https://github.com/inversepath/usbarmory/tree/master/hardware/mark-two-debug-accessory)
and the following `picocom` configuration:

```
picocom -b 115200 -eb /dev/ttyUSB2 --imap lfcrlf
```

The application can be debugged with GDB over JTAG using `openocd` and the
`imx6ull.cfg` and `gdbinit` debugging helpers published
[here](https://github.com/inversepath/tamago/tree/master/dev).

```
# start openocd daemon
openocd -f interface/ftdi/jtagkey.cfg -f imx6ull.cfg

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
	-machine mcimx6ul-evk -cpu cortex-a7 -m 512M
	\ -nographic -monitor none -serial null -serial stdio -net none
	\ -kernel example -semihosting -d unimp
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

tamago | https://github.com/inversepath/tamago  
Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
