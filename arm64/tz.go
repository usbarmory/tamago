// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

// NonSecure returns whether the CPU initialization was set as non-secure (e.g.
// TrustZone Normal World) at compile time.
func (cpu *CPU) NonSecure() bool {
	return scrNS == 1
}

// Secure returns whether the CPU initialization was set as secure (e.g.
// TrustZone Secure World) at compile time.
func (cpu *CPU) Secure() bool {
	return !cpu.NonSecure()
}
