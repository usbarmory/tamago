TamaGo - bare metal Go for ARM SoCs - i.MX 6 support
====================================================

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

The [imx6](https://github.com/f-secure-foundry/tamago/tree/master/soc/imx6) package
provides support for NXP i.MX 6 series of System-on-Chip (SoCs) parts.

Documentation
=============

For TamaGo see its [repository](https://github.com/f-secure-foundry/tamago) and
[project wiki](https://github.com/f-secure-foundry/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/f-secure-foundry/tamago).

Supported hardware
==================

| SoC                 | Related board packages                                                                                | Peripheral drivers                                                      |
|---------------------|-------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| NXP i.MX 6UltraLite | [usbarmory/mark-two](https://github.com/f-secure-foundry/tamago/tree/master/board/f-secure/usbarmory) | DCP, GPIO, I2C, RNGB, UART, USB, USDHC                                  |
| NXP i.MX 6Quad      | none, used under QEMU for testing                                                                     | UART                                                                    |

License
=======

tamago | https://github.com/f-secure-foundry/tamago  
Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
