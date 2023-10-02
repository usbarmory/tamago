TamaGo - bare metal Go for ARM SoCs - BCM2835 support
=====================================================

tamago | https://github.com/usbarmory/tamago

Copyright (c) the bcm2835 package authors  

Contributors
============

[Kenneth Bell](https://github.com/kenbell)

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM/RISC-V System-on-Chip (SoC) components.

The [bcm2835](https://github.com/usbarmory/tamago/tree/master/soc/bcm2835)
package provides support for the Broadcom BCM2835 series of SoC.

Documentation
=============

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

Supported hardware
==================

| SoC              | Related board packages                                                             | Peripheral drivers |
|------------------|------------------------------------------------------------------------------------|--------------------|
| Broadcom BCM2835 | [pizero](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi/pizero) | RNG, UART, GPIO    |
| Broadcom BCM2836 | [pi2](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi/pi2)       | RNG, UART, GPIO    |

See the [pi](https://github.com/usbarmory/tamago/tree/master/board/raspberrypi) package
for documentation on compiling and executing on these boards.

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

bcm2835 | https://github.com/usbarmory/tamago/tree/master/soc/bcm2835  
Copyright (c) the bcm2835 package authors

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/soc/bcm2835/LICENSE) file.
