// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan9696evb

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/devcpu"
	"github.com/usbarmory/tamago/soc/microchip/lan969x"
	"github.com/usbarmory/tamago/soc/microchip/miim"
)

const MAC_FID = 1

// On the LAN969x 24-port EVB the management network interface is port D29,
// connected through DEV_RGMII1 with a Microchip LAN8840 PHY.
//
// CPU port 0 (D30) is used for injection and extraction of frames.
var ManagementPort = &devcpu.Port{
	Index:    29,
	IRQ:      lan969x.XTR_READY_IRQ,
	Queue:    lan969x.DEVCPU_QS,
	Analyzer: lan969x.ANA,
	Enable:   enablePort,
	FID:      MAC_FID,
}

func enablePort() (err error) {
	// init LAN8840 PHY
	initPHY(lan969x.MIIM0)

	// init MAC controller
	initRGMII()

	// init VLAN on physical and CPU port
	initVLAN(PORT29)
	initVLAN(PORT30)

	// init capture on CPU port 0 (D30)
	initCapture(PORT_CFG30)

	return nil
}

func initPHY(miim *miim.MIIM) {
	// Table 2-7: GPIO alternate function assignments
	//
	// GPIO_9:  ALT1 - MIIM0_MDC
	// GPIO_10: ALT1 - MIIM0_MDIO
	lan969x.GPIO.Function(9, 1)
	lan969x.GPIO.Function(10, 1)

	// software reset
	miim.WritePHYRegister(PHY_ADDR, PHY_CTRL, (1 << CTRL_RESET))
	// 1000 Mbps, Auto-Negotiation, Full-duplex
	miim.WritePHYRegister(PHY_ADDR, PHY_CTRL, (0b1<<CTRL_SPEED1)|(1<<CTRL_ANEG)|(1<<CTRL_DUPLEX))
}

func initRGMII() {
	var val uint32

	// take RGMII out of reset and set speed to 1G
	bits.SetN(&val, TX_CLK_CFG, 0b111, 1)
	bits.Clear(&val, RGMII_TX_RST)
	bits.Clear(&val, RGMII_RX_RST)
	reg.Write(XMIICFG1+RGMII_CFG, val)

	// enable RGMII0 on the GPIOs
	reg.SetN(XMIICFG0+XMII_CFG, GPIO_XMII_CFG, 0b11, CFG_RGMII)

	// rx delay lines
	val = reg.Read(XMIICFG1 + DLL_CFG0)
	bits.Set(&val, DLL_ENA)                // start delay tuning state machine
	bits.Clear(&val, DLL_CLK_ENA)          // bypass DLL
	bits.SetN(&val, DLL_CLK_SEL, 0b111, 4) // DLL phase shift 90º or 2ns at 125MHz
	bits.Clear(&val, DLL_RST)              // bring DLL out of reset
	reg.Write(XMIICFG1+DLL_CFG0, val)

	// tx delay lines
	val = reg.Read(XMIICFG1 + DLL_CFG1)
	bits.Set(&val, DLL_ENA)                // start delay tuning state machine
	bits.Set(&val, DLL_CLK_ENA)            // use DLL
	bits.SetN(&val, DLL_CLK_SEL, 0b111, 4) // DLL phase shift 90º or 2ns at 125MHz
	bits.Clear(&val, DLL_RST)              // bring DLL out of reset
	reg.Write(XMIICFG1+DLL_CFG1, val)

	// enable MAC rx/tx, Full-Duplex
	reg.Set(DEVRGMII1+MAC_ENA_CFG, RX_ENA)
	reg.Set(DEVRGMII1+MAC_ENA_CFG, TX_ENA)
	reg.Set(DEVRGMII1+MAC_MODE_CFG, FDX_ENA)

	// set inter frame gaps
	reg.SetN(DEVRGMII1+MAC_IFG_CFG, TX_IFG, 0x1f, 4)  // tx inter frame gap
	reg.SetN(DEVRGMII1+MAC_IFG_CFG, RX_IFG2, 0x0f, 1) // rx inter frame gap (second part)
	reg.SetN(DEVRGMII1+MAC_IFG_CFG, RX_IFG1, 0x0f, 5) // rx inter frame gap (first part)

	// set 1000Mbps speed
	reg.SetN(DEVRGMII1+DEV_RST_CTRL, SPEED_SEL, 0b111, SPEED_1G)

	// clear reset from clock domains
	reg.Clear(DEVRGMII1+DEV_RST_CTRL, MAC_TX_RST)
	reg.Clear(DEVRGMII1+DEV_RST_CTRL, MAC_RX_RST)
}

func initVLAN(port uint32) {
	reg.Set(port+VLAN_CTRL, VLAN_AWARE_ENA)         // enable VLAN awareness
	reg.SetN(port+VLAN_CTRL, VLAN_POP_CNT, 0b11, 1) // number of VLAN tags to remove
	reg.SetN(port+VLAN_CTRL, PORT_VID, 0xfff, 1)    // set VLAN ID
}

func initCapture(port uint32) {
	// enable ports for any frame transfer
	reg.Set(SWITCH_PORT_MODE29, PORT_ENA) // mgmt port
	reg.Set(SWITCH_PORT_MODE30, PORT_ENA) // cpu port

	// configure CPU port
	reg.Set(port, NO_PREAMBLE_ENA)          // no preamble
	reg.Set(port, PAD_ENA)                  // enable padding
	reg.SetN(port, INJ_FORMAT_CFG, 0b11, 0) // no internal frame header

	// recalc injected frame FCS
	reg.Set(PORT30+FILTER_CTRL, FORCE_FCS_UPDATE_ENA)

	// CPU copy of frames found in MAC table
	reg.Set(FWD_CFG, CPU_DMAC_COPY_ENA)
}
