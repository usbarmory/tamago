// Raspberry Pi Zero Support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pizero package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pizero

import (
	"github.com/f-secure-foundry/tamago/board/pi-foundation"
)

type board struct{}

// Board provides access to the capabilities of the Pi Zero.
var Board pi.Board = &board{}
