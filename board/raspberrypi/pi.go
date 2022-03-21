// Raspberry Pi support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pi package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pi provides basic abstraction for support of different models of
// Raspberry Pi single board computers.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package pi

// Board provides a basic abstraction over the different models of Pi.
type Board interface {
	LED(name string, on bool) (err error)
}
