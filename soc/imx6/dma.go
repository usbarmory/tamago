// NXP i.MX6 DMA initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"github.com/usbarmory/tamago/dma"
)

const (
	IRAMStart uint32 = 0x00900000
	IRAMSize         = 0x20000
)

func init() {
	// use internal OCRAM (iRAM) by default
	dma.Init(IRAMStart, IRAMSize)
}
