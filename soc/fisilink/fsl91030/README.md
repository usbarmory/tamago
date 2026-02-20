TamaGo - bare metal Go - FSL91030 support
==========================================

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

The [fsl91030](https://github.com/usbarmory/tamago/tree/master/soc/fisilink/fsl91030)
package provides support for the Fisilink FSL91030 System-on-Chip (SoC), which
is based on the Nuclei UX600 RISC-V core.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

SoC Overview
============

The FSL91030 is a RISC-V 64-bit processor based on the Nuclei UX600 core:

* **ISA**: RV64IMAFDC (64-bit, M-mode, atomics, single/double-precision FPU, compressed instructions)
* **MMU**: sv39 (39-bit virtual addressing)
* **CPU Clock**: 400 MHz (nominal, measured dynamically)
* **HF Clock**: 200 MHz (hfclk), 100 MHz (hfclk2)
* **Timer Clock**: 32768 Hz (CLINT-compatible timer)
* **DRAM**: 240 MB at 0x41000000

Supported hardware
==================

| SoC     | Board packages | Peripheral drivers              | Status       |
|---------|----------------|---------------------------------|--------------|
| FSL91030| TBD            | UART (SiFive-compatible)        | Experimental |

Memory Map
==========

The FSL91030 memory map (from OpenSBI, U-Boot, and device trees):

| Address       | Size    | Description                          | Interrupt(s) |
|---------------|---------|--------------------------------------|--------------|
| 0x2000000     | -       | Nuclei Timer (base address)          | -            |
| 0x3000000     | -       | CLINT-compatible timer (Nuclei timer + 0x1000) | -  |
| 0x8000000     | 64 MB   | PLIC (Platform-Level Interrupt Controller, 53 sources) | - |
| 0x10001000    | -       | DDR Controller                       | -            |
| 0x10012000    | 4 KB    | GPIO (SiFive-compatible, UART/QSPI pinmux) | -      |
| 0x10013000    | 4 KB    | UART0 (SiFive UART0, console)        | 2            |
| 0x10014000    | 4 KB    | QSPI0 (SiFive SPI0, NOR Flash)       | 5            |
| 0x10016000    | 4 KB    | QSPI1 (SiFive SPI0, NAND Flash/MMC)  | 36           |
| 0x10018000    | -       | I2C0 (OpenCores I2C)                 | -            |
| 0x10023000    | 4 KB    | UART1 (SiFive UART0, secondary)      | 3            |
| 0x41000000    | 240 MB  | DRAM (0xF000000)                     | -            |
| 0x60000000    | -       | Local Bus Base                       | -            |
| 0x67800000    | 4 KB    | Ethernet MAC (FSL xy1000_eth)        | 10, 11, 12   |
| 0x68000000    | -       | Watchdog (FSL-specific)              | -            |
| 0xE084C000    | -       | System Clock Control (SYSCLK)        | -            |
| 0xE084E000    | -       | System Reset Control (SYSRST)        | -            |

Peripheral Support
==================

### Fully Implemented

* **CLINT Timer**: CLINT-compatible timer at 0x3000000 (Nuclei timer + 0x1000 offset)
  * Timer frequency: 32768 Hz
  * Provides `nanotime1()` for Go runtime
  * Based on OpenSBI platform.c implementation
* **UART0/UART1**: Uses `soc/sifive/uart` driver (SiFive UART0 compatible)
  * UART0: 0x10013000 (console)
  * UART1: 0x10023000 (secondary)
* **GPIO Pinmux**: UART0 pin configuration via GPIO IOF registers
  * Automatically initialized in `Init()` function
  * Required for UART0 to function properly
* **RNG**: Time-based DRBG initialization (requires hardware entropy source override)
* **Clock**: Fixed frequency reporting (400 MHz CPU, 200/100 MHz HF clocks, 32768 Hz timer)
* **Memory**: DRAM at 0x41000000 (240 MB)

### Stub Implementations (TODO)

* **DDR Controller**: Register map documented; initialization sequence from boot loader
* **Cache Control**: Nuclei-specific CSR operations (0x7CA)
* **GPIO Full Driver**: Only pinmux implemented; full GPIO driver needed for general I/O
* **QSPI Flash Controllers**:
  * QSPI0 (0x10014000): NOR Flash - Macronix MX25U51245G, SiFive SPI0-compatible
  * QSPI1 (0x10016000): NAND Flash/MMC for SD card boot, requires GPIO pinmux
* **I2C**: OpenCores I2C-compatible at 0x10018000
* **Ethernet MAC** (0x67800000): FSL xy1000_eth driver
  * Register structure: DMA at base+0x0, MAC at base+0x400
  * Interrupts: 10 (RX_END), 11 (RX_REQ), 12 (TX_END)
  * Requires 8MB DMA buffers (4MB RX + 4MB TX)
  * Reference: vega-u-boot/drivers/net/xy1000_eth.{c,h}
* **Watchdog**: FSL-specific at 0x68000000
* **System Control**:
  * SYSCLK (0xE084C000): Clock control for peripherals
  * SYSRST (0xE084E000): Reset control for peripherals
* **PLIC**: Interrupt controller at 0x8000000 (53 interrupt sources, deferred to board package)

Compilation Example
===================

```bash
# Set up TamaGo compiler (see main TamaGo README)
export TAMAGO=/path/to/tamago-go/bin/go

# Compile for FSL91030 (text segment at DRAM start + 64KB)
GOOS=tamago GOARCH=riscv64 ${TAMAGO} build \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o example \
    main.go
```

### Linker Flags

* `-T 0x41010000`: Text segment address (DRAM base + 64 KB)
* `-R 0x1000`: Read-only segment alignment (4 KB)

Adjust text segment address based on boot loader requirements.

Boot Sequence
=============

The FSL91030 typically boots through the following sequence:

1. **BootROM**: Initial ROM code (vendor-specific)
2. **First-stage loader**: `freeloader.S` (vega-loader-entire)
   * Disables I/D cache
   * Sets exception vector
   * Initializes DDR controller
   * Configures QSPI flash
   * Enables I/D cache
   * Loads payloads (OpenSBI, U-Boot, kernel)
3. **OpenSBI**: Provides SBI firmware
4. **TamaGo application**: Bare metal Go runtime

If using TamaGo, the application can be loaded:
* **By boot loader**: DDR and cache already initialized
* **As boot loader replacement**: Must implement DDR/cache initialization

Build Tags
==========

The following build tags allow applications to override the package's own
definitions of [external functions required by the runtime](https://pkg.go.dev/github.com/usbarmory/tamago/doc):

* `linkramstart`: Exclude `ramStart` from `mem.go`
* `linkramsize`: Exclude `ramSize` from `mem.go`
* `linkprintk`: Exclude `printk` implementation
* `linkcpuinit`: Exclude `cpuinit` imported from `riscv64/init.s`

Limitations
===========

This is an **experimental** implementation with the following status:

### Working (Based on OpenSBI Reference)

1. **CLINT Timer**: Fully implemented at 0x3000000, frequency 32768 Hz
2. **UART Support**: Both UART0 and UART1 with correct addresses
3. **GPIO Pinmux**: UART0 pins automatically configured
4. **Memory Layout**: Correct DRAM base (0x41000000) and size (240 MB)

### Limitations

1. **Cache control stubbed**: Requires CSR manipulation (Nuclei-specific, register 0x7CA)
2. **DDR initialization stubbed**: Sequence documented but not implemented (done by boot loader)
3. **No hardware testing**: Implementation based on OpenSBI, device tree, and boot loader analysis
4. **Peripheral drivers incomplete**: Only UART and timer fully working

### Implementation Sources

All register addresses and initialization sequences are derived from:
* **OpenSBI**: `vega-opensbi/platform/nuclei/ux600/platform.c` (primary reference for timer, UART, GPIO)
* **U-Boot**: `vega-u-boot/arch/riscv/dts/nuclei-ux608.dts` and driver source (ethernet, QSPI details)
* **Device trees**:
  * `vega-buildroot-sdk/conf/nuclei_ux600fd.dts`
  * `vega-u-boot/arch/riscv/dts/nuclei-ux608.dts`
* **Boot loader**: `vega-loader-entire/freeloader.S` (DDR initialization, cache control)
* **Ethernet driver**: `vega-u-boot/drivers/net/xy1000_eth.{c,h}` (MAC register structure, system control)

The implementation follows OpenSBI's tested patterns, particularly for:
* Timer address (0x3000000) and frequency (32768 Hz)
* UART addresses (0x10013000, 0x10023000) and pinmux configuration
* GPIO IOF register configuration for UART0
* Early initialization sequence

References
==========

### Primary Implementation Sources

* **OpenSBI**: `vega-opensbi/platform/nuclei/ux600/platform.c` (timer, UART, GPIO initialization)
* **U-Boot**: `vega-u-boot/` (device tree, board code, ethernet driver)
  * Device tree: `arch/riscv/dts/nuclei-ux608.dts`
  * Board file: `board/nuclei/ux608/ux608.c`
  * Ethernet: `drivers/net/xy1000_eth.{c,h}`
* **Device tree**: `vega-buildroot-sdk/conf/nuclei_ux600fd.dts`
* **Boot loader**: `vega-loader-entire/freeloader.S` (DDR init, cache control)

### Datasheets and Documentation

* **FSL91030 Datasheets**: `vega-docs/development-documentation/`
  * FSL91030M芯片数据手册-G版本.pdf (Datasheet)
  * FSL91030M寄存器说明书-D.pdf (Register Manual)
  * FSL91030(M)芯片SoC使用说明书_V10.pdf (SoC User Manual)
  * FSL91030(M)芯片原理文档_V12.pdf (Architecture Document)
* **Hardware Schematics**: `vega-docs/hardware/`
  * vega_schematic_v1.1.pdf (Board schematic)
  * vega-mechanical-drawing.pdf (Mechanical drawing)

### External References

* [Nuclei UX600 Documentation](https://doc.nucleisys.com/)
* [SiFive FU540 Manual](https://sifive.cdn.prismic.io/sifive/b5e7a29c-d3c2-44ea-85fb-acc1df282e21_FU540-C000-v1.4.pdf) (for compatible peripherals)
* [OpenSBI Documentation](https://github.com/riscv-software-src/opensbi)
* [RISC-V Specifications](https://riscv.org/technical/specifications/)

License
=======

tamago | https://github.com/usbarmory/tamago
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
