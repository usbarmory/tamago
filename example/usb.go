// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package main

import (
	"github.com/inversepath/tamago/imx6"
)

func TestUSB() {
	// TODO: work in progress
	imx6.USB1.Init()
	imx6.USB1.DeviceMode()
}
