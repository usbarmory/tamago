// Fisilink FSL91030 configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

import (
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// GPIO block; the Nuclei UX600 variant places the IOF (I/O Function)
	// registers at offsets 0x44 (enable) and 0x48 (select).
	GPIO_BASE    = 0x10011000
	GPIO_IOF_EN  = GPIO_BASE + 0x44
	GPIO_IOF_SEL = GPIO_BASE + 0x48
)

// ConfigureGPIO changes the GPIO control mode to software control (false) or
// internal hardware peripheral (true).
func ConfigureGPIO(num int, iof bool) {
	reg.Clear(GPIO_IOF_SEL, num)
	reg.SetTo(GPIO_IOF_EN, num, iof)
}
