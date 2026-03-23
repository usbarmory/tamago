// AI Foundry ET-SoC-1 emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package sys_emu

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aifoundry/etsoc1"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	etsoc1.UART0.Tx(c)
}
