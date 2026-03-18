TamaGo - bare metal Go - EVB-LAN9696-24port support
===================================================

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

The [lan9696evb](https://github.com/usbarmory/tamago/tree/master/board/microchip/lan9696evb)
package provides support for the Microchip [EVB-LAN9696-24port](https://www.microchip.com/en-us/development-tool/ev23x71a)
(ev23x71a) development board.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[lan969x](https://github.com/usbarmory/tamago/tree/master/soc/microchip/lan969x).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC               | Board                                                                           | SoC package                                                                      | Board package                                                                            |
|-------------------|---------------------------------------------------------------------------------|----------------------------------------------------------------------------------|------------------------------------------------------------------------------------------|
| Microchip LAN969x | [EVB-LAN9696-24port](https://www.microchip.com/en-us/development-tool/ev23x71a) | [lan969x](https://github.com/usbarmory/tamago/tree/master/soc/microchip/lan969x) | [lan9696evb](https://github.com/usbarmory/tamago/tree/master/board/microchip/lan9696evb) |

Compiling
=========

Go distribution supporting `GOOS=tamago`
---------------------------------------

The [tamago](https://github.com/usbarmory/tamago/tree/latest/cmd/tamago)
command downloads, compiles, and runs the `go` command from the
[TamaGo distribution](https://github.com/usbarmory/tamago-go) matching the                                                                                                                                          tamago module version from the application `go.mod`.

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
[latest TamaGo distribution](https://github.com/usbarmory/tamago-go/tree/latest) can be
manually built or the
[latest binary release](https://github.com/usbarmory/tamago-go/releases/latest) can be used:

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
overlay and import the relevant board package to ensure that hardware
initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/tamago/board/microchip/lan9696evb"
)
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=arm64 \
	${TAMAGO} build -ldflags "-T 0x60010000 -R 0x1000" main.go
```

Build tags
==========

The following build tags allow applications to override the package own
definition for the `runtime/goos` overlay:

* `linkramsize`: exclude `ramSize` from `mem.go`
* `linkprintk`: exclude `printk` from `console.go`

Executing and debugging
=======================

On LAN969x platform bare metal Go unikernels can be loaded by directly by ARM
[Trusted Firmware-A](https://github.com/ARM-software/arm-trusted-firmware)
as BL31 or BL33 after
[BL1 and BL2](https://microchip-ung.github.io/bsp-doc/bsp/2025.09/supported-hw/lan969x-boot.html).

This is required as this library does not provide the required DDR
configuration support, done in BL2, to run directly from the LAN969x on-chip
RAM

A patch for the
[Microchip Arm Trusted Firmware](https://github.com/microchip-ung/arm-trusted-firmware),
included in this package, allows to boot with either of the following options:

- directly in external DDR memory as EL3 Secure mode
- directly in external DDR memory as EL1 Non-secure mode (default)

Alternatively [U-Boot](https://github.com/u-boot/u-boot), bundled as `BL33` in
a patched ATF image, can be used for chain loading.

Tusted Firmware-A
-----------------

The compiled executable should be converted to binary as follows:

```sh
aarch64-linux-gnu-objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
	-j .itablink -j .gopclntab -j .go.buildinfo -j .go.module -j .noptrdata -j .data \
	-j .bss --set-section-flags .bss=alloc,load,contents \
	-j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	main -O binary main.bin
```

A FIP image can be prepared as follows:

```sh
git clone -b v2.28.0 https://github.com/microchip-ung/mbedtls.git
git clone -b v2.8.17-mchp2 https://github.com/microchip-ung/arm-trusted-firmware
cd arm-trusted-firmware
patch -p1 < 0001-Allow-loading-TamaGo-as-BL31.patch
patch -p1 < 0002-Allow-loading-TamaGo-as-BL33.patch
make realclean

# Example of entry point extraction for TAMAGO_ENTRY_POINT value
TAMAGO_ENTRY_POINT=$(aarch64-linux-gnu-readelf -a main|grep -i 'Entry point' | awk '{print $4}')

# Compile for loading a TamaGo unikernel as BL31 (EL3, secure)
make CROSS_COMPILE=aarch64-linux-gnu- PLAT=lan969x_a0 ARCH=aarch64 \
	TAMAGO_BL31=1 TAMAGO_ENTRY_POINT=0x60072d90 TAMAGO_TEXT_START=0x60010000 \
	GENERATE_COT=1 MBEDTLS_DIR=../mbedtls KEY_ALG=ecdsa \
	NEED_BL33=no BL31=example.bin all fip

# Compile for loading a TamaGo unikernel as BL33 (EL1, non-secure)
make CROSS_COMPILE=aarch64-linux-gnu- PLAT=lan969x_a0 ARCH=aarch64 \
	TAMAGO_BL33=1 TAMAGO_ENTRY_POINT=0x60074f90 TAMAGO_TEXT_START=0x60010000 \
	GENERATE_COT=1 MBEDTLS_DIR=../mbedtls KEY_ALG=ecdsa \
	BL33=example.bin all fip
```

The FIP image can be found in  `build/lan969x_a0/release/fip.bin`.
The generated `build/lan969x_a0/release/fwu.html` can be used to flash the FIP image
in the eMMC.

U-Boot
------

Based on the patched _Trusted Firmware-A_, U-Boot can be bundled in the FIP
image as folows:

```sh
# Compile for loading U-Boot as BL33 (EL1, non-secure)
make CROSS_COMPILE=aarch64-linux-gnu- PLAT=lan969x_a0 ARCH=aarch64 \
	GENERATE_COT=1 MBEDTLS_DIR=../mbedtls KEY_ALG=ecdsa \
	BL33=u-boot.bin all fip
```

A TamaGo unikernel can then be chain loaded:

```
ext4load mmc 0:5 0x60010000 main.bin
mj $TAMAGO_ENTRY_POINT
```

Standard output
---------------

The standard output can be accessed through the debug console found on micro
USB-C connector and the following `picocom` configuration:

```sh
picocom -b 115200 -eb /dev/ttyACM0 --imap lfcrlf
```

Debugging
---------

The application can be debugged with GDB over JTAG using `openocd` (version >
0.12.0) and the following `gdbinit` debugging helper:

```
target remote localhost:3333
set remote hardware-breakpoint-limit 6
set remote hardware-watchpoint-limit 4
```

Create the file `openocd-lan969x.cfg` with the following content:

```
set _CHIPNAME lan969x
jtag newtap $_CHIPNAME cpu -irlen 4 -ircapture 0x01 -irmask 0x0f -expected-id 0x4ba00477
dap create $_CHIPNAME.dap -chain-position $_CHIPNAME.cpu
cti create $_CHIPNAME.a53_cti.0 -dap $_CHIPNAME.dap -ap-num 0 -baseaddr 0x80420000
target create $_CHIPNAME.a53.0 aarch64 -dap $_CHIPNAME.dap -cti $_CHIPNAME.a53_cti.0 -dbgbase 0x80410000
targets $_CHIPNAME.a53.0
```

Example:

```sh
# start openocd daemon
openocd -f interface/jlink.cfg -f openocd-lan969x.cfg

# connect to the OpenOCD command line
telnet localhost 4444

# debug with GDB
aarch64-linux-gnu-gdb -x gdbinit example
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
