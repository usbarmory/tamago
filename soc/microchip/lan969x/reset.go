// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import (
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// CPU registers
	CPU_RESET           = CPU_BASE + 0x84
	CPU_RESET_PROT_STAT = CPU_BASE + 0x88

	// GCB registers
	GCB_SOFT_RST = GCB_BASE + 0x0c
	SOFT_SWC_RST = 1
)

// ResetTarget specifies which parts of the chip Reset resets.
type ResetTarget uint8

const (
	// Core0ColdReset stops the primary ARM core and its debug interface.
	// Execution does not resume automatically and requires external
	// recovery.
	Core0ColdReset ResetTarget = 1 << iota

	// WatchdogReset holds the watchdog in reset until a later Reset call
	// clears it.
	WatchdogReset

	// DDRReset holds DDR memory in reset and stops programs that depend on
	// it.
	DDRReset

	// L2CacheReset resets the ARM L2 cache and clears automatically when
	// complete.
	L2CacheReset

	// JTAGReset holds the ARM JTAG controller in reset until a later Reset
	// call clears it.
	JTAGReset

	// ProcessorDebugReset holds the ARM debug components in reset until a
	// later Reset call clears it.
	ProcessorDebugReset

	// Core0WarmReset restarts the primary ARM core without a full
	// processor reboot.
	Core0WarmReset

	// VCoreReset performs a full processor reboot without restarting the
	// entire chip.
	VCoreReset
)

// ResetProtection specifies which parts of the chip are protected from reset.
type ResetProtection uint8

const (
	// PCIeProtection keeps the PCIe controller running during a VCore
	// reset.
	PCIeProtection ResetProtection = 1 << iota

	// WatchdogResetStatus reports that a watchdog timeout reset VCore. It
	// is not a protection setting.
	WatchdogResetStatus

	// WatchdogSelfProtection keeps the watchdog running when its own
	// timeout resets VCore.
	WatchdogSelfProtection

	// WatchdogProtection keeps the watchdog running during VCore and
	// watchdog-triggered resets.
	WatchdogProtection

	// AMBAProtection keeps the processor interconnect running during VCore
	// and watchdog-triggered resets.
	AMBAProtection

	// VCoreProtection keeps the processor subsystem running during
	// switch-core soft reset.
	VCoreProtection
)

// Reset writes the selected targets to CPU_RESET.
func Reset(target ResetTarget) {
	reg.Write(CPU_RESET, uint32(target))
}

// SoftReset writes the reset protection and resets the switch core.
func SoftReset(protection ResetProtection) {
	reg.Write(CPU_RESET_PROT_STAT, uint32(protection))
	reg.Write(GCB_SOFT_RST, 1<<SOFT_SWC_RST)
}
