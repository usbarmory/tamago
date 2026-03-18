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
* **Timer Clock**: 32768 Hz (CLINT-compatible timer at 0x2001000)
* **DRAM**: 240 MB at 0x41000000
* **NOR Flash**: Single 64 MB NOR flash (Infineon S25HL512T) at 0x20000000 XIP

Supported hardware
==================

| SoC     | Board packages | Peripheral drivers              | Status       |
|---------|----------------|---------------------------------|--------------|
| FSL91030| TBD            | UART (SiFive-compatible)        | Experimental |

Memory Map
==========

The FSL91030 memory map:

| Address       | Size    | Description                          | Interrupt(s) |
|---------------|---------|--------------------------------------|--------------|
| 0x2000000     | -       | Nuclei Timer base (mtime raw read)   | -            |
| 0x2001000     | -       | CLINT-compatible timer (Nuclei timer 0x2000000 + 0x1000); mtime at 0x200CFF8 | - |
| 0x8000000     | 64 MB   | PLIC (Platform-Level Interrupt Controller, 53 sources) | - |
| 0x10001000    | -       | DDR Controller                       | -            |
| 0x10012000    | 4 KB    | GPIO (SiFive-compatible, UART/QSPI pinmux) | -      |
| 0x10013000    | 4 KB    | UART0 (SiFive UART0, console)        | 2            |
| 0x10014000    | 4 KB    | QSPI0 (SiFive SPI0, NOR Flash ctrl)  | 5            |
| 0x10016000    | 4 KB    | QSPI1 (SiFive SPI0, NAND Flash/MMC)  | 36           |
| 0x10018000    | -       | I2C0 (OpenCores I2C)                 | -            |
| 0x10023000    | 4 KB    | UART1 (SiFive UART0, secondary)      | 3            |
| 0x20000000    | 64 MB   | NOR Flash XIP (Infineon S25HL512T)           | -    |
| 0x41000000    | 240 MB  | DRAM (0xF000000)                     | -            |
| 0x60000000    | -       | Local Bus Base                       | -            |
| 0x67800000    | 4 KB    | Ethernet MAC (FSL xy1000_eth)        | 10, 11, 12   |
| 0x68000000    | -       | Watchdog (FSL-specific)              | -            |
| 0xE084C000    | -       | System Clock Control (SYSCLK)        | -            |
| 0xE084E000    | -       | System Reset Control (SYSRST)        | -            |

Flash Layout (Reworked Board)
=============================

The reworked MilkV Vega board has a **single 64 MB NOR flash** (Infineon
S25HL512T, replacing the original dual-flash design). The single flash is
accessible at:

* **XIP window**: 0x20000000 (execute-in-place)
* **Controller**: QSPI0 at 0x10014000

Boot layout suggestion for single 64 MB flash:

```
0x20000000  Offset 0x000000  Bootloader / TamaGo image (XIP or copied to DRAM)
0x20000000  Offset 0x040000  Optional: redundant image / filesystem
```

For TamaGo bare-metal testing, load the ELF directly to DRAM at 0x41010000
via JTAG or the bootloader. See also: board package mem.go.

Peripheral Support
==================

### Fully Implemented

* **CLINT Timer**: CLINT-compatible timer at 0x2001000 (Nuclei timer base 0x2000000 + 0x1000 offset); mtime at 0x200CFF8
  * Timer frequency: 32768 Hz
  * Provides `nanotime()` for the Go runtime via unified `riscv64.CPU.GetTime()`
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
  * Register layout: DMA registers at base+0x0, MAC registers at base+0x400
  * Interrupts: PLIC 10 (RX_END), 11 (RX_REQ), 12 (TX_END)
  * Requires 8 MB DMA buffers (4 MB RX + 4 MB TX)
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
# The -T flag sets the text segment base, but the actual entry point
# (_rt0_tamago_start) is at e_entry in the ELF, which is HIGHER than -T.
# When loading via QEMU or bootloader, jump to e_entry, not to -T.
GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o example \
    main.go

# Get the actual entry point (not the text segment base):
riscv64-linux-gnu-readelf -h example | grep "Entry point"
```

### Linker Flags

* `-T 0x41010000`: Text segment address (DRAM base + 64 KB)
* `-R 0x1000`: Read-only segment alignment (4 KB)

**CRITICAL**: The ELF `e_entry` (the address of `_rt0_tamago_start`) is NOT
equal to the `-T` text segment base. The text segment base contains Go linker
metadata headers. Always extract `e_entry` from the ELF and use that as the
CPU jump target. Jumping to the text segment base (`0x41010000`) will result
in an illegal instruction trap.

Boot Sequence
=============

The FSL91030 typically boots through the following sequence:

1. **BootROM**: Initial ROM code (vendor-specific)
2. **First-stage loader**: Disables I/D cache, sets the exception vector,
   initializes the DDR controller, configures QSPI flash, re-enables cache,
   then loads secondary payloads.
3. **SBI firmware**: Provides the Supervisor Binary Interface layer.
4. **TamaGo application**: Bare metal Go runtime

If using TamaGo, the application can be loaded:
* **By boot loader**: DDR and cache already initialized; jump to ELF `e_entry`
* **As boot loader replacement**: Must implement DDR/cache initialization via `linkcpuinit` build tag

QEMU Testing
============

Use the Nuclei QEMU (nuclei/9.0 fork) with the `nuclei_evalsoc` machine:

```bash
# Build
GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o main.elf ./cmd/...

# Extract actual entry point
ENTRY=$(riscv64-linux-gnu-readelf -h main.elf | awk '/Entry point/{print $4}')

# Generate soc-cfg with correct startaddr = e_entry (NOT text base!)
cat > fsl91030.json << EOF
{
    "general_config": {
        "ddr":    { "base": "0x41000000", "size": "200M" },
        "norflash": { "base": "0x20000000", "size": "64M" },
        "uart0":  { "base": "0x10013000", "irq": "33" },
        "uart1":  { "base": "0x10023000", "irq": "34" },
        "iregion": { "base": "0x4000000" },
        "cpu_freq": "400000000",
        "timer_freq": "32768",
        "irqmax": "64"
    },
    "download": { "ddr": { "startaddr": "$ENTRY" } }
}
EOF

# Run (use -bios not -kernel for PLIC mode)
qemu-system-riscv64 \
    -M nuclei_evalsoc,download=ddr,soc-cfg=/abs/path/fsl91030.json \
    -cpu nuclei-ux600fd \
    -m 200M -smp 1 -nodefaults -nographic \
    -serial stdio \
    -bios main.elf
```

**Important notes**:
- Use `-bios` (not `-kernel`) to select PLIC+CLINT interrupt mode
- `iregion` at `0x4000000` places PLIC at `0x8000000` (matching FSL91030)
- RAM limited to 200 MB to avoid overlap with PLIC region (240 MB would extend to 0x50000000, overlapping PLIC at ~0x4F000000)
- `startaddr` must be the ELF `e_entry`, not the `-T` text segment address

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

### Working

1. **CLINT Timer**: Fully implemented at 0x2001000, frequency 32768 Hz
2. **UART Support**: Both UART0 and UART1 with correct addresses
3. **GPIO Pinmux**: UART0 pins automatically configured
4. **Memory Layout**: Correct DRAM base (0x41000000) and size (240 MB)

### Limitations

1. **Cache control stubbed**: Requires CSR manipulation (Nuclei-specific, register 0x7CA)
2. **DDR initialization stubbed**: Sequence documented but not implemented (done by boot loader)
3. **No hardware testing**: Implementation based on OpenSBI, device tree, and boot loader analysis
4. **Peripheral drivers incomplete**: Only UART and timer fully working

References
==========

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
