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
* **CPU Clock**: 400 MHz (nominal)
* **HF Clock**: 200 MHz (hfclk), 100 MHz (hfclk2)
* **Timer Clock**: 32768 Hz (CLINT-compatible timer at 0x2001000)
* **DRAM**: 240 MB at 0x41000000
* **NOR Flash**: Single 64 MB NOR flash (Infineon S25HL512T) at 0x20000000 XIP

Supported hardware
==================

| SoC      | Board packages      | Peripheral drivers                                    | Status |
|----------|---------------------|-------------------------------------------------------|--------|
| FSL91030 | milkv/vega          | UART, CLINT timer, GPIO pinmux, cache control, DDR, WDT, SYSCLK/SYSRST | Active |

Memory Map
==========

| Address       | Size    | Description                                          | Interrupt(s) |
|---------------|---------|------------------------------------------------------|--------------|
| 0x2000000     | -       | Nuclei Timer base (mtime raw read)                   | -            |
| 0x2001000     | -       | CLINT-compatible timer (base + 0x1000); mtime at 0x200CFF8 | -      |
| 0x8000000     | 64 MB   | PLIC (Platform-Level Interrupt Controller, 53 sources) | -          |
| 0x10001000    | -       | DDR Controller                                       | -            |
| 0x10011000    | 4 KB    | GPIO (SiFive-compatible, UART/QSPI pinmux)           | 1            |
| 0x10013000    | 4 KB    | UART0 (SiFive UART0, console)                        | 2            |
| 0x10014000    | 4 KB    | QSPI0 (SiFive SPI0, NOR Flash ctrl)                  | 5            |
| 0x10016000    | 4 KB    | QSPI1 (SiFive SPI0, NAND Flash/MMC)                  | 36           |
| 0x10018000    | -       | I2C0 (OpenCores I2C)                                 | -            |
| 0x10023000    | 4 KB    | UART1 (SiFive UART0, secondary)                      | 3            |
| 0x20000000    | 64 MB   | NOR Flash XIP (Infineon S25HL512T)                   | -            |
| 0x40000000    | 32 KB   | Uncached DMA SRAM (16 KB RX + 16 KB TX)              | -            |
| 0x41000000    | 240 MB  | DRAM                                                 | -            |
| 0x60000000    | -       | Local Bus Base                                       | -            |
| 0x67800000    | 4 KB    | Ethernet MAC (FSL xy1000_eth)                        | 10, 11, 12   |
| 0x68000000    | 32 B    | Watchdog (Andes ATCWDT200)                           | 8            |
| 0xE084C000    | -       | System Clock Control (SYSCLK)                        | -            |
| 0xE084E000    | -       | System Reset Control (SYSRST)                        | -            |

Flash Layout (64 MB / Infineon S25HL512T)
==========================================

The MilkV Vega board (reworked) has a **single 64 MB NOR flash** (Infineon
S25HL512T) accessible at XIP address 0x20000000. The flash layout is:

```
0x20000000  Offset 0x00000000  64 KB     flashboot stub (DDR init + copy + jump)
0x20000008  Offset 0x00000008  8 B       ELF e_entry (LE64, patched at build time)
0x20000010  Offset 0x00000010  8 B       Binary size (LE64, patched at build time)
0x20010000  Offset 0x00010000  ~60 MB    TamaGo runtime binary
0x23C00000  Offset 0x03C00000  2 MB      Config partition (log-structured KV store)
0x23E00000  Offset 0x03E00000  1 MB      OTA staging area
0x23F00000  Offset 0x03F00000  1 MB      Reserved
```

**Note**: The S25HL512T requires 4-byte addressing for offsets beyond 16 MB
(enter via QSPI command 0xB7). The flashboot stub handles this automatically.

Boot Sequence
=============

TamaGo is the **sole software** on this chip. The complete boot flow is:

1. **BootROM** (vendor-supplied, on-chip): Starts execution from NOR flash XIP
   base 0x20000000.

2. **flashboot stub** (`tools/flashboot.s`, placed at flash offset 0):
   - Disables the Nuclei UX600 I/D cache (CSR 0x7CA)
   - Initializes the DDR SDRAM controller (register values from vendor
     freeloader.S reference, ported to standalone RISC-V assembly)
   - Configures the QSPI0 clock divider
   - Copies the TamaGo binary from flash offset 0x10000 to DRAM 0x41010000
   - Re-enables I/D cache
   - Reads `e_entry` from flash header at offset 0x8 and jumps there

3. **TamaGo runtime**: Bare metal Go application executes in M-mode.
   Board `Init()` (Hwinit1) runs after the Go scheduler is up.

There is no SBI firmware, no OpenSBI, no U-Boot, and no separate bootloader.
The flashboot stub and the TamaGo binary together constitute the full
firmware image.

**Alternative: linkcpuinit build tag**

Building with `-tags linkcpuinit` embeds DDR init directly into the TamaGo
binary (via `boot_riscv64.s`), which then overrides the default `cpuinit`.
This is useful for XIP boot (running directly from flash without the copy
step) or for testing without the flashboot stub. The standard
`vega-baremetal` production build does NOT use `linkcpuinit`; it relies on
`flashboot.s` for hardware init.

Peripheral Support
==================

### Fully Implemented

* **CLINT Timer**: CLINT-compatible timer at 0x2001000; mtime at 0x200CFF8
  * Frequency: 32768 Hz
  * Provides `nanotime()` for the Go runtime
* **UART0 / UART1**: SiFive UART0-compatible at 0x10013000 / 0x10023000
* **GPIO Pinmux**: UART0 pin configuration via GPIO IOF registers at 0x10011000
  * Nuclei UX600 IOF register offsets: IOF_EN at 0x44, IOF_SEL at 0x48
  * (These differ from standard SiFive FE310/FU540 offsets of 0x38/0x3C)
* **Cache Control**: Nuclei UX600 custom CSR_MCACHE_CTL (0x7CA), I+D cache
  enable/disable via `EnableCache()` / `DisableCache()`
* **DDR Init**: Full initialization sequence in three forms:
  * `tools/flashboot.s` — standalone stub (standard production path)
  * `boot_riscv64.s` — embedded in TamaGo binary (`linkcpuinit` tag)
  * `InitDDR()` — Go-callable function (available after runtime start)
* **Watchdog**: Andes ATCWDT200 at 0x68000000 — `WDT.Start()`, `WDT.Feed()`,
  `WDT.Stop()`, `WDT.ForceReset()`
* **System Control**: SYSCLK (0xE084C000) and SYSRST (0xE084E000) — peripheral
  clock gating and reset sequencing via `SysCtl`
* **RNG**: Timer-seeded DRBG (override with `SetRNG()` for hardware entropy)
* **Clock constants**: 400 MHz CPU, 200/100 MHz HF, 32768 Hz timer

### Pending

* **GPIO full driver**: SiFive GPIO0 driver for general-purpose I/O (read,
  write, interrupts). IOF configuration is already working for UART0.
  Blocked on `soc/sifive/gpio/` package creation.
* **QSPI0 / QSPI1**: SiFive SPI0-compatible flash controllers. XIP reads
  work via hardware. Command mode (erase/write) not yet implemented.
* **I2C0**: OpenCores I2C at 0x10018000. Not yet needed by any consumer.
* **Ethernet MAC**: FSL xy1000_eth at 0x67800000. Implemented in the
  vega-baremetal application layer (`pkg/hal/eth/`).

Compilation
===========

```bash
# Set up TamaGo compiler
export TAMAGO=/path/to/tamago-go/bin/go

# Build for FSL91030 hardware (text segment at DRAM base + 64 KB)
GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o example \
    main.go

# Get the actual entry point (always use e_entry, not the text segment base):
riscv64-linux-gnu-readelf -h example | grep "Entry point"
```

### Linker Flags

* `-T 0x41010000`: Text segment at DRAM base 0x41000000 + 64 KB
* `-R 0x1000`: Read-only segment alignment (4 KB)

**CRITICAL**: The ELF `e_entry` (`_rt0_tamago_start`) is NOT equal to the
`-T` text segment base. The Go linker places metadata before the entry point.
Always extract `e_entry` from the ELF header and use that as the CPU jump
target. The flashboot stub and `make flash` / `make qemu` handle this
automatically.

QEMU Testing
============

Use the Nuclei QEMU (nuclei/9.0 fork) with the `nuclei_evalsoc` machine:

```bash
# Build
GOOS=tamago GOARCH=riscv64 GOOSPKG=github.com/usbarmory/tamago ${TAMAGO} build \
    -ldflags "-T 0x41010000 -R 0x1000" \
    -o main.elf ./cmd/...

# Extract actual entry point (not the text segment base)
ENTRY=$(riscv64-linux-gnu-readelf -h main.elf | awk '/Entry point/{print $4}')

# Generate QEMU soc-cfg with correct startaddr = e_entry
cat > fsl91030.json << EOF
{
    "general_config": {
        "ddr":      { "base": "0x41000000", "size": "200M" },
        "norflash": { "base": "0x20000000", "size": "64M"  },
        "uart0":    { "base": "0x10013000", "irq": "33"    },
        "uart1":    { "base": "0x10023000", "irq": "34"    },
        "iregion":  { "base": "0x4000000" },
        "cpu_freq":   "400000000",
        "timer_freq": "32768",
        "irqmax":     "64"
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

**Notes**:
- Use `-bios` (not `-kernel`) to select PLIC+CLINT interrupt mode
- `iregion` at `0x4000000` places PLIC at `0x8000000` (matching FSL91030)
- RAM limited to 200 MB to avoid PLIC address overlap (240 MB would extend
  to 0x50000000, overlapping PLIC at ~0x4F000000)
- `startaddr` must be the ELF `e_entry`, not the `-T` text segment base
- Under QEMU, DDR is always ready; no flashboot stub or linkcpuinit needed

Build Tags
==========

| Build Tag      | Effect                                                          |
|----------------|-----------------------------------------------------------------|
| `linkramstart` | Exclude `ramStart`; provide your own via `go:linkname`          |
| `linkramsize`  | Exclude `ramSize`; provide your own via `go:linkname`           |
| `linkprintk`   | Exclude `printk`; provide your own via `go:linkname`            |
| `linkcpuinit`  | Override `cpuinit` with full DDR/cache/QSPI init (boot_riscv64.s) |

Limitations
===========

1. **No hardware entropy source**: The FSL91030 has no documented hardware
   RNG. The DRBG seeded from the CLINT timer (`initRNG`) is unsuitable for
   cryptographic use. Override with `SetRNG()`.
2. **QSPI command mode**: XIP (read-only) works. Erase/write requires
   SiFive SPI0 command mode driver (not yet implemented).
3. **GPIO general-purpose I/O**: IOF pinmux works. Full GPIO read/write/
   interrupt driver pending `soc/sifive/gpio/` implementation.
4. **DRAM size discrepancy**: The vendor DTS uses 240 MB; the board package
   (`milkv/vega`) uses 256 MB (physical chip capacity). The SoC package
   conservatively reports 240 MB to match the DTS. Board package overrides
   via `linkramsize`.
5. **Switch pipeline registers**: The FSL91030 contains an integrated L2
   Ethernet switch. Pipeline register formats are partially reverse-engineered;
   hardware VLAN/ACL/QoS pushes are pending.

References
==========

* [vega-baremetal project](https://github.com/mdr164/vega-baremetal) — application using this SoC package
* [Nuclei UX600 Documentation](https://doc.nucleisys.com/)
* [SiFive FU540 Manual](https://sifive.cdn.prismic.io/sifive/b5e7a29c-d3c2-44ea-85fb-acc1df282e21_FU540-C000-v1.4.pdf) (for SiFive-compatible peripherals)
* [Infineon S25HL512T datasheet](https://www.infineon.com/) (NOR flash)
* [Andes ATCWDT200 IP](https://www.andestech.com/) (Watchdog)
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
