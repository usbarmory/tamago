TamaGo - bare metal Go for ARM SoCs
===================================

tamago | https://github.com/inversepath/tamago  

Copyright (c) F-Secure Corporation  
https://foundry.f-secure.com

![TamaGo gopher](https://github.com/inversepath/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com | andrej@inversepath.com  

Introduction
============

TamaGo is a project that aims to provide compilation and execution of
unencumbered Go applications for bare metal ARM System-on-Chip (SoC)
components.

The projects spawns from the desire of reducing the attack surface of embedded
systems firmware by removing any runtime dependency on C code and Operating
Systems.

The TamaGo framework consists of the following components:

 - A modified [Go compiler](https://github.com/inversepath/tamago-go)
   which extends `GOOS` support to the `tamago` target, allowing bare metal
   execution for ARMv7 architecture.

 - Go packages for SoC driver support.

 - Go packages for board support.

The modifications are meant to be minimal for both the Go compiler (< ~4000 LOC
changed) and the target application (one import required), with a clean
separation/architecture from the rest of the Go compiler.

Strong emphasis is placed on code re-use from existing architectures already
included within the standard Go runtime, see
[Internals](https://github.com/inversepath/tamago/wiki/Internals).

Both aspects are motivated by the desire of providing a framework that allows
secure Go firmware development on embedded systems.

Current release level
=====================

The current release is
[tamago1.13.6](https://github.com/inversepath/tamago-go/tree/tamago1.13.6),
which [adds](https://github.com/golang/go/compare/go1.13.6...inversepath:tamago1.13.6)
`GOOS=tamago` support to go-1.13.6 release.

TamaGo is in early stages of development, all code should be considered at
alpha stage and yet ready for production use.

Documentation
=============

The main documentation can be found on the
[project wiki](https://github.com/inversepath/tamago/wiki).

Supported hardware
==================

The following table summarizes currently supported SoCs and boards.

| SoC           | Board                                                             | SoC package                                                    | Board package                                                                              |
|---------------|-------------------------------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------------|
| NXP i.MX6ULL | [USB armory Mk II](https://github.com/inversepath/usbarmory/wiki) | [imx6](https://github.com/inversepath/tamago/tree/master/imx6) | [usbarmory/mark-two](https://github.com/inversepath/tamago/tree/master/usbarmory) |

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
git clone https://github.com/inversepath/tamago-go -b tamago1.13.6
cd tamago-go/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled with the compiler built in the previous step
but with the addition of the following flags/variables, also make sure that the
required SoC and board packages are available in your `GOPATH`:

```
# USB armory Mk II example from the root directory of this repository
cd example &&
  GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x80010000  -E _rt0_arm_tamago -R 0x1000"
```

Executing and debugging
=======================

See the respective board package README file for execution and debugging
information for each specific target (real or emulated).

An emulated run of the [example application](https://github.com/inversepath/tamago/tree/master/example)
can be launched as follows:

```
make clean && make qemu
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
