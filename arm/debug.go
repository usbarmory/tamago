// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// defined in debug.s
func read_dbgauthstatus() uint32

// DebugStatus returns the contents of the ARM DBGAUTHSTATUS register, useful
// to get the current state of the processor debug permissions
// (C11.11.1, ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition).
func (cpu *CPU) DebugStatus() uint32 {
	return read_dbgauthstatus()
}
