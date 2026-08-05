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
//
// The bus is expected to be already initialized when passed to a PHY driver.
type MIIM interface {
	// ReadPHYRegister reads a standard management register of a connected Ethernet
	// PHY (IEEE 802.3-2008 Clause 22).
	ReadPHYRegister(pa int, ra int) (data uint16, err error)

	// WritePHYRegister writes a standard management register of a connected
	// Ethernet PHY (IEEE 802.3-2008 Clause 22).
	WritePHYRegister(pa int, ra int, data uint16) (err error)
}

// PHY is the common interface used by sub-packages to implement Ethernet PHY
// drivers.
type PHY interface {
	// Init represents the PHY initialization function.
	Init(addr int, miim MIIM) error

	// Negotiate represents the PHY auto-negotiation function.
	Negotiate() error

	// Address reports the PHY address.
	Address() int

	// Speed reports the PHY auto-negotiated speed.
	Speed() int
}
