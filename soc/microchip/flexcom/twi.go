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
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

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

// TWI_TIMEOUT is the default timeout for TWI operations.
const TWI_TIMEOUT = 100 * time.Millisecond

// InitTWI configures and enables TWI initiator mode. Generic clock generation
// and pin routing are configured separately by the SoC or board package.
func (hw *FLEXCOM) InitTWI() (err error) {
	hw.Lock()
	defer hw.Unlock()

	switch {
	case hw.Base == 0:
		return errors.New("invalid FLEXCOM controller instance")
	case hw.TWIClockDivider > TWI_CWGR_CKDIV_MASK:
		return fmt.Errorf("invalid TWI clock divider %d", hw.TWIClockDivider)
	case hw.TWITimeout < 0:
		return fmt.Errorf("invalid TWI timeout %s", hw.TWITimeout)
	}

	if hw.TWITimeout == 0 {
		hw.TWITimeout = TWI_TIMEOUT
	}

	bits.SetN(&hw.twiClock, TWI_CWGR_CLDIV, TWI_CWGR_CLDIV_MASK, uint32(hw.TWIClockLowDivider))
	bits.SetN(&hw.twiClock, TWI_CWGR_CHDIV, TWI_CWGR_CHDIV_MASK, uint32(hw.TWIClockHighDivider))
	bits.SetN(&hw.twiClock, TWI_CWGR_CKDIV, TWI_CWGR_CKDIV_MASK, uint32(hw.TWIClockDivider))
	bits.SetTo(&hw.twiClock, TWI_CWGR_GCK, hw.TWIGenericClock)

	reg.SetN(hw.Base+FLEX_MR, MR_OPMODE, MR_OPMODE_MASK, MR_OPMODE_TWI)
	hw.resetTWI()

	return
}

func (hw *FLEXCOM) resetTWI() {
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_SWRST)
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CWGR, hw.twiClock)
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_MSEN)
}

func (hw *FLEXCOM) configureTWI(address uint8, internal uint32, internalSize int, read bool) (err error) {
	switch {
	case hw.Base == 0:
		return errors.New("invalid FLEXCOM controller instance")
	case address > TWI_MMR_DADR_MASK:
		return fmt.Errorf("invalid I2C address %#x", address)
	case internalSize < 0 || internalSize > TWI_IADRSZ_MAX:
		return fmt.Errorf("invalid I2C internal address size %d", internalSize)
	case internalSize == 0 && internal != 0:
		return fmt.Errorf("invalid I2C internal address %#x", internal)
	case internalSize > 0 && internal >= 1<<uint(internalSize*8):
		return fmt.Errorf("invalid I2C internal address %#x", internal)
	}

	var mode uint32
	bits.SetN(&mode, TWI_MMR_IADRSZ, TWI_MMR_IADRSZ_MASK, uint32(internalSize))
	bits.SetTo(&mode, TWI_MMR_MREAD, read)
	bits.SetN(&mode, TWI_MMR_DADR, TWI_MMR_DADR_MASK, uint32(address))
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_MMR, mode)
	reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_IADR, internal)

	return
}

func (hw *FLEXCOM) waitTWI(bit int) (err error) {
	start := time.Now()

	for {
		status := reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_SR)

		switch {
		case status&(1<<TWI_SR_NACK) != 0:
			hw.resetTWI()
			return errors.New("i2c target did not acknowledge")
		case status&(1<<uint(bit)) != 0:
			return nil
		case time.Since(start) >= hw.TWITimeout:
			hw.resetTWI()
			return fmt.Errorf("i2c timeout waiting for status bit %d (%#08x)", bit, status)
		}

		runtime.Gosched()
	}
}

// ReadTWI receives bytes from a 7-bit I2C target. internalSize selects zero to
// three address bytes; internal must fit the selected width.
func (hw *FLEXCOM) ReadTWI(address uint8, internal uint32, internalSize int, buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if len(buf) == 0 {
		return
	}

	if err = hw.configureTWI(address, internal, internalSize, true); err != nil {
		return
	}

	reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_SR)

	if len(buf) == 1 {
		reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_START|1<<TWI_CR_STOP)
		if err = hw.waitTWI(TWI_SR_RXRDY); err != nil {
			return
		}
		buf[0] = byte(reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_RHR))
	} else {
		reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_START)
		for i := range buf {
			if i == len(buf)-1 {
				reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_STOP)
			}
			if err = hw.waitTWI(TWI_SR_RXRDY); err != nil {
				return
			}
			buf[i] = byte(reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_RHR))
		}
	}

	err = hw.waitTWI(TWI_SR_TXCOMP)

	return
}

// WriteTWI sends bytes to a 7-bit I2C target. internalSize selects zero to
// three address bytes; internal must fit the selected width.
func (hw *FLEXCOM) WriteTWI(address uint8, internal uint32, internalSize int, buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if len(buf) == 0 {
		return
	}

	if err = hw.configureTWI(address, internal, internalSize, false); err != nil {
		return
	}

	reg.Read(hw.Base + FLEX_TWI_OFFSET + FLEX_TWI_SR)

	for i, value := range buf {
		reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_THR, uint32(value))
		if i == len(buf)-1 {
			reg.Write(hw.Base+FLEX_TWI_OFFSET+FLEX_TWI_CR, 1<<TWI_CR_STOP)
		}
		if err = hw.waitTWI(TWI_SR_TXRDY); err != nil {
			return
		}
	}

	err = hw.waitTWI(TWI_SR_TXCOMP)

	return
}
