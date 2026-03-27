// AI Foundry Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package minion

// hart initialization counter
var ncpu int

// Minions returns the number of available ET-Minion cores.
func Minions() (n int) {
	return ncpu
}
