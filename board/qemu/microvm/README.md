TamaGo - bare metal Go - QEMU microvm support
=============================================

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

The [microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm)
package provides support for [QEMU microvm](https://www.qemu.org/docs/master/system/i386/microvm.html)
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
[microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| CPU              | Board                                                                     | CPU package                                                    | Board package                                                                      |
|------------------|---------------------------------------------------------------------------|----------------------------------------------------------------|------------------------------------------------------------------------------------|
| AMD/Intel 64-bit | [QEMU microvm](https://www.qemu.org/docs/master/system/i386/microvm.html) | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64) | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm) |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/qemu/microvm"
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

An example application, targeting the QEMU microvm platform,
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
reference usage and a Makefile target for automatic creation of an ELF image as
well as paravirtualized execution.

QEMU
----

```
qemu-system-x86_64 \
	-machine microvm,x-option-roms=on,pit=off,pic=off,rtc=on \
	-global virtio-mmio.force-legacy=false \
	-enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
	-m 4G -nographic -monitor none -serial stdio \
        -device virtio-net-device,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
	-kernel example
```

The paravirtualized target can be debugged with GDB by adding the `-S -s` flags
to the previous execution command, this will make qemu waiting for a GDB
connection that can be launched as follows:

```
gdb -ex "target remote 127.0.0.1:1234" example
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
