// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mk2

// The USB armory Mk II has the following components accessible as I²C targets.
const (
	// Power management controller
	PF1510_ADDR = 0x08
	// Cryptographic co-processor (rev. γ only)
	SE050_ADDR = 0x48
	// Cryptographic co-processor (rev. β only)
	A71CH_ADDR = 0x48
	// Cryptographic co-processor (rev. β only)
	ATECC_ADDR = 0x60
	// Type-C plug port controller
	TUSB320_ADDR = 0x61
	// Type-C receptacle port controller
	FUSB303_ADDR = 0x31
)

func init() {
	// I2C1 is used to enable the USB receptacle controller as well as low
	// switch SD card signaling to low voltage.
	I2C1.Init()
}
