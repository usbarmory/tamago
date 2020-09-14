// Raspberry Pi 2 support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi2

type board struct{}

// Board provides access to the capabilities of the Pi2.
var Board pi.Board = &board{}
