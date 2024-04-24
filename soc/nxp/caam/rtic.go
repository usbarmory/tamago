// NXP Cryptographic Acceleration and Assurance Module (CAAM) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package caam

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// CAAM RTIC tuning
var (
	// RTICThrottle defines the clock cycles between RTIC hash operations
	RTICThrottle uint32 = 0x000000ff
	// RTICWatchdog defines the starting value for the RTIC Watchdog Timer
	RTICWatchdog uint64 = 0x0000ffffffffffff
)

// CAAM RTIC registers
const (
	CAAM_RSTA = 0x6004
	RSTA_CS   = 25
	RSTA_AE   = 8
	RSTA_MIS  = 4
	RSTA_HE   = 3
	RSTA_SV   = 2
	RSTA_HD   = 1
	RSTA_BSY  = 0

	CAAM_RCMD = 0x600c
	RCMD_RTC  = 2
	RCMD_HO   = 1

	CAAM_RCTL  = 0x6014
	RCTL_RTME  = 8
	RCTL_HOME  = 4
	RCTL_RREQS = 1

	CAAM_RTHR  = 0x601c
	CAAM_RWDOG = 0x6028

	// RTIC Memory Block Address
	CAAM_RMaAb = 0x6100
	// RTIC Memory Block Length
	CAAM_RMaLb = 0x610c
)

// CAAM RTIC states
const (
	CS_IDLE = iota
	CS_SINGLE
	CS_RUN
	CS_ERR
)

// MemoryBlock represents a memory region for RTIC monitoring.
type MemoryBlock struct {
	Address uint32
	Length  uint32
}

// RSTA returns the Run Time Integrity Checker (RTIC) Status.
func (hw *CAAM) RSTA() (cs uint32, err error) {
	rsta := reg.Read(hw.Base + CAAM_RSTA)

	switch {
	case bits.Get(&rsta, RSTA_AE, 0xf) != 0:
		err = errors.New("illegal address")
	case bits.Get(&rsta, RSTA_MIS, 0xf) != 0:
		err = errors.New("memory block corruption")
	case bits.Get(&rsta, RSTA_HE, 1) != 0:
		err = errors.New("hashing error")
	case bits.Get(&rsta, RSTA_SV, 1) != 0:
		err = errors.New("security violation")
	case bits.Get(&rsta, RSTA_BSY, 1) != 0:
		err = errors.New("RTIC busy")
	}

	cs = bits.Get(&rsta, RSTA_CS, 0b11)

	return
}

func (hw *CAAM) initRTIC(blocks []MemoryBlock) error {
	if len(blocks) == 0 || len(blocks) > 4 {
		return errors.New("invalid memory block count")
	}

	for i, b := range blocks {
		reg.Write(hw.Base+CAAM_RMaAb+uint32(i)*0x20+0x4, b.Address)
		reg.Write(hw.Base+CAAM_RMaLb+uint32(i)*0x20, b.Length)
	}

	return nil
}

// EnableRTIC enables the CAAM Run Time Integrity Checker (RTIC) for up to 4
// memory blocks.
//
// Once enabled, the RTIC performs periodic (see RTICThrottle) hardware backed
// SHA256 hashing and raising a security violation (see RSTA()) in case of
// mismatch with the first computed hash.
//
// Any security violation (which also affects the SNVS SSM) or memory block
// reconfiguration require a hardware reset.
func (hw *CAAM) EnableRTIC(blocks []MemoryBlock) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 {
		return errors.New("invalid CAAM instance")
	}

	if err = hw.initRTIC(blocks); err != nil {
		return
	}

	if cs, err := hw.RSTA(); err != nil || cs != CS_IDLE {
		return fmt.Errorf("invalid state (cs:%d, err:%v)", cs, err)
	}

	// hash maximum number of blocks at each iteration
	reg.SetN(hw.Base+CAAM_RCTL, RCTL_RREQS, 0b111, 0b111)

	// apply run-time tuning settings
	reg.Write(hw.Base+CAAM_RTHR, RTICThrottle)
	reg.Write(hw.Base+CAAM_RWDOG, uint32(RTICWatchdog>>32))
	reg.Write(hw.Base+CAAM_RWDOG+0x4, uint32(RTICWatchdog))

	// set memory blocks for hash once and run-time check operations
	for i := 0; i < len(blocks); i++ {
		reg.Set(hw.Base+CAAM_RCTL, RCTL_HOME+i)
		reg.Set(hw.Base+CAAM_RCTL, RCTL_RTME+i)
	}

	// generate reference hash value and start run-time verification
	reg.SetN(hw.Base+CAAM_RCMD, RCMD_HO, 0b11, 0b11)

	return
}
