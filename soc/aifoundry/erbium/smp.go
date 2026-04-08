// AI Foundry Erbium initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package erbium

// hart initialization counter
var ncpu int

// NumCPU returns the number of logical CPUs initialized on the platform.
func NumCPU() (n int) {
	return ncpu
}
