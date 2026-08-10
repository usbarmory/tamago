// Microchip FLEXCOM Two-wire Interface support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package flexcom

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// TWI registers
const (
	FLEX_TWI_OFFSET = 0x600

	FLEX_TWI_CR  = 0x00
	TWI_CR_START = 0
	TWI_CR_STOP  = 1
	TWI_CR_MSEN  = 2
	TWI_CR_SWRST = 7

	FLEX_TWI_MMR        = 0x04
	TWI_MMR_IADRSZ      = 8
	TWI_MMR_IADRSZ_MASK = 0x3
	TWI_MMR_MREAD       = 12
	TWI_MMR_DADR        = 16
	TWI_MMR_DADR_MASK   = 0x7f

	FLEX_TWI_IADR = 0x0c

	FLEX_TWI_CWGR       = 0x10
	TWI_CWGR_CLDIV      = 0
	TWI_CWGR_CLDIV_MASK = 0xff
	TWI_CWGR_CHDIV      = 8
	TWI_CWGR_CHDIV_MASK = 0xff
	TWI_CWGR_CKDIV      = 16
	TWI_CWGR_CKDIV_MASK = 0x7
	TWI_CWGR_GCK        = 20

	FLEX_TWI_SR   = 0x20
	TWI_SR_TXCOMP = 0
	TWI_SR_RXRDY  = 1
	TWI_SR_TXRDY  = 2
	TWI_SR_NACK   = 8

	FLEX_TWI_RHR = 0x30
	FLEX_TWI_THR = 0x34

	TWI_IADRSZ_MAX = 3
)

// TWITimeout is the default timeout for TWI operations.
const TWITimeout = 100 * time.Millisecond

// TWI represents the FLEXCOM TWI mode instance.
type TWI struct {
	sync.Mutex

	// Base register
	Base uint32

	// ClockLowDivider configures the TWI low-period divider.
	ClockLowDivider uint8

	// ClockHighDivider configures the TWI high-period divider.
	ClockHighDivider uint8

	// ClockDivider applies a 2^n scale to both TWI clock periods.
	ClockDivider uint8

	// GenericClock selects GCLK instead of the peripheral clock.
	GenericClock bool

	// Timeout bounds each wait for TWI controller progress. Zero selects 100 ms.
	Timeout time.Duration

	clock uint32
}

func (hw *TWI) reset() {
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_SWRST)
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CWGR, hw.clock)
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_MSEN)
}

func (hw *TWI) init() (err error) {
	if hw.ClockDivider > TWI_CWGR_CKDIV_MASK {
		return fmt.Errorf("invalid TWI clock divider %d", hw.ClockDivider)
	}

	if hw.Timeout <= 0 {
		hw.Timeout = TWITimeout
	}

	bits.SetN(&hw.clock, TWI_CWGR_CLDIV, TWI_CWGR_CLDIV_MASK, uint32(hw.ClockLowDivider))
	bits.SetN(&hw.clock, TWI_CWGR_CHDIV, TWI_CWGR_CHDIV_MASK, uint32(hw.ClockHighDivider))
	bits.SetN(&hw.clock, TWI_CWGR_CKDIV, TWI_CWGR_CKDIV_MASK, uint32(hw.ClockDivider))
	bits.SetTo(&hw.clock, TWI_CWGR_GCK, hw.GenericClock)

	hw.reset()

	return
}

func (hw *TWI) configure(target uint8, addr uint32, alen int, read bool) (err error) {
	switch {
	case hw.Base == 0:
		return errors.New("invalid FLEXCOM controller instance")
	case target > TWI_MMR_DADR_MASK:
		return fmt.Errorf("invalid device address %#x", target)
	case alen < 0 || alen > TWI_IADRSZ_MAX:
		return fmt.Errorf("invalid internal device address size %d", alen)
	case alen == 0 && addr != 0,
		alen > 0 && addr >= 1<<uint(alen*8):
		return fmt.Errorf("invalid internal device address %#x", addr)
	}

	var mode uint32
	bits.SetN(&mode, TWI_MMR_IADRSZ, TWI_MMR_IADRSZ_MASK, uint32(alen))
	bits.SetTo(&mode, TWI_MMR_MREAD, read)
	bits.SetN(&mode, TWI_MMR_DADR, TWI_MMR_DADR_MASK, uint32(target))
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_MMR, mode)
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_IADR, addr)

	return
}

func (hw *TWI) wait(bit int) (err error) {
	start := time.Now()

	for {
		status := reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_SR)

		switch {
		case bits.Get(&status, TWI_SR_NACK):
			hw.reset()
			return errors.New("target did not acknowledge")
		case bits.Get(&status, bit):
			return nil
		case time.Since(start) >= hw.Timeout:
			hw.reset()
			return fmt.Errorf("timeout waiting for status bit %d (%#08x)", bit, status)
		}

		runtime.Gosched()
	}
}

// Read reads a sequence of bytes from a target device.
//
// The address length (`alen`) parameter should be set greater than 0 for
// ordinary I2C reads (`TARGET W|ADDR|TARGET R|DATA`) and equal to 0 when not
// sending a register address (`TARGET R|DATA`), values less than 0 or greater
// than 3 are not valid.
//
// The buffer is filled to its full length, on error it may be partially
// filled.
func (hw *TWI) Read(target uint8, addr uint32, alen int, buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if len(buf) == 0 {
		return
	}

	if err = hw.configure(target, addr, alen, true); err != nil {
		return
	}

	reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_SR)

	for i := range buf {
		var cr uint32

		if i == 0 {
			bits.Set(&cr, TWI_CR_START)
		}

		if i == len(buf)-1 {
			bits.Set(&cr, TWI_CR_STOP)
		}

		if cr != 0 {
			reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, cr)
		}

		if err = hw.wait(TWI_SR_RXRDY); err != nil {
			return
		}

		buf[i] = byte(reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_RHR))
	}

	return hw.wait(TWI_SR_TXCOMP)
}

// Write writes a sequence of bytes to a target device.
//
// The address length (`alen`) parameter should be set greater than 0 for
// ordinary I2C writes (`TARGET W|ADDR|DATA`) and equal to 0 when not sending
// a register address (`TARGET W|DATA`), values less than 0 or greater than 3
// are not valid.
func (hw *TWI) Write(target uint8, addr uint32, alen int, buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if len(buf) == 0 {
		return
	}

	if err = hw.configure(target, addr, alen, false); err != nil {
		return
	}

	reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_SR)

	for i := range buf {
		reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_THR, uint32(buf[i]))

		if i == len(buf)-1 {
			reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_STOP)
		}

		if err = hw.wait(TWI_SR_TXRDY); err != nil {
			return
		}
	}

	return hw.wait(TWI_SR_TXCOMP)
}
