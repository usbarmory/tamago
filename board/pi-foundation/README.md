TamaGo - bare metal Go for ARM SoCs - Raspberry Pi Support
==========================================================

tamago | https://github.com/f-secure-foundry/tamago

Contributors
============

Kenneth Bell

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM System-on-Chip (SoC) components.

The [pi](https://github.com/f-secure-foundry/tamago/tree/master/pi)
package provide support for the [Raspberry Pi](https://www.raspberrypi.org/)
Series of Single Board Computer.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/f-secure-foundry/tamago) and
[project wiki](https://github.com/f-secure-foundry/tamago/wiki).

For the underlying driver support for this board see package
[bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/bcm2835).

Supported hardware
==================

| SoC     | Board               | SoC package                                                          | Board package                                                                 |
|---------|---------------------|----------------------------------------------------------------------|-------------------------------------------------------------------------------|
| BCM2835 | Pi Zero             | [pi](https://github.com/f-secure-foundry/tamago/tree/master/bcm2835) | [pi/pizero](https://github.com/f-secure-foundry/tamago/tree/master/pi/pizero) |
| BCM2836 | Pi 2 Model B (v1.1) | [pi](https://github.com/f-secure-foundry/tamago/tree/master/bcm2835) | [pi/pi2](https://github.com/f-secure-foundry/tamago/tree/master/pi/pi2)       |

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support takes place:

```golang
import (
    _ "github.com/f-secure-foundry/tamago/pi/pi2"
)
```

OR

```golang
import (
    _ "github.com/f-secure-foundry/tamago/pi/pizero"
)
```

Build the [TamaGo compiler](https://github.com/f-secure-foundry/tamago-go)
(or use the [latest binary release](https://github.com/f-secure-foundry/tamago-go/releases/latest)):

```sh
git clone https://github.com/f-secure-foundry/tamago-go -b tamago1.14.6
cd tamago-go/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables and
ensuring that the required SoC and board packages are available in `GOPATH`:

```sh
GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=5 GOARCH=arm \
  ${TAMAGO} build -ldflags "-T 0x80010000  -E _rt0_arm_tamago -R 0x1000"
```

GOARM & Examples
----------------

The GOARM environment variable can be set appropriate to the model of Raspberry Pi:

| Model | GOARM | Example                                            |
|-------|-------|----------------------------------------------------|
| Zero  |   5   | <https://github.com/kenbell/tamago-example-pizero> |
| 2B    |   7   | <https://github.com/kenbell/tamago-example-pi2>    |

NOTE: The Pi Zero is ARMv6, but does not have support for all floating point instructions the Go compiler
generates when `GOARM=6`.  Using `GOARM=5` causes Go to include a software floating point implementation.

Executing and debugging
=======================

The example applications create an ELF image and use [U-Boot](https://www.denx.de/wiki/U-Boot) to bootstrap.

U-Boot
------

Configure, compile and copy U-Boot onto an existing Raspberry Pi bootable SDcard (note critical Pi firmware
is loaded from the SDcard, so a blank SDcard cannot be used).

```sh
    cd u-boot

    # Config: This is for Pi Zero, use rpi_2_defconfig for Pi 2
    make rpi_0_w_defconfig

    # Build
    make

    # Copy
    cp u-boot.bin <path_to_sdcard>
```

Configure Firmware
------------------

The Raspberry Pi firmware must be configured to use U-Boot.  Enabling the
[UART](https://www.raspberrypi.org/documentation/configuration/uart.md) is recommended to diagnose boot issues.

These settings seem to work well in `config.txt`:

```text
enable_uart=1
uart_2ndstage=1
dtparam=uart0=on
kernel=u-boot-pi0.bin
```

Executing
---------

Copy the ELF binary on top of an existing bootable Raspberry Pi microSD card an external microSD card, then launch
it from the U-Boot console as follows:

```sh
ext2load mmc 0:1 0x8000000 example
bootelf 0x8000000
```

For non-interactive execution modify the U-Boot configuration accordingly.

Standard output
---------------

The standard output can be accessed through the UART pins on the Raspberry Pi.  A 3.3v USB-to-serial cable, such as
the [Adafruit USB to TTL Serial Cable](https://www.adafruit.com/product/954) works well.  Use any suitable terminal
app to access stdout.

NOTE: Go outputs 'LF' for newline, for best results use a terminal app capable of mapping 'LF' to 'CRLF' as-needed.
