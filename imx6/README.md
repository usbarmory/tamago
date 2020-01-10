TamaGo - bare metal Go for ARM SoCs - i.MX 6 support
====================================================

tamago | https://github.com/inversepath/tamago  

Copyright (c) F-Secure Corporation  
https://foundry.f-secure.com

![TamaGo gopher](https://github.com/inversepath/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com

Andrej Rosano  
andrej.rosano@f-secure.com

Introduction
============

TamaGo is a project that aims to provide compilation and execution of
unencumbered Go applications for bare metal ARM System-on-Chip (SoC)
components.

The [imx6](https://github.com/inversepath/tamago/tree/master/imx6) package
provide support for NXP i.MX 6 series of System-on-Chip (SoCs) parts.

Documentation
=============

For TamaGo see its [repository](https://github.com/inversepath/tamago) and
[project wiki](https://github.com/inversepath/tamago/wiki) for information.

Supported hardware
==================

| SoC                 | Related board packages                                                                     | Peripheral drivers                                                      |
|---------------------|--------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| NXP i.MX 6UltraLite | [usbarmory/mark-two](https://github.com/inversepath/tamago/tree/master/usbarmory/mark-two) | DCP, RNGB, UART, USB                                                    |
| NXP i.MX 6Quad      | none, used under QEMU for testing                                                          | UART                                                                    |

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
