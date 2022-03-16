// BCM2835 mini-UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// This mini-UART is specifically intended for use as a
// console. See BCM2835-ARM-Peripherals.pdf that is
// widely available.
//

package bcm2835

import (
	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	AUX_ENABLES     = 0x215004
	AUX_MU_IO_REG   = 0x215040
	AUX_MU_IER_REG  = 0x215044
	AUX_MU_IIR_REG  = 0x215048
	AUX_MU_LCR_REG  = 0x21504C
	AUX_MU_MCR_REG  = 0x215050
	AUX_MU_LSR_REG  = 0x215054
	AUX_MU_MSR_REG  = 0x215058
	AUX_MU_SCRATCH  = 0x21505C
	AUX_MU_CNTL_REG = 0x215060
	AUX_MU_STAT_REG = 0x215064
	AUX_MU_BAUD_REG = 0x215068
)

type miniUART struct {
	lsr uint32
	io  uint32
}

// MiniUART is a secondary low throughput UART intended to be
// used as a console.
var MiniUART = &miniUART{}

// Init initializes the MiniUART.
func (hw *miniUART) Init() {
	reg.Write(PeripheralAddress(AUX_ENABLES), 1)
	reg.Write(PeripheralAddress(AUX_MU_IER_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_CNTL_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_LCR_REG), 3)
	reg.Write(PeripheralAddress(AUX_MU_MCR_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_IER_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_IIR_REG), 0xc6)
	reg.Write(PeripheralAddress(AUX_MU_BAUD_REG), 270)

	// Not using GPIO abstraction here because at the point
	// we initialize mini-UART during initialization, to
	// provide 'console', calling Lock on sync.Mutex fails.
	ra := reg.Read(PeripheralAddress(GPFSEL1))
	ra &= ^(uint32(7) << 12) // gpio14
	ra |= 2 << 12            // alt5
	ra &= ^(uint32(7) << 15) // gpio15
	ra |= 2 << 15            // alt5
	reg.Write(PeripheralAddress(GPFSEL1), ra)

	reg.Write(PeripheralAddress(GPPUD), 0)
	arm.Busyloop(150)

	reg.Write(PeripheralAddress(GPPUDCLK0), (1<<14)|(1<<15))
	arm.Busyloop(150)

	reg.Write(PeripheralAddress(GPPUDCLK0), 0)
	reg.Write(PeripheralAddress(AUX_MU_CNTL_REG), 3)

	hw.lsr = PeripheralAddress(AUX_MU_LSR_REG)
	hw.io = PeripheralAddress(AUX_MU_IO_REG)
}

// TX transmits a single character to the serial port.
func (hw *miniUART) Tx(c byte) {
	for {
		if reg.Read(hw.lsr)&0x20 != 0 {
			break
		}
	}

	reg.Write(hw.io, uint32(c))
}

// Write data from buffer to serial port.
func (hw *miniUART) Write(buf []byte) {
	for i := 0; i < len(buf); i++ {
		hw.Tx(buf[i])
	}
}
