// Nuvoton NUC980 Advanced Interrupt Controller (AIC) support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package nuc980

import (
	"github.com/usbarmory/tamago/soc/nuvoton/aic"
)

// AIC_BA is the Advanced Interrupt Controller register base.
const AIC_BA = 0xb0042000

// AIC interrupt source numbers.
const (
	IRQ_ETMR0 = 16 // Enhanced Timer 0
	IRQ_ETMR1 = 17 // Enhanced Timer 1
	IRQ_UART0 = 36 // UART 0
)

// AIC is the NUC980 Advanced Interrupt Controller.
var AIC = &aic.AIC{
	Base: AIC_BA,
}
