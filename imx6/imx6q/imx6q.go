// NXP i.MX6Q initialization
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package imx6q provides hardware initialization, automatically on import, for
// the i.MX6Q family of System-on-Chip components.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package imx6q

import (
	_ "unsafe"
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100000
