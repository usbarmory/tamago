// DP83825I Ethernet PHY support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dp83825i

import "errors"

// PHY LED control registers
const (
	DP_LEDCR1           = 0x18
	LEDCR1_LINK_LED_DRV = 4
	LEDCR1_LINK_LED_OFF = 1

	DP_LEDCR2           = 0x469
	LEDCR2_LED2_DRV_VAL = 5
	LEDCR2_LED2_DRV_EN  = 4
)

// Table 22–9, MMD access control register bit definitions, 802.3-2008
const (
	MMD_FN_ADDR = 0b00
	MMD_FN_DATA = 0b01
)

// LED controls the PHY connected LED state.
func (hw *PHY) LED(n int, on bool) (err error) {
	switch n {
	case 0, 1:
		val := uint16(1 << LEDCR1_LINK_LED_DRV)

		if !on {
			val |= 1 << LEDCR1_LINK_LED_OFF
		}

		return hw.miim.WritePHYRegister(hw.pa, DP_LEDCR1, val)
	case 2:
		val := uint16(1 << LEDCR2_LED2_DRV_EN)

		if !on {
			val |= 1 << LEDCR2_LED2_DRV_VAL
		}

		// Clause 22 access to Clause 45 MMD registers (802.3-2008)

		// set general MMD registers access
		devad := uint16(0x1f)

		// set address function
		if err = hw.miim.WritePHYRegister(hw.pa, DP_REGCR, uint16(MMD_FN_ADDR<<14)|devad); err != nil {
			return
		}

		// write address value
		if err = hw.miim.WritePHYRegister(hw.pa, DP_ADDAR, DP_LEDCR2); err != nil {
			return
		}

		// set data function
		if err = hw.miim.WritePHYRegister(hw.pa, DP_REGCR, uint16(MMD_FN_DATA<<14)|devad); err != nil {
			return
		}

		// write data value
		return hw.miim.WritePHYRegister(hw.pa, DP_ADDAR, val)
	}

	return errors.New("invalid LED")
}
