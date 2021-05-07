TamaGo - bare metal Go for ARM SoCs
===================================

tamago | https://github.com/f-secure-foundry/tamago  

Copyright (c) F-Secure Corporation  
https://foundry.f-secure.com

![TamaGo gopher](https://github.com/f-secure-foundry/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com   | andrej@inversepath.com  

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM System-on-Chip (SoC) components.

The projects spawns from the desire of reducing the attack surface of embedded
systems firmware by removing any runtime dependency on C code and Operating
Systems.

The TamaGo framework consists of the following components:

 - A modified [Go distribution](https://github.com/f-secure-foundry/tamago-go)
   which extends `GOOS` support to the `tamago` target, allowing bare metal
   execution.

 - Go packages for SoC driver support.

 - Go packages for board support.

The modifications are meant to be minimal for both the Go distribution (< ~4000
LOC changed) and the target application (one import required), with a clean
separation from other architectures.

Strong emphasis is placed on code re-use from existing architectures already
included within the standard Go runtime, see
[Internals](https://github.com/f-secure-foundry/tamago/wiki/Internals).

Both aspects are motivated by the desire of providing a framework that allows
secure Go firmware development on embedded systems.

Current release level
=====================
[![GitHub release](https://img.shields.io/github/v/release/f-secure-foundry/tamago-go)](https://github.com/f-secure-foundry/tamago-go/tree/latest) [![Build Status](https://github.com/f-secure-foundry/tamago-go/workflows/Build%20Go%20compiler/badge.svg)](https://github.com/f-secure-foundry/tamago-go/actions)

The current release for the [TamaGo modified Go distribution](https://github.com/f-secure-foundry/tamago-go) is
[tamago1.16.4](https://github.com/f-secure-foundry/tamago-go/tree/tamago1.16.4),
which [adds](https://github.com/golang/go/compare/go1.16.4...f-secure-foundry:tamago1.16.4)
`GOOS=tamago` support to go1.16.4.

Binary releases for amd64 and armv7l Linux hosts [are available](https://github.com/f-secure-foundry/tamago-go/releases/latest).

Documentation
=============

The main documentation can be found on the
[project wiki](https://github.com/f-secure-foundry/tamago/wiki).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/f-secure-foundry/tamago).

Supported hardware
==================

The following table summarizes currently supported SoCs and boards.

| SoC           | Board                                                                                                                                                                                | SoC package                                                                   | Board package                                                                                          |
|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| NXP i.MX6ULZ  | [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki)                                                                                                               | [imx6](https://github.com/f-secure-foundry/tamago/tree/master/soc/imx6)       | [usbarmory/mark-two](https://github.com/f-secure-foundry/tamago/tree/master/board/f-secure/usbarmory)  |
| NXP i.MX6ULL  | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6](https://github.com/f-secure-foundry/tamago/tree/master/soc/imx6)       | [mx6ullevk](https://github.com/f-secure-foundry/tamago/tree/master/board/nxp/mx6ullevk)                |
| BCM2835       | [Raspberry Pi Zero](https://www.raspberrypi.org/products/raspberry-pi-zero)                                                                                                          | [bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/soc/bcm2835) | [pi/pizero](https://github.com/f-secure-foundry/tamago/tree/master/board/raspberrypi)                  |
| BCM2835       | [Raspberry Pi 1 Model A+](https://www.raspberrypi.org/products/raspberry-pi-1-model-a-plus/)                                                                                         | [bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/soc/bcm2835) | [pi/pi1](https://github.com/f-secure-foundry/tamago/tree/master/board/raspberrypi)                     |
| BCM2835       | [Raspberry Pi 1 Model B+](https://www.raspberrypi.org/products/raspberry-pi-1-model-b-plus/)                                                                                         | [bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/soc/bcm2835) | [pi/pi1](https://github.com/f-secure-foundry/tamago/tree/master/board/raspberrypi)                     |
| BCM2836       | [Raspberry Pi 2 Model B](https://www.raspberrypi.org/products/raspberry-pi-2-model-b)                                                                                                | [bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/soc/bcm2835) | [pi/pi2](https://github.com/f-secure-foundry/tamago/tree/master/board/raspberrypi)                     |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
	// Example for USB armory Mk II
	_ "github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
)
```

Build the [TamaGo compiler](https://github.com/f-secure-foundry/tamago-go)
(or use the [latest binary release](https://github.com/f-secure-foundry/tamago-go/releases/latest)):

```
wget https://github.com/f-secure-foundry/tamago-go/archive/refs/tags/latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled with the compiler built in the previous step,
with the addition of a few flags/variables:

```
# Example for USB armory Mk II
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x80010000  -E _rt0_arm_tamago -R 0x1000"
```

See the respective board package README file for compilation information for
each specific target.

Executing and debugging
=======================

See the respective board package README file for execution and debugging
information for each specific target (real or emulated).

The [example application](https://github.com/f-secure-foundry/tamago-example)
provides sample driver usage and instructions for native as well as emulated
execution.

License
=======

tamago | https://github.com/f-secure-foundry/tamago  
Copyright (c) F-Secure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/f-secure-foundry/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
