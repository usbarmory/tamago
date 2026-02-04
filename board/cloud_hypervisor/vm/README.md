TamaGo - bare metal Go - Cloud Hypervisor VM support
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
applications on bare metal processors.

The [vm](https://github.com/usbarmory/tamago/tree/master/board/cloud_hypervisor/vm)
package provides support for [Cloud Hypervisor VMs](https://www.cloudhypervisor.org)
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
[vm](https://github.com/usbarmory/tamago/tree/master/board/cloud_hypervisor/vm).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU              | Board                                                                | CPU package                                                    | Board package                                                                                    |
|------------------|----------------------------------------------------------------------|----------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| AMD/Intel 64-bit | [Cloud Hypervisor](https://www.cloudhypervisor.org)                  | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64) | [cloud_hypervisor/vm](https://github.com/usbarmory/tamago/tree/master/board/cloud_hypervisor/vm) |

Compiling
=========

Go applications are required to set `GOOSPKG` to the desired package overlay
and import the relevant board package to ensure that hardware initialization
and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/cloud_hypervisor/vm"
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
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago@v1.26.0 GOARCH=amd64 \
	${TAMAGO} build -ldflags "-T 0x10010000 -R 0x1000" main.go
```

An example application, targeting the Cloud Hypervisor platform, is
[available](https://github.com/usbarmory/tamago-example).

Build tags
==========

The following build tags allow application to override the package own
definition for the `runtime/goos` overlay:

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing and debugging
=======================

The [example application](https://github.com/usbarmory/tamago-example) provides
reference usage and a Makefile target for automatic creation of an ELF image
which can be executed under paravirtualization with
[cloud-hypervisor](https://www.cloudhypervisor.org/docs/prologue/quick-start/#firmware-booting):

```
cloud-hypervisor --kernel example --cpus boot=1 --memory size=4096M --net "tap=tap0" --serial tty --console off
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
