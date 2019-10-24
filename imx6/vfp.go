// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	_ "unsafe"
)

// defined in vfp.s
func vfp_enable()

//go:linkname enableVFP runtime.vfp_enable
func enableVFP() {
	vfp_enable()
}
