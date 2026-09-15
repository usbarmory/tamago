// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

import (
	"errors"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

// Interrupt defines the interrupt types for [GVE.EnableInterrupt].
type Interrupt int

const (
	TX         Interrupt = iota
	RX                   = 1
	Management           = 2
)

func (hw *GVE) addCapability(off uint32, hdr *pci.CapabilityHeader) error {
	switch hdr.Vendor {
	case pci.MSIX:
		c := &pci.CapabilityMSIX{}

		if err := c.Unmarshal(hw.Device, off); err != nil {
			return err
		}

		hw.msix = c
	}

	return nil
}

// EnableInterrupt enables MSI-X interrupt vector routing to a LAPIC instance
// for the desired [Interrupt].
func (hw *GVE) EnableInterrupt(id int, kind Interrupt) (err error) {
	if kind < 0 || kind > Management {
		return errors.New("invalid interrypt type")
	}

	if hw.msix == nil {
		return errors.New("missing required capabilities")
	}

	if hw.msix.TableSize() < 3 {
		return errors.New("invalid MSI-X table size")
	}

	addr := uint64(amd64.LAPIC_BASE)
	data := uint32(id)

	return hw.msix.EnableInterrupt(int(kind), addr, data)
}
