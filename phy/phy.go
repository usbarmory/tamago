// Ethernet PHY support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package phy provides abstraction for Ethernet PHY driver support.
package phy

// MIIM is the common interface used by sub-packages to access Media
// Independent Interface Management (MIIM) buses.
type MIIM interface {
	// Init represents the MIIM bus initializattion function.
	Init() error

	// ReadPHYRegister reads a standard management register of a connected Ethernet
	// PHY (IEE 802.3-2008 Clause 22).
	ReadPHYRegister(pa int, ra int) (data uint16, err error)

	// WritePHYRegister writes a standard management register of a connected
	// Ethernet PHY (IEE 802.3-2008 Clause 22).
	WritePHYRegister(pa int, ra int, data uint16) (err error)
}
