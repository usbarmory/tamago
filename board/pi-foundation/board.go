// Raspberry Pi support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pi package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi

// Board provides a basic abstraction over the different models of Pi.
type Board interface {
	LED(name string, on bool) (err error)
}
