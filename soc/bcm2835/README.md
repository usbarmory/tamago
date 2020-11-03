TamaGo - bare metal Go for ARM SoCs - BCM2835 support
=====================================================

tamago | https://github.com/f-secure-foundry/tamago

Copyright (c) the bcm2835 package authors  

Contributors
============

[Kenneth Bell](https://github.com/kenbell)

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM System-on-Chip (SoC) components.

The [bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/soc/bcm2835)
package provides support for the Broadcom BCM2835 series of SoC.

Documentation
=============

For TamaGo see its [repository](https://github.com/f-secure-foundry/tamago) and
[project wiki](https://github.com/f-secure-foundry/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/f-secure-foundry/tamago).

Supported hardware
==================

| SoC     | Related board packages                                                                    | Peripheral drivers |
|---------|-------------------------------------------------------------------------------------------|--------------------|
| BCM2835 | [pizero](https://github.com/f-secure-foundry/tamago/tree/master/board/raspberrypi/pizero) | RNG, UART, GPIO    |
| BCM2836 | [pi2](https://github.com/f-secure-foundry/tamago/tree/master/board/raspberrypi/pi2)       | RNG, UART, GPIO    |

See the [pi](https://github.com/f-secure-foundry/tamago/tree/master/pi) package
for documentation on compiling and executing on these boards.

License
=======

tamago | https://github.com/f-secure-foundry/tamago  
Copyright (c) F-Secure Corporation

bcm2835 | https://github.com/f-secure-foundry/tamago/tree/master/soc/bcm2835  
Copyright (c) the bcm2835 package authors

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
