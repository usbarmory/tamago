// MCIMX6ULL-EVK support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mx6ullevk

import (
	"errors"

	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// KSZ8081RNB configuration constants.
//
// On the MCIMX6ULL-EVK the ENET MACs are each connected to an KSZ8081RNB PHY,
// this board package configures them at 100 Mbps / Full-duplex mode.
const (
	KSZ_CTRL    = 0x00
	CTRL_RESET  = 15
	CTRL_SPEED  = 13
	CTRL_DUPLEX = 8

	KSZ_INT = 0x1b

	KSZ_PHYCTRL2  = 0x1f
	CTRL2_HP_MDIX = 15
	CTRL2_RMII    = 7
	CTRL2_LED     = 4
)

func init() {
	imx6ul.ENET1.EnablePHY = EnablePHY
	imx6ul.ENET2.EnablePHY = EnablePHY

	imx6ul.ENET1.RMII = true
	imx6ul.ENET2.RMII = true
}

func EnablePHY(eth *enet.ENET) error {
	var pa int

	switch eth.Index {
	case 1:
		pa = 2
	case 2:
		pa = 1
	default:
		return errors.New("invalid index")
	}

	// Software reset
	eth.WriteMII(pa, KSZ_CTRL, (1 << CTRL_RESET))
	// HP Auto MDI/MDI-X mode, RMII 50MHz, LEDs: Activity/Link
	eth.WriteMII(pa, KSZ_PHYCTRL2, (1 << CTRL2_HP_MDIX) | (1 << CTRL2_RMII) | (1 << CTRL2_LED))
	// 100 Mbps, Full-duplex
	eth.WriteMII(pa, KSZ_CTRL, (1 << CTRL_SPEED) | (1 << CTRL_DUPLEX))
	// enable interrupts
	eth.WriteMII(pa, KSZ_INT, 0xff00)

	return nil
}
