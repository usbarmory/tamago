// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// The usbarmory package provides hardware initialization, automatically on
// import, for the USB armory Mk II single board computer. It is meant to be
// used on bare metal with tamago/arm.
package usbarmory

import (
	_ "github.com/f-secure-foundry/tamago/imx6/imx6ul"
)
