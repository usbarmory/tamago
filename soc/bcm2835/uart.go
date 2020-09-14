// BCM2835 UART support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

// UART is a common interface for UARTs.
//
// The BCM2835 has 2 UARTs with very different register layouts.  This layout
// provides a common interface shared by the UARTs.
type UART interface {
	Init()
	Tx(c byte)
	Rx(c byte, valid bool)
	Write(buf []byte)
	Read(buf []byte) (n int)
}
