// NXP Bus Encryption Engine (BEE) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bee implements a driver for the NXP Bus Encryption Engine (BEE)
// adopting the following reference specifications:
//   - IMX6ULSRM - i.MX6UL Security Reference Manual - Rev 0 04/2016
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package bee

import (
	"crypto/aes"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/nxp/snvs"
)

// BEE registers
const (
	BEE_CTRL                 = 0x00
	CTRL_CLK_EN_LOCK         = 31
	CTRL_SFTRST_N_LOCK       = 30
	CTRL_AES_MODE_LOCK       = 29
	CTRL_SECURITY_LEVEL_LOCK = 24
	CTRL_AES_KEY_SEL_LOCK    = 20
	CTRL_BEE_ENABLE_LOCK     = 16
	CTRL_CLK_EN              = 15
	CTRL_SFTRST_N            = 14
	CTRL_AES_MODE            = 13
	CTRL_SECURITY_LEVEL      = 8
	CTRL_AES_KEY_SEL         = 4
	CTRL_BEE_ENABLE          = 0

	BEE_ADDR_OFFSET0 = 0x04
	BEE_ADDR_OFFSET1 = 0x08

	// AES key
	BEE_AES_KEY0_W0 = 0x0c
	BEE_AES_KEY0_W1 = 0x10
	BEE_AES_KEY0_W2 = 0x14
	BEE_AES_KEY0_W3 = 0x18

	// AES CTR nonce
	BEE_AES_KEY1_W0 = 0x20
	BEE_AES_KEY1_W1 = 0x24
	BEE_AES_KEY1_W2 = 0x28
	BEE_AES_KEY1_W3 = 0x2c
)

const (
	AliasRegion0    = 0x10000000
	AliasRegion1    = 0x30000000
	AliasRegionSize = 0x20000000
)

// BEE represents the Bus Encryption Engine instance.
type BEE struct {
	mu sync.Mutex

	// Base register
	Base uint32
	// SNVS instance
	SNVS *snvs.SNVS

	// control registers
	ctrl  uint32
	addr0 uint32
	addr1 uint32

	// AES key pointer
	key uint32
	// AES CTR nonce pointer
	nonce uint32
}

// Init initializes the BEE module.
func (hw *BEE) Init() {
	hw.mu.Lock()
	defer hw.mu.Unlock()

	if hw.Base == 0 {
		panic("invalid BEE instance")
	}

	hw.ctrl = hw.Base + BEE_CTRL
	hw.addr0 = hw.Base + BEE_ADDR_OFFSET0
	hw.addr1 = hw.Base + BEE_ADDR_OFFSET1
	hw.key = hw.Base + BEE_AES_KEY0_W0
	hw.nonce = hw.Base + BEE_AES_KEY1_W0

	// enable clock
	reg.Set(hw.ctrl, CTRL_CLK_EN)
	// disable reset
	reg.Set(hw.ctrl, CTRL_SFTRST_N)

	// disable
	reg.Clear(hw.ctrl, CTRL_BEE_ENABLE)
}

func (hw *BEE) generateKey(ptr uint32) (err error) {
	// avoid key exposure to external RAM
	key, err := dma.NewRegion(uint(ptr), aes.BlockSize, false)

	if err != nil {
		return
	}

	addr, buf := key.Reserve(aes.BlockSize, 0)

	if n, err := rand.Read(buf); n != aes.BlockSize || err != nil {
		return errors.New("could not set random key")
	}

	if addr != uint(ptr) {
		return errors.New("invalid key address")
	}

	return
}

func checkRegion(region uint32, offset uint32) error {
	if region&0xffff != 0 {
		return errors.New("address must be 64KB aligned")
	}

	if offset >= AliasRegion0 && offset < AliasRegion0+AliasRegionSize ||
		offset >= AliasRegion1 && offset < AliasRegion1+AliasRegionSize {
		return errors.New("invalid region (offset overalps with aliased region)")
	}

	return nil
}

// Enable activates the BEE using the argument memory regions, each can be up
// to AliasRegionSize (512 MB) in size.
//
// After activation the regions are encrypted using AES CTR. On secure booted
// systems the internal OTPMK is used as key, otherwise a random one is
// generated and assigned.
//
// After enabling, both regions should only be accessed through their
// respective aliased spaces (see AliasRegion0 and AliasRegion1) and only with
// caching enabled (see arm.ConfigureMMU).
func (hw *BEE) Enable(region0 uint32, region1 uint32) (err error) {
	hw.mu.Lock()
	defer hw.mu.Unlock()

	if err = checkRegion(region0, region0-AliasRegion0); err != nil {
		return fmt.Errorf("region0 error: %v", err)
	}

	if err = checkRegion(region1, region1-AliasRegion1); err != nil {
		return fmt.Errorf("region1 error: %v", err)
	}

	if hw.addr0 == 0 {
		return errors.New("invalid BEE instance")
	}

	reg.Write(hw.addr0, region0>>16)
	reg.Write(hw.addr1, region1>>16)

	if hw.SNVS != nil && hw.SNVS.Available() {
		// use OTPMK if SNVS is secure booted
		reg.Clear(hw.ctrl, CTRL_AES_KEY_SEL)
	} else {
		// set random AES key
		if err = hw.generateKey(hw.key); err != nil {
			return
		}

		// select software AES key
		reg.Set(hw.ctrl, CTRL_AES_KEY_SEL)
	}

	// set random nonce for CTR mode
	if err = hw.generateKey(hw.nonce); err != nil {
		return
	}

	// set AES CTR mode
	reg.Set(hw.ctrl, CTRL_AES_MODE)
	// set maximum security level
	reg.SetN(hw.ctrl, CTRL_SECURITY_LEVEL, 0b11, 3)

	// enable
	reg.Set(hw.ctrl, CTRL_BEE_ENABLE)

	return
}

// Lock restricts BEE registers writing.
func (hw *BEE) Lock() {
	hw.mu.Lock()
	defer hw.mu.Unlock()

	reg.Set(hw.ctrl, CTRL_CLK_EN_LOCK)
	reg.Set(hw.ctrl, CTRL_SFTRST_N_LOCK)
	reg.Set(hw.ctrl, CTRL_AES_MODE_LOCK)
	reg.SetN(hw.ctrl, CTRL_SECURITY_LEVEL_LOCK, 0b11, 0b11)
	reg.Set(hw.ctrl, CTRL_AES_KEY_SEL_LOCK)
	reg.Set(hw.ctrl, CTRL_BEE_ENABLE_LOCK)

	reg.SetN(hw.addr0, 16, 0xffff, 0xffff)
	reg.SetN(hw.addr1, 16, 0xffff, 0xffff)
}
