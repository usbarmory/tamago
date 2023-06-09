// NXP Cryptographic Acceleration and Assurance Module (CAAM) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package caam implements a driver for the NXP Cryptographic Acceleration and
// Assurance Module (CAAM) adopting the following reference specifications:
//   - IMX6ULSRM - i.MX6UL Security Reference Manual - Rev 0 04/2016
//   - IMX7DSSRM - i.MX7DS Security Reference Manual - Rev 0 03/2017
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package caam

import (
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// CAAM registers
const (
	CAAM_SCFGR     = 0xc
	SCFGR_RNGSH0   = 9
	SCFGR_RANDDPAR = 8

	CAAM_RTMCTL     = 0x600
	RTMCTL_PRGM     = 16
	RTMCTL_ENT_VAL  = 10
	RTMCTL_RST_DEF  = 6
	RTMCTL_TRNG_ACC = 5

	CAAM_RTENT0  = 0x640
	CAAM_RTENT15 = 0x67c

	CAAM_C0CWR = 0x8044
	C0CWR_C1M  = 0
)

// CAAM represents the Cryptographic Acceleration and Assurance Module
// instance.
type CAAM struct {
	sync.Mutex

	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int

	// DeriveKeyMemory represents the DMA memory region where the CAAM blob
	// key encryption key (BKEK), derived from the hardware unique key, is
	// placed to derive diversified keys. The memory region must be
	// initialized before DeriveKey().
	//
	// When BEE is not used to encrypt external RAM it is recommended to
	// use a DMA region within the internal RAM (e.g. i.MX6 On-Chip
	// OCRAM/iRAM).
	//
	// The DeriveKey() function uses DeriveKeyMemory only if the default
	// DMA region start does not overlap with it.
	DeriveKeyMemory *dma.Region

	// Disable Timing Equalization protections (when supported)
	DisableTimingEqualization bool

	// control registers
	scfgr   uint32
	rtmctl  uint32
	rtent0  uint32
	rtent15 uint32

	// current RTENTa register
	rtenta uint32

	// default job ring
	jr *jobRing
}

// Init initializes the CAAM module.
func (hw *CAAM) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid CAAM instance")
	}

	hw.scfgr = hw.Base + CAAM_SCFGR
	hw.rtmctl = hw.Base + CAAM_RTMCTL
	hw.rtent0 = hw.Base + CAAM_RTENT0
	hw.rtent15 = hw.Base + CAAM_RTENT15

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// enter program mode
	reg.Set(hw.rtmctl, RTMCTL_PRGM)
	// reset defaults
	reg.Set(hw.rtmctl, RTMCTL_RST_DEF)

	// enable entropy generation
	hw.rtenta = hw.rtent0

	// force entropy re-generation
	reg.Set(hw.rtmctl, RTMCTL_TRNG_ACC)
	defer reg.Clear(hw.rtmctl, RTMCTL_TRNG_ACC)

	// disable RNG deterministic mode
	reg.Set(hw.scfgr, SCFGR_RNGSH0)
	// enable Random Differential Power Analysis Resistance
	reg.Set(hw.scfgr, SCFGR_RANDDPAR)

	// enter run mode
	reg.Clear(hw.rtmctl, RTMCTL_PRGM)
}
