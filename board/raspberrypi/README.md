TamaGo - bare metal Go for ARM SoCs - Raspberry Pi Support
==========================================================

tamago | https://github.com/usbarmory/tamago  

Copyright (c) the pi/pi2/pizero package authors  

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Contributors
============

[Kenneth Bell](https://github.com/kenbell)

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM/RISC-V System-on-Chip (SoC) components.

The [pi](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi)
package provides support for the [Raspberry Pi](https://www.raspberrypi.org/)
series of Single Board Computer.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For the underlying driver support for this board see package
[bcm2835](https://github.com/usbarmory/tamago/tree/master/bcm2835).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC              | Board                | SoC package                                                            | Board package                                                                         |
|------------------|----------------------|------------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| Broadcom BCM2835 | Pi Zero              | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835) | [pi/pizero](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi/pizero) |
| Broadcom BCM2835 | Pi 1 Model A+ (v1.2) | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835) | [pi/pi1](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi/pi1)       |
| Broadcom BCM2835 | Pi 1 Model B+ (v1.2) | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835) | [pi/pi1](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi/pi1)       |
| Broadcom BCM2836 | Pi 2 Model B (v1.1)  | [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835) | [pi/pi2](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi/pi2)       |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
    _ "github.com/usbarmory/tamago/board/raspberrypi/pi2"
)
```

OR

```golang
import (
    _ "github.com/usbarmory/tamago/board/raspberrypi/pi1"
)
```

OR

```golang
import (
    _ "github.com/usbarmory/tamago/board/raspberrypi/pizero"
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
previous step, but with the addition of the following flags/variables and
ensuring that the required SoC and board packages are available in `GOPATH`:

```sh
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=5 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x00010000 -E _rt0_arm_tamago -R 0x1000"
```

GOARM & Examples
----------------

The GOARM environment variable must be set according to the Raspberry Pi model:

| Model | GOARM | Example                                                     |
|-------|-------|-------------------------------------------------------------|
| Zero  |   5   | <https://github.com/usbarmory/tamago-example-pizero>        |
| 1A+   |   5   | <https://github.com/prusnak/tamago-example-pi1>             |
| 1B+   |   5   | <https://github.com/prusnak/tamago-example-pi1>             |
| 2B    |   7   | <https://github.com/kenbell/tamago-example-pi2>             |

NOTE: The Pi Zero and Pi 1 are ARMv6, but do not have support for all floating point instructions the Go compiler
generates with `GOARM=6`.  Using `GOARM=5` causes Go to include a software floating point implementation.

Executing
=========

Compiled binaries can be executed by converting Go binaries to emulate the
Linux boot protocol and have the Pi firmware load and execute the binary as a
Linux kernel.

A minimal set of Raspberry Pi firmware must be present on the SD card to
initialize the Raspberry Pi using the VideoCore GPU. The following minimum
files are required:

* bootcode.bin
* fixup.dat
* start.elf

These files are available [here](https://github.com/raspberrypi/firmware/tree/master/boot).

Direct
------

Linux kernels are expected to have executable code as the first bytes of the binary. The Go compiler
does not natively support creating such binaries, so a stub is generated and pre-pended that will jump
to the Go entrypoint. In this way, the Linux boot protocol is satisfied.

The example projects (linked above) use the direct approach. The GNU cross-compiler toolchain is
required. This method is in some ways more complex, but the Makefile code from the examples can be
used as an example implementation.

1. Build the Go ELF binary as normal
2. Use `objcopy` from the GNU cross-compiler toolchain to convert the binary to 'bin' format
3. Extract the entrypoint from the ELF format file
4. Compile a stub that will jump to the real entrypoint
5. Prepend the stub with sufficient padding for alignment
6. Configure the Pi to treat the binary as the Linux kernel to load

In the examples, this code performs steps 1-5:

```sh
$(CROSS_COMPILE)objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
    -j .bss --set-section-flags .bss=alloc,load,contents \
    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents\
    $(APP) -O binary $(APP).o
${CROSS_COMPILE}gcc -D ENTRY_POINT=`${CROSS_COMPILE}readelf -e $(APP) | grep Entry | sed 's/.*\(0x[a-zA-Z0-9]*\).*/\1/'` -c boot.S -o boot.o
${CROSS_COMPILE}objcopy boot.o -O binary stub.o
# Truncate pads the stub out to correctly align the binary
# 32768 = 0x10000 (TEXT_START) - 0x8000 (Default kernel load address)
truncate -s 32768 stub.o
cat stub.o $(APP).o > $(APP).bin
```

The bootstrap code is something equivalent to this:

```S
    .global _boot

    .text
_boot:
    LDR r1, addr
    BX r1

addr:
    .word ENTRY_POINT
```

Direct: Configuring the firmware
--------------------------------

An example config.txt is:

```txt
enable_uart=1
uart_2ndstage=1
dtparam=uart0=on
kernel=example.bin
kernel_address=0x8000
disable_commandline_tags=1
core_freq=250
```

See <http://rpf.io/configtxt> for more configuration options.

NOTE: Do not be tempted to set the kernel address to 0x0:

1. TamaGo places critical data-structures at RAMSTART
2. The Pi firmware parks all but 1 CPU core in wait-loops, controlled by bytes starting at 0x000000CC
(see <https://github.com/raspberrypi/tools/blob/master/armstubs/armstub7.S>)

Direct: Executing
-----------------

Copy the binary and config.txt to an SD card alongside the Pi firmware binaries and power-up the Pi.

Debugging: Standard output
==========================

The standard output can be accessed through the UART pins on the Raspberry Pi.
A 3.3v USB-to-serial cable, such as the [Adafruit USB to TTL Serial Cable](https://www.adafruit.com/product/954)
can be used. Any suitable terminal emulator can be used to access standard output.

The UART clock is based on the VPU clock in some Pi models, if the UART output
appears corrupted, ensure the VPU clock frequency is fixed using `core_freq=250`
in `config.txt`.

NOTE: Go outputs 'LF' for newline, for best results use a terminal app capable
of mapping 'LF' to 'CRLF' as-needed.

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

raspberrypi | https://github.com/usbarmory/tamago/tree/master/board/raspberrypi  
Copyright (c) the pi package authors

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/board/raspberrypi/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
