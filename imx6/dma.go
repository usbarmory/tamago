// NXP i.MX6 DMA initialization
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"github.com/f-secure-foundry/tamago/internal/dma"
)

const iramStart uint32 = 0x00900000
const iramSize = 0x20000

func init() {
	// use internal OCRAM (iRAM) by default
	SetDMA(iramStart, iramSize)
}

// SetDMA allows relocation of DMA buffers from the default internal memory to
// other ranges (which applications must guarantee they are never used by Go
// runtime).
func SetDMA(start uint32, size int) {
	dma.Init(start, size)
}
