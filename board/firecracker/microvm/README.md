TamaGo - bare metal Go - Firecracker microvm support
====================================================

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
applications on bare metal AMD64/ARM/RISC-V processors.

The [microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm)
package provides support for [Firecracker microvm](https://firecracker-microvm.github.io/)
paravirtualized Kernel-based Virtual Machine (KVM) configured with single or
multiple AMD64 cores.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[amd64](https://github.com/usbarmory/tamago/tree/master/amd64) and
[microvm](https://github.com/usbarmory/tamago/tree/master/board/firecraker/microvm).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU              | Board                                                                | CPU package                                                    | Board package                                                                                    |
|------------------|----------------------------------------------------------------------|----------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| AMD/Intel 64-bit | [Firecracker microvm](https://firecracker-microvm.github.io)         | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64) | [firecracker/microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm) |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/firecracker/microvm"
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
previous step, but with the addition of the following flags/variables:

```
GOOS=tamago GOARCH=amd64 ${TAMAGO} build -ldflags "-T 0x10010000 -R 0x1000" main.go
```

An example application, targeting the Firecracker microvm platform,
is [available](https://github.com/usbarmory/tamago-example).

Build tags
==========

The following build tags allow application to override the package own definition of
[external functions required by the runtime](https://pkg.go.dev/github.com/usbarmory/tamago/doc):

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing and debugging
=======================

The [example application](https://github.com/usbarmory/tamago-example) provides
reference usage and a Makefile target for automatic creation of an ELF image
which can be executed under paravirtualization with
[firectl](https://github.com/firecracker-microvm/firectl) or
[firecracker](https://github.com/firecracker-microvm/firecracker):

Firectl
-------

```
firectl --kernel example --root-drive /dev/null --tap-device tap0/06:00:AC:10:00:01 -c 1 -m 4096
```

Firecracker
-----------

```
firecracker --config-file vm_config.json
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
