// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	_ "unsafe"
)

// NonSecure returns whether the processor security mode is non-secure (e.g.
// TrustZone Normal World.
func (cpu *CPU) NonSecure() bool {
	return scrNS == 1
}

// Secure returns whether the processor security mode is secure (e.g. TrustZone
// Secure World).
func (cpu *CPU) Secure() bool {
	return !cpu.NonSecure()
}
