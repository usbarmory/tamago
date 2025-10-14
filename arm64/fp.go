// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

// defined in fp.s
func fp_enable()

// EnableFP activates floating-point operations.
func (cpu *CPU) EnableFP() {
	fp_enable()
}
