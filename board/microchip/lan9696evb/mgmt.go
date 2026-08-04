// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan9696evb

import (
	"fmt"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/microchip/devcpu"
	"github.com/usbarmory/tamago/soc/microchip/lan969x"
	"github.com/usbarmory/tamago/soc/microchip/miim"
)

const (
	MAC_FID             = 1
	ManagementPortIndex = PORT29
)

var (
	// WaitTimeout represents the timeout for PHY register writes
	WaitTimeout = 10 * time.Second
	// PollInterval represents the delay between PHY register write attempts
	PollInterval = 10 * time.Millisecond
)

// On the LAN969x 24-port EVB the management network interface is port D29,
// connected through DEV_RGMII1 with a Microchip LAN8840 PHY.
//
// CPU port 0 (D30) is used for injection and extraction of frames.
var ManagementPort = &devcpu.Port{
	Index:    PORT29,
	IRQ:      lan969x.XTR_READY_IRQ,
	Queue:    lan969x.DEVCPU_QS,
	Analyzer: lan969x.ANA,
	Enable:   enablePort,
	FID:      MAC_FID,
}

func resetInjectionFlowControl(port uint32) {
	reg.Set(DEV_TX_STOP_WM_CFG+port*4, DEV_TX_CNT_CLR)
}

func enablePort() (err error) {
	// init LAN8840 PHY
	speed, err := initPHY(lan969x.MIIM0)

	if err != nil {
		return
	}

	// init MAC controller
	if err = initRGMII(speed); err != nil {
		return
	}

	// init VLAN on physical and CPU port
	initVLAN(PORT29)
	initVLAN(PORT30)

	// init capture on CPU port 0 (D30)
	initCapture(PORT_CFG30)

	// reset injection flow control
	resetInjectionFlowControl(PORT29)

	return nil
}

func initGPIO(num, fn int) {
	pin, err := lan969x.GPIO.Init(num)

	if err != nil {
		return
	}

	pin.Function(fn)
}

func initPHY(miim *miim.MIIM) (speed int, err error) {
	var control uint16

	// Table 2-7: GPIO alternate function assignments
	//
	// GPIO_9:  ALT1 - MIIM0_MDC
	// GPIO_10: ALT1 - MIIM0_MDIO
	initGPIO(9, 1)
	initGPIO(10, 1)

	miim.Init()

	// software reset
	if err = miim.WritePHYRegister(PHY_ADDR, PHY_CTRL, (1 << CTRL_RESET)); err != nil {
		return
	}

	if control, err = waitPHYRegister(miim, PHY_CTRL, 1<<CTRL_RESET, 0); err != nil {
		return 0, fmt.Errorf("could not reset PHY, %v", err)
	}

	// enable and restart auto-negotiation
	control |= (1 << CTRL_ANEG) | (1 << CTRL_ANEG_RESTART)

	if err = miim.WritePHYRegister(PHY_ADDR, PHY_CTRL, control); err != nil {
		return
	}

	statusMask := uint16((1 << STATUS_LINK) | (1 << STATUS_ANEG_COMPLETE))

	if _, err = waitPHYRegister(miim, PHY_STATUS, statusMask, statusMask); err != nil {
		return 0, fmt.Errorf("auto-negotiation status error, %w", err)
	}

	return negotiatedPHYSpeed(miim)
}

func waitPHYRegister(miim *miim.MIIM, address int, mask uint16, value uint16) (data uint16, err error) {
	deadline := time.Now().Add(WaitTimeout)

	for {
		if data, err = miim.ReadPHYRegister(PHY_ADDR, address); err != nil {
			return
		}

		if data&mask == value {
			return
		}

		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timed out waiting for PHY register %#x", address)
		}

		time.Sleep(PollInterval)
	}
}

func negotiatedPHYSpeed(miim *miim.MIIM) (speed int, err error) {
	var local, partner uint16

	if local, err = miim.ReadPHYRegister(PHY_ADDR, PHY_1000_CTRL); err != nil {
		return 0, fmt.Errorf("could not read 1000BASE-T control register, %v", err)
	}

	if partner, err = miim.ReadPHYRegister(PHY_ADDR, PHY_1000_STATUS); err != nil {
		return 0, fmt.Errorf("could not read 1000BASE-T status register, %v", err)
	}

	if local&ANEG_ADV_1000_FULL != 0 && partner&ANEG_LPA_1000_FULL != 0 {
		return 1000, nil
	}

	if local&ANEG_ADV_1000_HALF != 0 && partner&ANEG_LPA_1000_HALF != 0 {
		return 0, fmt.Errorf("unsupported half-duplex management link")
	}

	if local, err = miim.ReadPHYRegister(PHY_ADDR, PHY_ANEG_ADV); err != nil {
		return 0, fmt.Errorf("could not read auto-negotiation advertisement register, %v", err)
	}

	if partner, err = miim.ReadPHYRegister(PHY_ADDR, PHY_ANEG_LPA); err != nil {
		return 0, fmt.Errorf("could not read auto-negotiation link partner ability register, %v", err)
	}

	common := local & partner

	switch {
	case common&ANEG_ADV_100_FULL != 0:
		speed = 100
	case common&ANEG_ADV_100_HALF != 0:
		err = fmt.Errorf("unsupported half-duplex management link")
	case common&(ANEG_ADV_10_FULL|ANEG_ADV_10_HALF) != 0:
		err = fmt.Errorf("unsupported 10 Mbps management link")
	default:
		err = fmt.Errorf("management PHY has no common advertised mode")
	}

	return
}

func initRGMII(speed int) (err error) {
	var val uint32
	var txClock uint32
	var macSpeed uint32

	switch speed {
	case 100:
		txClock = 2
		macSpeed = SPEED_100M
	case 1000:
		txClock = 1
		macSpeed = SPEED_1G
	default:
		return fmt.Errorf("invalid management link speed (%d)", speed)
	}

	// take RGMII out of reset and match the negotiated link speed
	bits.SetN(&val, TX_CLK_CFG, 0b111, txClock)
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

	// set MAC speed
	reg.SetN(DEVRGMII1+DEV_RST_CTRL, SPEED_SEL, 0b111, macSpeed)

	// clear reset from clock domains
	reg.Clear(DEVRGMII1+DEV_RST_CTRL, MAC_TX_RST)
	reg.Clear(DEVRGMII1+DEV_RST_CTRL, MAC_RX_RST)

	return
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
