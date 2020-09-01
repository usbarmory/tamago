TamaGo - bare metal Go for ARM SoCs - BCM2835 Support
=====================================================

tamago | https://github.com/f-secure-foundry/tamago

Contributors
============

Kenneth Bell

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM System-on-Chip (SoC) components.

The [bcm2835](https://github.com/f-secure-foundry/tamago/tree/master/bcm2835)
package provides support for the Broadcom BCM2835 series of SOC.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/f-secure-foundry/tamago) and
[project wiki](https://github.com/f-secure-foundry/tamago/wiki).

Supported hardware
==================

| SoC     | Related board packages                                                        | Peripheral drivers |
|---------|-------------------------------------------------------------------------------|--------------------|
| BCM2835 | [pi/pizero](https://github.com/f-secure-foundry/tamago/tree/master/pi/pizero) | RNG, UART, GPIO    |
| BCM2836 | [pi/pi2](https://github.com/f-secure-foundry/tamago/tree/master/pi/pi2)       | RNG, UART, GPIO    |

See the [Raspberry Pi](https://github.com/f-secure-foundry/tamago/tree/master/pi) package for documentation on
compiling and executing on these boards.
