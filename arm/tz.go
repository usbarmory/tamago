// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// defined in tz.s
func read_scr() uint32
func write_nsacr(scr uint32)

// NonSecure returns whether the processor security mode is non-secure (e.g.
// TrustZone Normal World.
func (cpu *CPU) NonSecure() bool {
	if !cpu.security {
		return false
	}

	vecTable := cpu.vbar + 8*4
	undefinedHandler := reg.Read(vecTable + UNDEFINED)

	// NonSecure World cannot read the NS bit, the only way to infer it
	// status is to trap the exception while attempting to read it.
	reg.Write(vecTable+UNDEFINED, vector(nullHandler))
	defer reg.Write(vecTable+UNDEFINED, undefinedHandler)

	return read_scr()&1 == 1
}

// Secure returns whether the processor security mode is secure (e.g. TrustZone
// Secure World).
func (cpu *CPU) Secure() bool {
	return !cpu.NonSecure()
}

// NonSecureAccessControl sets the NSACR register value, which defines the
// Non-Secure access permissions to coprocessors.
func (cpu *CPU) NonSecureAccessControl(nsacr uint32) {
	if !cpu.security {
		return
	}

	write_nsacr(nsacr)
}
