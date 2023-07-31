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
	"time"
)

// Front USB port controller constants
// (plug: UA-MKII-β, UA-MKII-γ, receptacle: UA-MKII-LAN).
const (
	TUSB320_CSR_2      = 0x09
	CSR_ATTACHED_STATE = 6
)

// Front USB port controller modes
// (plug: UA-MKII-β, UA-MKII-γ, receptacle: UA-MKII-LAN).
const (
	STATE_NOT_ATTACHED = iota
	STATE_ATTACHED_SRC
	STATE_ATTACHED_SNK
	STATE_ATTACHED_ACC
)

// Side receptacle USB port controller constants (UA-MKII-β, UA-MKII-γ)
const (
	FUSB303_CONTROL1 = 0x05
	CONTROL1_ENABLE  = 3

	FUSB303_TYPE = 0x13
)

// Side receptacle USB port controller modes (UA-MKII-β, UA-MKII-γ)
const (
	TYPE_DEBUGSRC    = 1 << 6
	TYPE_DEBUGSNK    = 1 << 5
	TYPE_SINK        = 1 << 4
	TYPE_SOURCE      = 1 << 3
	TYPE_ACTIVECABLE = 1 << 2
	TYPE_AUDIOVBUS   = 1
	TYPE_AUDIO       = 0
)

// FrontPortMode returns the type of attached device detected by the front USB
// port controller.
func FrontPortMode() (mode int, err error) {
	t, err := I2C1.Read(TUSB320_ADDR, TUSB320_CSR_2, 1, 1)

	if err != nil {
		return
	}

	return int(t[0] >> CSR_ATTACHED_STATE), err
}

// EnableReceptacleController activates the side receptacle USB port
// controller.
func EnableReceptacleController() (err error) {
	a, err := I2C1.Read(FUSB303_ADDR, FUSB303_CONTROL1, 1, 1)

	if err != nil {
		return
	}

	a[0] |= 1 << CONTROL1_ENABLE
	return I2C1.Write(a, FUSB303_ADDR, FUSB303_CONTROL1, 1)
}

// ReceptacleMode returns the type of device or accessory detected by the side
// receptacle USB port controller.
func ReceptacleMode() (mode int, err error) {
	t, err := I2C1.Read(FUSB303_ADDR, FUSB303_TYPE, 1, 1)

	if err != nil {
		return
	}

	return int(t[0]), err
}

// EnableDebugAccessory enables debug accessory detection on the side
// receptacle USB port controller.
//
// A debug accessory allows access, among all other debug signals, to the UART2
// serial console.
//
// Note that there is a delay (typically up to 200ms) between the return of
// this call and the actual enabling of the debug accessory, for this reason
// the serial console is not immediately available.
//
// To wait detection of a debug accessory use DetectDebugAccessory() instead.
func EnableDebugAccessory() (err error) {
	return EnableReceptacleController()
}

func waitDebugAccessory(timeout time.Duration, ch chan<- bool) {
	start := time.Now()

	for time.Since(start) < timeout {
		mode, err := ReceptacleMode()

		if err != nil {
			break
		}

		if mode == TYPE_DEBUGSRC || mode == TYPE_DEBUGSNK {
			ch <- true
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	ch <- false
}

// DetectDebugAccessory enables debug accessory detection on the side
// receptacle USB port controller and polls successful detection (typically
// done in up to 200ms).
//
// An error is returned if no debug accessory is detected within the timeout.
//
// On the returned boolean channel are sent successful detection (true) or
// timeout (false).
func DetectDebugAccessory(timeout time.Duration) (<-chan bool, error) {
	ch := make(chan bool)

	if err := EnableReceptacleController(); err != nil {
		return nil, err
	}

	go waitDebugAccessory(timeout, ch)

	return ch, nil
}
