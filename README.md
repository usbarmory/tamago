TamaGo - bare metal Go for ARM/RISC-V SoCs
==========================================

tamago | https://github.com/usbarmory/tamago  

Copyright (c) WithSecure Corporation  
https://foundry.withsecure.com

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM/RISC-V System-on-Chip (SoC) components.

The projects spawns from the desire of reducing the attack surface of embedded
systems firmware by removing any runtime dependency on C code and Operating
Systems.

The TamaGo framework consists of the following components:

 - A modified [Go distribution](https://github.com/usbarmory/tamago-go)
   which extends `GOOS` support to the `tamago` target, allowing bare metal
   execution.

 - Go packages for SoC driver support.

 - Go packages for board support.

The modifications are meant to be minimal for both the Go distribution (< ~4000
LOC changed) and the target application (one import required), with a clean
separation from other architectures.

Strong emphasis is placed on code re-use from existing architectures already
included within the standard Go runtime, see
[Internals](https://github.com/usbarmory/tamago/wiki/Internals).

Both aspects are motivated by the desire of providing a framework that allows
secure Go firmware development on embedded systems.

Current release level
=====================
[![GitHub release](https://img.shields.io/github/v/release/usbarmory/tamago-go)](https://github.com/usbarmory/tamago-go/tree/latest) [![Build Status](https://github.com/usbarmory/tamago-go/workflows/Build%20Go%20compiler/badge.svg)](https://github.com/usbarmory/tamago-go/actions)

The current release for the [TamaGo modified Go distribution](https://github.com/usbarmory/tamago-go) is
[tamago1.19.4](https://github.com/usbarmory/tamago-go/tree/tamago1.19.4),
which [adds](https://github.com/golang/go/compare/go1.19.4...usbarmory:tamago1.19.4)
`GOOS=tamago` support to go1.19.4.

Binary releases for amd64 and armv7l Linux hosts [are available](https://github.com/usbarmory/tamago-go/releases/latest).

Documentation
=============

The main documentation can be found on the
[project wiki](https://github.com/usbarmory/tamago/wiki).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).


Supported ARM hardware 
=======================

The following table summarizes currently supported ARM SoCs and boards
(`GOOS=tamago GOARCH=arm`).

| SoC          | Board                                                                                                                                                                                | SoC package                                                              | Board package                                                                    |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|----------------------------------------------------------------------------------|
| NXP i.MX6ULZ | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki)                                                                                                                      | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul) | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory) |
| NXP i.MX6ULL | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul) | [mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk) |
| BCM2835      | [Raspberry Pi Zero](https://www.raspberrypi.org/products/raspberry-pi-zero)                                                                                                          | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835)   | [pi/pizero](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi)   |
| BCM2835      | [Raspberry Pi 1 Model A+](https://www.raspberrypi.org/products/raspberry-pi-1-model-a-plus/)                                                                                         | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835)   | [pi/pi1](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi)      |
| BCM2835      | [Raspberry Pi 1 Model B+](https://www.raspberrypi.org/products/raspberry-pi-1-model-b-plus/)                                                                                         | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835)   | [pi/pi1](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi)      |
| BCM2836      | [Raspberry Pi 2 Model B](https://www.raspberrypi.org/products/raspberry-pi-2-model-b)                                                                                                | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835)   | [pi/pi2](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi)      |

Supported RISC-V hardware 
==========================

The following table summarizes currently supported RISC-V SoCs and boards
(`GOOS=tamago GOARCH=riscv64`).

| SoC          | Board                                                                        | SoC package                                                               | Board package                                                                        |
|--------------|------------------------------------------------------------------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| SiFive FU540 | [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html) | [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u) |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
	// Example for USB armory Mk II
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

Go applications can be compiled with the compiler built in the previous step,
with the addition of a few flags/variables:

```
# Example for USB armory Mk II
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x80010000 -E _rt0_arm_tamago -R 0x1000"

# Example for QEMU RISC-V sifive_u
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARCH=riscv64 \
  ${TAMAGO} build -ldflags "-T 0x80010000 -E _rt0_riscv64_tamago -R 0x1000"
```

See the respective board package README file for compilation information for
each specific target.

Executing and debugging
=======================

See the respective board package README file for execution and debugging
information for each specific target (real or emulated).

The [example application](https://github.com/usbarmory/tamago-example)
provides sample driver usage and instructions for native as well as emulated
execution.

An emulated run of the [example application](https://github.com/usbarmory/tamago-example)
can be launched as follows:

```
git clone https://github.com/usbarmory/tamago-example
cd tamago-example && make qemu
```

Applications using TamaGo
=========================

* [GoKey](https://github.com/usbarmory/GoKey) - the bare metal Go smartcard
* [GoTEE](https://github.com/usbarmory/GoTEE) - Go Trusted Execution Environment
* [armory-drive](https://github.com/usbarmory/armory-drive) - USB encrypted drive
* [armory-ums](https://github.com/usbarmory/armory-ums) - USB Mass Storage firmware
* [armory-boot](https://github.com/usbarmory/armory-boot) - primary bootloader
* [tamago-example](https://github.com/usbarmory/tamago-example) - TamaGo example application

External drivers using TamaGo
=============================

* [imx-usbnet](https://github.com/usbarmory/imx-usbnet) - i.MX Ethernet over USB driver
* [imx-enet](https://github.com/usbarmory/imx-enet) - i.MX Ethernet driver

Additional resources
====================

* [Package API](https://pkg.go.dev/github.com/usbarmory/tamago)
* [Internals](https://github.com/usbarmory/tamago/wiki/Internals)
* [Go stdlib support](https://github.com/usbarmory/tamago/wiki/Import-report)
* [FAQ](https://github.com/usbarmory/tamago/wiki/Frequently-Asked-Questions-(FAQ))
* [TamaGo presentation](https://github.com/abarisani/abarisani.github.io/tree/master/research/tamago)

Maintainers
===========

Andrea Barisani  
andrea.barisani@withsecure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@withsecure.com   | andrej@inversepath.com  

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

This project is distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
