// Fisilink FSL91030 Watchdog Timer (Andes ATCWDT200)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

import "github.com/usbarmory/tamago/internal/reg"

// Watchdog Timer register offsets (base address: 0x68000000).
//
// The FSL91030 integrates the Andes ATCWDT200 watchdog IP, which is also
// used in VeriSilicon-based RISC-V SoCs. Register layout from the Linux
// kernel driver (fsl_wdt.c, compatible = "riscv,fsl-wdt"):
//
//	DTS node: watchdog@68000000, reg size 0x20, IRQ 8, clock hfclk (200 MHz)
const (
	WDT_IDREV   = 0x00 // ID and revision register (read-only)
	WDT_CTRL    = 0x10 // Control register
	WDT_RESTART = 0x14 // Restart register (write WDT_RESTART_KEY to feed)
	WDT_WEN     = 0x18 // Write-enable register (write WDT_UNLOCK_KEY before any write)
	WDT_ST      = 0x1C // Status register
)

// WDT control register bit fields.
const (
	WDT_CTRL_ENABLE       = 0 // Bit 0: watchdog enable
	WDT_CTRL_CLKSEL       = 1 // Bit 1: clock select (0 = hfclk/512, 1 = hfclk/1)
	WDT_CTRL_INTEN        = 2 // Bit 2: interrupt-before-reset enable
	WDT_CTRL_RSTEN        = 3 // Bit 3: system-reset-on-timeout enable
	WDT_CTRL_INTTIME      = 4 // Bits 4-6: interrupt timeout period (3-bit field)
	WDT_CTRL_RSTTIME      = 8 // Bits 8-10: reset timeout period (3-bit field)
	WDT_CTRL_INTTIME_MASK = 7 // 3-bit mask for INTTIME field
	WDT_CTRL_RSTTIME_MASK = 7 // 3-bit mask for RSTTIME field
)

// WDT magic values.
const (
	// WDT_UNLOCK_KEY must be written to WDT_WEN before writing any other
	// watchdog register. This write-protection mechanism prevents accidental
	// modification of the watchdog configuration.
	WDT_UNLOCK_KEY = 0x5aa55aa5

	// WDT_RESTART_KEY must be written to WDT_RESTART to feed (kick) the
	// watchdog and reset its counter, preventing a timeout.
	WDT_RESTART_KEY = 0xcafecafe
)

// WDT timeout period codes for WDT_CTRL INTTIME/RSTTIME fields.
// Each code N corresponds to a timeout of 2^(N+6) watchdog clock cycles.
// With clksel=0 (hfclk/512 ≈ 390 kHz), the timeout periods are:
//
//	WDT_TO_6   = 2^6  /  390kHz ≈   0.16 ms
//	WDT_TO_8   = 2^8  /  390kHz ≈   0.65 ms
//	WDT_TO_10  = 2^10 /  390kHz ≈   2.6  ms
//	WDT_TO_12  = 2^12 /  390kHz ≈  10.5  ms
//	WDT_TO_14  = 2^14 /  390kHz ≈  42    ms
//	WDT_TO_17  = 2^17 /  390kHz ≈ 335    ms
//	WDT_TO_19  = 2^19 /  390kHz ≈   1.34 s
//	WDT_TO_21  = 2^21 /  390kHz ≈   5.4  s
const (
	WDT_TO_64K  = 0 // 2^6  clocks  (~0.16 ms at hfclk/512)
	WDT_TO_256K = 1 // 2^8  clocks  (~0.65 ms)
	WDT_TO_1M   = 2 // 2^10 clocks  (~2.6  ms)
	WDT_TO_4M   = 3 // 2^12 clocks  (~10.5 ms)
	WDT_TO_16M  = 4 // 2^14 clocks  (~42   ms)
	WDT_TO_128M = 5 // 2^17 clocks  (~335  ms)
	WDT_TO_512M = 6 // 2^19 clocks  (~1.34 s)
	WDT_TO_2G   = 7 // 2^21 clocks  (~5.4  s)
)

// WDT represents the FSL91030 watchdog timer peripheral.
type WDT struct {
	// Base is the MMIO base address of the watchdog block (0x68000000).
	Base uint32
}

// unlock writes the write-enable key to WDT_WEN. This must be called before
// every write to WDT_CTRL or WDT_RESTART; the write-protection re-engages
// automatically after each protected write.
func (w *WDT) unlock() {
	reg.Write(w.Base+WDT_WEN, WDT_UNLOCK_KEY)
}

// Revision returns the ATCWDT200 IP identification and revision word from
// WDT_IDREV. This can be used to verify the peripheral is present and to
// determine the silicon revision.
func (w *WDT) Revision() uint32 {
	return reg.Read(w.Base + WDT_IDREV)
}

// Start enables the watchdog with the specified interrupt and reset timeout
// period codes (WDT_TO_* constants). A timeout first fires the interrupt
// (if INTEN is set) and then resets the system.
//
// The sequence follows the Linux fsl_wdt_start order: write RESTART key
// first (resets counter to zero), then write CTRL (enables the watchdog).
// This ensures the counter starts from a known-zero state the moment the
// watchdog is enabled, avoiding a brief window where a stale counter value
// could cause an unexpectedly early timeout.
//
//  1. Stop any running watchdog
//  2. Build CTRL value with ENABLE, RSTEN, INTEN, and the timeout codes
//  3. Unlock and write RESTART_KEY (reset the counter)
//  4. Unlock and write CTRL (enable the watchdog)
func (w *WDT) Start(intTimeout, rstTimeout int) {
	w.Stop()

	ctrl := uint32(1<<WDT_CTRL_ENABLE) |
		uint32(1<<WDT_CTRL_RSTEN) |
		uint32(1<<WDT_CTRL_INTEN) |
		uint32((intTimeout&WDT_CTRL_INTTIME_MASK)<<WDT_CTRL_INTTIME) |
		uint32((rstTimeout&WDT_CTRL_RSTTIME_MASK)<<WDT_CTRL_RSTTIME)

	w.unlock()
	reg.Write(w.Base+WDT_RESTART, WDT_RESTART_KEY)

	w.unlock()
	reg.Write(w.Base+WDT_CTRL, ctrl)
}

// Stop disables the watchdog by clearing the ENABLE bit in WDT_CTRL.
func (w *WDT) Stop() {
	ctrl := reg.Read(w.Base + WDT_CTRL)
	ctrl &^= 1 << WDT_CTRL_ENABLE
	w.unlock()
	reg.Write(w.Base+WDT_CTRL, ctrl)
}

// Feed (kick) the watchdog by writing the restart key, resetting the counter
// and preventing a timeout. Must be called periodically while the watchdog is
// running (typically more frequently than the interrupt timeout period).
func (w *WDT) Feed() {
	w.unlock()
	reg.Write(w.Base+WDT_RESTART, WDT_RESTART_KEY)
}

// ForceReset triggers an immediate system reset via the watchdog. The watchdog
// is configured with the shortest possible timeout and RSTEN enabled, then
// started. The function does not return.
func (w *WDT) ForceReset() {
	// Configure minimum timeout (WDT_TO_64K) with reset enabled
	w.Start(WDT_TO_64K, WDT_TO_64K)

	// Spin: reset fires within ~0.16 ms
	for {
	}
}

// Status returns the current value of WDT_ST (the status register). Bit 0
// is set when the interrupt-timeout event has occurred.
func (w *WDT) Status() uint32 {
	return reg.Read(w.Base + WDT_ST)
}
