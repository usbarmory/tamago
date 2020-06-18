// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbarmory

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/imx6"
)

// BLE module configuration constants.
//
// On the USB armory Mk II a u-blox ANNA-B112 Bluetooth module is connected as
// illustrated in the following constants.
const (
	// BT_UART_TX
	IOMUXC_SW_MUX_CTL_PAD_UART1_TX_DATA = 0x020e0084
	IOMUXC_SW_PAD_CTL_PAD_UART1_TX_DATA = 0x020e0310

	// BT_UART_RX
	IOMUXC_SW_MUX_CTL_PAD_UART1_RX_DATA = 0x020e0088
	IOMUXC_SW_PAD_CTL_PAD_UART1_RX_DATA = 0x020e0314

	// BT_UART_CTS
	IOMUXC_SW_MUX_CTL_PAD_UART1_CTS_B = 0x020e008c
	IOMUXC_SW_PAD_CTL_PAD_UART1_CTS_B = 0x020e0318

	// BT_UART_RTS (UART1_RTS_B)
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07 = 0x020e0078
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07 = 0x020e0304

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
	IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO05 = 0x020e0070
	IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO05 = 0x020e02fc

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

	DEFAULT_MODE     = 0
	GPIO_MODE        = 5
	UART1_RTS_B_MODE = 8

	RESET_GRACE_TIME = 1 * time.Second
	AT_GRACE_TIME    = 500 * time.Millisecond
)

func configureBLEPad(mux uint32, pad uint32, mode uint32, ctl uint32) {
	p, err := imx6.NewPad(mux, pad, 0)

	if err != nil {
		panic(err)
	}

	p.Mode(mode)
	p.Ctl(ctl)
}

func configureBLEGPIO(num int, instance int, mux uint32, pad uint32, ctl uint32) (gpio *imx6.GPIO) {
	var err error

	gpio, err = imx6.NewGPIO(num, instance, mux, pad)

	if err != nil {
		panic(err)
	}

	gpio.Pad.Mode(GPIO_MODE)
	gpio.Pad.Ctl(ctl)
	gpio.Out()

	return
}

// ANNA implements the interface to the ANNA-B112 module for serial
// communication, reset and mode select.
type ANNA struct {
	sync.Mutex

	uart    *imx6.UART
	reset   *imx6.GPIO
	switch1 *imx6.GPIO
	switch2 *imx6.GPIO
}

// BLE module instance
var BLE = &ANNA{}

func (ble *ANNA) Init() (err error) {
	ble.Lock()
	defer ble.Unlock()

	ctl := uint32(0)
	bits.Set(&ctl, imx6.SW_PAD_CTL_HYS)
	bits.Set(&ctl, imx6.SW_PAD_CTL_PUE)
	bits.Set(&ctl, imx6.SW_PAD_CTL_PKE)

	bits.SetN(&ctl, imx6.SW_PAD_CTL_PUS, 0b11, imx6.SW_PAD_CTL_PUS_PULL_UP_100K)
	bits.SetN(&ctl, imx6.SW_PAD_CTL_SPEED, 0b11, imx6.SW_PAD_CTL_SPEED_100MHZ)
	bits.SetN(&ctl, imx6.SW_PAD_CTL_DSE, 0b111, imx6.SW_PAD_CTL_DSE_2_R0_6)

	// BT_UART_TX
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_UART1_TX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART1_TX_DATA,
		DEFAULT_MODE, ctl)

	// BT_UART_RX
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_UART1_RX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART1_RX_DATA,
		DEFAULT_MODE, ctl)

	// BT_UART_CTS
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_UART1_CTS_B,
		IOMUXC_SW_PAD_CTL_PAD_UART1_CTS_B,
		DEFAULT_MODE, ctl)

	bits.SetN(&ctl, imx6.SW_PAD_CTL_PUS, 0b11, imx6.SW_PAD_CTL_PUS_PULL_DOWN_100K)

	// BT_UART_RTS
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO07,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO07,
		UART1_RTS_B_MODE, ctl)

	bits.SetN(&ctl, imx6.SW_PAD_CTL_PUS, 0b11, imx6.SW_PAD_CTL_PUS_PULL_UP_22K)
	bits.SetN(&ctl, imx6.SW_PAD_CTL_SPEED, 0b11, imx6.SW_PAD_CTL_SPEED_50MHZ)
	bits.SetN(&ctl, imx6.SW_PAD_CTL_DSE, 0b111, imx6.SW_PAD_CTL_DSE_2_R0_4)

	// BT_UART_DSR
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_UART3_TX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART3_TX_DATA,
		ctl, GPIO_MODE)

	// BT_SWDCLK
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO04,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO04,
		GPIO_MODE, ctl)

	// BT_SWDIO
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO05,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO05,
		GPIO_MODE, ctl)

	// BT_RESET
	BLE.reset = configureBLEGPIO(BT_RESET, 1,
		IOMUXC_SW_MUX_CTL_PAD_GPIO1_IO09,
		IOMUXC_SW_PAD_CTL_PAD_GPIO1_IO09,
		ctl)

	// BT_SWITCH_1
	BLE.switch1 = configureBLEGPIO(BT_SWITCH_1, 1,
		IOMUXC_SW_MUX_CTL_PAD_UART3_RTS_B,
		IOMUXC_SW_PAD_CTL_PAD_UART3_RTS_B,
		ctl)

	// BT_SWITCH_2
	BLE.switch2 = configureBLEGPIO(BT_SWITCH_2, 1,
		IOMUXC_SW_MUX_CTL_PAD_UART3_CTS_B,
		IOMUXC_SW_PAD_CTL_PAD_UART3_CTS_B,
		ctl)

	ctl = 0
	bits.SetN(&ctl, imx6.SW_PAD_CTL_DSE, 0b111, imx6.SW_PAD_CTL_DSE_2_R0_4)
	bits.Set(&ctl, imx6.SW_PAD_CTL_HYS)

	// BT_UART_DTR
	configureBLEPad(IOMUXC_SW_MUX_CTL_PAD_UART3_RX_DATA,
		IOMUXC_SW_PAD_CTL_PAD_UART3_RX_DATA,
		GPIO_MODE, ctl)

	BLE.uart = imx6.UART1
	BLE.uart.Init(imx6.UART_DEFAULT_BAUDRATE)

	return
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

// Enter bootloader mode by driving low SWITCH_1 and SWITCH_2 during a module
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

// Enter normal mode by driving high SWITCH_1 and SWITCH_2 during a module
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

// AT transmits an AT command to the BLE module serial interface.
func (ble *ANNA) AT(cmd string) (res string, err error) {
	var r []byte

	if ble.uart == nil {
		err = errors.New("module is not initialized")
		return
	}

	cmd = "AT" + cmd + "\r"

	ble.Lock()
	defer ble.Unlock()

	for i := 0; i < len(cmd); i++ {
		ble.uart.Write(cmd[i])
	}

	time.Sleep(AT_GRACE_TIME)

	for {
		c, ok := ble.uart.Read()

		if !ok {
			break
		}

		r = append(r, c)

		if strings.Contains(string(r), "OK") {
			break
		}

		if strings.Contains(string(r), "ERROR") {
			err = errors.New("response error")
			return
		}
	}

	return string(r), nil
}
