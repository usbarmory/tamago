// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mk2

import (
	"errors"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/soc/nxp/gpio"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/iomuxc"
	"github.com/usbarmory/tamago/soc/nxp/uart"
)

// BLE module configuration constants.
//
// On models UA-MKII-β and UA-MKII-γ a u-blox ANNA-B112 Bluetooth module is
// connected to UART1.
//
// On the USB armory Mk II β revision, due to an errata, the RTS/CTS signals
// are connected inverted on the Bluetooth module side. This is automatically
// handled with a workaround by the RTS() and CTS() functions, which use the
// lines as GPIOs to invert their direction.
const (
	// BT_UART_TX (UART1_TX_DATA)
	IOMUXC_SW_MUX_CTL_PAD_UART1_TX_DATA = 0x020e0084
	IOMUXC_SW_PAD_CTL_PAD_UART1_TX_DATA = 0x020e0310

	// BT_UART_RX (UART1_RX_DATA)
	IOMUXC_SW_MUX_CTL_PAD_UART1_RX_DATA = 0x020e0088
	IOMUXC_SW_PAD_CTL_PAD_UART1_RX_DATA = 0x020e0314
	IOMUXC_UART1_RX_DATA_SELECT_INPUT   = 0x020e0624
	DAISY_UART1_RX_DATA                 = 0b11

	// BT_UART_CTS (UART1_CTS_B)
	IOMUXC_SW_MUX_CTL_PAD_UART1_CTS_B = 0x020e008c
	IOMUXC_SW_PAD_CTL_PAD_UART1_CTS_B = 0x020e0318

	// BT_UART_RTS (UART1_RTS_B)
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07 = 0x020e0078
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07 = 0x020e0304
	IOMUXC_UART1_RTS_B_SELECT_INPUT  = 0x020e0620
	UART1_RTS_B_MODE                 = 8
	DAISY_GPIO1_IO07                 = 0b01

	// BT_UART_DSR (GPIO1_IO24)
	IOMUXC_SW_MUX_CTL_PAD_UART3_TX_DATA = 0x020e00a4
	IOMUXC_SW_PAD_CTL_PAD_UART3_TX_DATA = 0x020e0330

	// BT_UART_DTR (GPIO1_IO25)
	IOMUXC_SW_MUX_CTL_PAD_UART3_RX_DATA = 0x020e00a8
	IOMUXC_SW_PAD_CTL_PAD_UART3_RX_DATA = 0x020e0334

	// BT_SWDCLK
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO04 = 0x020e006c
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO04 = 0x020e02f8

	// BT_SWDIO
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO06 = 0x020e0074
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO06 = 0x020e0300

	// BT_RESET
	BT_RESET                         = 9
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO09 = 0x020e0080
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO09 = 0x020e030c

	// BT_SWITCH_1 (GPIO1_IO27)
	BT_SWITCH_1                       = 27
	IOMUXC_SW_MUX_CTL_PAD_UART3_RTS_B = 0x020e00b0
	IOMUXC_SW_PAD_CTL_PAD_UART3_RTS_B = 0x020e033c

	// BT_SWITCH_2 (GPIO1_IO26)
	BT_SWITCH_2                       = 26
	IOMUXC_SW_MUX_CTL_PAD_UART3_CTS_B = 0x020e00ac
	IOMUXC_SW_PAD_CTL_PAD_UART3_CTS_B = 0x020e0338

	DEFAULT_MODE = 0
	GPIO_MODE    = 5

	RESET_GRACE_TIME = 1 * time.Second
)

func configureBLEPad(mux uint32, pad uint32, daisy uint32, mode uint32, ctl uint32) (p *iomuxc.Pad) {
	p = &iomuxc.Pad{
		Mux:   mux,
		Pad:   pad,
		Daisy: daisy,
	}

	p.Mode(mode)
	p.Ctl(ctl)

	return
}

func configureBLEGPIO(num int, gpio *gpio.GPIO, mux uint32, pad uint32, ctl uint32) (pin *gpio.Pin) {
	var err error

	if pin, err = gpio.Init(num); err != nil {
		panic(err)
	}

	pin.Out()

	p := iomuxc.Init(mux, pad, GPIO_MODE)
	p.Ctl(ctl)

	return
}

// ANNA implements the interface to the ANNA-B112 module for serial
// communication, reset and mode select.
type ANNA struct {
	sync.Mutex

	UART *uart.UART

	reset   *gpio.Pin
	switch1 *gpio.Pin
	switch2 *gpio.Pin

	// On β revisions RTS/CTS are implemented as GPIO due to errata.
	errata bool
	rts    *gpio.Pin
	cts    *gpio.Pin
}

// BLE module instance
var BLE = &ANNA{}

// Init initializes, in normal mode, a BLE module instance.
func (ble *ANNA) Init() (err error) {
	ble.Lock()
	defer ble.Unlock()

	ble.UART = UART1

	ctl := uint32(0)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_HYS)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_PUE)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_PKE)

	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_PUS, 0b11, iomuxc.SW_PAD_CTL_PUS_PULL_UP_100K)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_SPEED, 0b11, iomuxc.SW_PAD_CTL_SPEED_100MHZ)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_DSE, 0b111, iomuxc.SW_PAD_CTL_DSE_2_R0_6)

	// BT_UART_TX
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_UART1_TX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART1_TX_DATA,
		0, DEFAULT_MODE, ctl)

	// BT_UART_RX
	pad := configureBLEPad(
		IOMUXC_SW_MUX_CTL_PAD_UART1_RX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART1_RX_DATA,
		IOMUXC_UART1_RX_DATA_SELECT_INPUT,
		DEFAULT_MODE, ctl)
	pad.Select(DAISY_UART1_RX_DATA)

	switch model() {
	case LAN:
		return errors.New("unavailable on this model")
	case BETA:
		ble.errata = true

		// On β BT_UART_RTS is set to GPIO for CTS due to errata.
		ble.cts = configureBLEGPIO(7, imx6ul.GPIO1,
			IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07,
			IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07,
			ctl)
		ble.cts.Out()

		bits.SetN(&ctl, iomuxc.SW_PAD_CTL_PUS, 0b11, iomuxc.SW_PAD_CTL_PUS_PULL_DOWN_100K)

		// On β BT_UART_CTS is set to GPIO for RTS due to errata.
		ble.rts = configureBLEGPIO(18, imx6ul.GPIO1,
			IOMUXC_SW_MUX_CTL_PAD_UART1_CTS_B,
			IOMUXC_SW_PAD_CTL_PAD_UART1_CTS_B,
			ctl)
		ble.rts.In()

		ble.UART.Flow = false
	default:
		// BT_UART_CTS
		pad = configureBLEPad(
			IOMUXC_SW_MUX_CTL_PAD_UART1_CTS_B,
			IOMUXC_SW_PAD_CTL_PAD_UART1_CTS_B,
			0, DEFAULT_MODE, ctl)

		bits.SetN(&ctl, iomuxc.SW_PAD_CTL_PUS, 0b11, iomuxc.SW_PAD_CTL_PUS_PULL_DOWN_100K)

		// BT_UART_RTS
		pad = configureBLEPad(
			IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07,
			IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07,
			IOMUXC_UART1_RTS_B_SELECT_INPUT,
			UART1_RTS_B_MODE, ctl)
		pad.Select(DAISY_GPIO1_IO07)

		ble.UART.Flow = true
	}

	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_PUS, 0b11, iomuxc.SW_PAD_CTL_PUS_PULL_UP_22K)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_SPEED, 0b11, iomuxc.SW_PAD_CTL_SPEED_50MHZ)
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_DSE, 0b111, iomuxc.SW_PAD_CTL_DSE_2_R0_4)

	// BT_UART_DSR
	configureBLEPad(
		IOMUXC_SW_MUX_CTL_PAD_UART3_TX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART3_TX_DATA,
		0, ctl, GPIO_MODE)

	// BT_SWDCLK
	configureBLEPad(
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO04,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO04,
		0, GPIO_MODE, ctl)

	// BT_SWDIO
	configureBLEPad(
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO06,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO06,
		0, GPIO_MODE, ctl)

	// BT_RESET
	ble.reset = configureBLEGPIO(BT_RESET, imx6ul.GPIO1,
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO09,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO09,
		ctl)

	// BT_SWITCH_1
	ble.switch1 = configureBLEGPIO(BT_SWITCH_1, imx6ul.GPIO1,
		IOMUXC_SW_MUX_CTL_PAD_UART3_RTS_B,
		IOMUXC_SW_PAD_CTL_PAD_UART3_RTS_B,
		ctl)

	// BT_SWITCH_2
	ble.switch2 = configureBLEGPIO(BT_SWITCH_2, imx6ul.GPIO1,
		IOMUXC_SW_MUX_CTL_PAD_UART3_CTS_B,
		IOMUXC_SW_PAD_CTL_PAD_UART3_CTS_B,
		ctl)

	ctl = 0
	bits.SetN(&ctl, iomuxc.SW_PAD_CTL_DSE, 0b111, iomuxc.SW_PAD_CTL_DSE_2_R0_4)
	bits.Set(&ctl, iomuxc.SW_PAD_CTL_HYS)

	// BT_UART_DTR
	configureBLEPad(
		IOMUXC_SW_MUX_CTL_PAD_UART3_RX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART3_RX_DATA,
		0, GPIO_MODE, ctl)

	ble.UART.Init()

	// reset in normal mode
	ble.switch1.High()
	ble.switch2.High()
	ble.reset.Low()
	ble.reset.High()

	return
}

// RTS returns whether the BLE module allows to send or not data, only useful
// on β boards when a workaround to the RTS/CTS errata is required.
func (ble *ANNA) RTS() (ready bool) {
	if !ble.errata {
		return
	}

	return !ble.rts.Value()
}

// CTS signals the BLE module whether it is allowed to send or not data, only
// useful on β boards when a workaround to the RTS/CTS errata is required.
func (ble *ANNA) CTS(clear bool) {
	if !ble.errata {
		return
	}

	if clear {
		ble.cts.Low()
	} else {
		ble.cts.High()
	}
}

// Reset the BLE module by toggling the RESET_N pin.
func (ble *ANNA) Reset() (err error) {
	ble.Lock()
	defer ble.Unlock()

	if ble.reset == nil {
		return errors.New("module is not initialized")
	}

	ble.reset.Low()
	defer ble.reset.High()

	time.Sleep(RESET_GRACE_TIME)

	return
}

// Enter normal mode by driving high SWITCH_1 and SWITCH_2 during a module
// reset cycle.
func (ble *ANNA) NormalMode() (err error) {
	ble.Lock()
	defer ble.Unlock()

	if ble.switch1 == nil || ble.switch2 == nil {
		return errors.New("module is not initialized")
	}

	ble.switch1.High()
	ble.switch2.High()

	return ble.Reset()
}

// Enter bootloader mode by driving low SWITCH_1 and SWITCH_2 during a module
// reset cycle.
func (ble *ANNA) BootloaderMode() (err error) {
	ble.Lock()
	defer ble.Unlock()

	if ble.switch1 == nil || ble.switch2 == nil {
		return errors.New("module is not initialized")
	}

	ble.switch1.Low()
	ble.switch2.Low()

	return ble.Reset()
}
