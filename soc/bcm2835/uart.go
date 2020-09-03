// BCM2835 mini-UART
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//
// This mini-UART is specifically intended for use as a
// console.  See BCM2835-ARM-Peripherals.pdf that is
// widely available.
//

package bcm2835

import (
	// uses go:linkname
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

//go:linkname printk runtime.printk
func printk(c byte) {
	uartPutc(uint32(c))
}

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

	GPFSEL1   = 0x200004
	GPPUD     = 0x200094
	GPPUDCLK0 = 0x200098
)

func uartInit() {
	reg.Write(PeripheralAddress(AUX_ENABLES), 1)
	reg.Write(PeripheralAddress(AUX_MU_IER_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_CNTL_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_LCR_REG), 3)
	reg.Write(PeripheralAddress(AUX_MU_MCR_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_IER_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_IIR_REG), 0xC6)
	reg.Write(PeripheralAddress(AUX_MU_BAUD_REG), 270)

	ra := reg.Read(PeripheralAddress(GPFSEL1))
	ra &= ^(uint32(7) << 12) // gpio14
	ra |= 2 << 12            // alt5
	ra &= ^(uint32(7) << 15) // gpio15
	ra |= 2 << 15            // alt5
	reg.Write(PeripheralAddress(GPFSEL1), ra)

	reg.Write(PeripheralAddress(GPPUD), 0)
	delay(150)
	reg.Write(PeripheralAddress(GPPUDCLK0), (1<<14)|(1<<15))
	delay(150)
	reg.Write(PeripheralAddress(GPPUDCLK0), 0)

	reg.Write(PeripheralAddress(AUX_MU_CNTL_REG), 3)
}

func uartPutc(c uint32) {
	for {
		if (reg.Read(PeripheralAddress(AUX_MU_LSR_REG)) & 0x20) != 0 {
			break
		}
	}

	reg.Write(PeripheralAddress(AUX_MU_IO_REG), c)
}

func delay(c int) {
	for i := 0; i < c; i++ {
		dummy()
	}
}

//go:noinline
func dummy() {}
