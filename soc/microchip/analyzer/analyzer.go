// Microchip Analyzer (ANA)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package analyzer implements a driver for the Microchip Analyzer block
// (Analyzer), responsible for ingress frame processing, adopting the following
// reference specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package analyzer

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// Group identifiers
const (
	// flood section forwarding offsets (address:ports+pgid)
	PGID_UNICAST   = 0
	PGID_MULTICAST = 1
	PGID_BROADCAST = 2
	PGID_HOST      = 3

	// VLAN filtering identifier
	MAC_VID = 1
)

// Port Group registers
const (
	PGID          = 0x20000
	PGID_CFG      = 0x00
	PGID_MISC_CFG = 0x0c
)

// Learning block registers
const (
	COMMON_ACCESS_CTRL         = 0x00
	CTRL_CPU_ACCESS_CMD        = 1
	CTRL_MAC_TABLE_ACCESS_SHOT = 0

	CMD_LEARN         = 0
	CMD_FORGET        = 1
	CMD_LOOKUP        = 2
	CMD_READ_DIRECT   = 3
	CMD_WRITE_DIRECT  = 4
	CMD_SCAN          = 5
	CMD_FIND_SMALLEST = 6
	CMD_CLEAR_ALL     = 7

	MAC_ACCESS_CFG_0    = 0x04
	CFG_0_MAC_ENTRY_FID = 16
	CFG_0_MAC_ENTRY_MSB = 0

	MAC_ACCESS_CFG_1    = 0x08
	CFG_1_MAC_ENTRY_LSB = 0

	MAC_ACCESS_CFG_2          = 0x0c
	CFG_2_MAC_ENTRY_CPU_QU    = 24
	CFG_2_MAC_ENTRY_CPU_COPY  = 23
	CFG_2_MAC_ENTRY_LOCKED    = 16
	CFG_2_MAC_ENTRY_VLD       = 15
	CFG_2_MAC_ENTRY_ADDR_TYPE = 12
	CFG_2_MAC_ENTRY_ADDR      = 0

	TYPE_UNICAST      = 0 // UPSID_PN
	TYPE_MANAGEMENT   = 1 // GCPU_UPS
	TYPE_UNICAST_GLAG = 2 // GLAG
	TYPE_MULTICAST    = 3 // MC_IDX
)

// Timeout is the default timeout for learn operations.
const Timeout = 2 * time.Second

// ANA represents the Analyzer block.
type ANA struct {
	sync.Mutex

	// Access Control block base register
	AccessControl uint32
	// Learn Block base register
	Learn uint32
	// Timeout for learn operations
	Timeout time.Duration

	// Ports represents the number of physical switch ports
	Ports int
}

// Init initializes the analyzer instance.
func (a *ANA) Init() (err error) {
	if a.AccessControl == 0 || a.Learn == 0 {
		return errors.New("invalid analyzer instance")
	}

	if a.Timeout == 0 {
		a.Timeout = Timeout
	}

	// forward broadcast frames to all ports
	pgid := uint32(a.Ports) + PGID_BROADCAST
	reg.Write(a.AccessControl+PGID+(pgid*16)+PGID_CFG, 0xffffffff)

	// isolate host frames to CPU queue
	pgid = uint32(a.Ports) + PGID_HOST
	reg.Write(a.AccessControl+PGID+(pgid*16)+PGID_CFG, 0x00000000)

	return
}

// Learning issues a MAC table access command.
func (a *ANA) Learning(mac net.HardwareAddr, addrType, vid, pgid, cmd uint32) {
	a.Lock()
	defer a.Unlock()

	if a.AccessControl == 0 {
		return
	}

	if len(mac) != 6 {
		return
	}

	msb := uint32(mac[1])
	msb |= uint32(mac[0]) << 8

	lsb := uint32(mac[5])
	lsb |= uint32(mac[4]) << 8
	lsb |= uint32(mac[3]) << 16
	lsb |= uint32(mac[2]) << 24

	// set MAC address
	var cfg0 uint32
	bits.SetN(&cfg0, CFG_0_MAC_ENTRY_FID, 0x1fff, vid)
	bits.SetN(&cfg0, CFG_0_MAC_ENTRY_MSB, 0xffff, msb)
	reg.Write(a.Learn+MAC_ACCESS_CFG_0, cfg0)
	reg.Write(a.Learn+MAC_ACCESS_CFG_1, lsb)

	// set forwarding entry handling
	var cfg2 uint32
	bits.SetN(&cfg2, CFG_2_MAC_ENTRY_CPU_QU, 0b111, 0) // queue number
	bits.Set(&cfg2, CFG_2_MAC_ENTRY_CPU_COPY)          // copy forwarding entry to queue
	bits.Set(&cfg2, CFG_2_MAC_ENTRY_LOCKED)
	bits.Set(&cfg2, CFG_2_MAC_ENTRY_VLD)
	bits.SetN(&cfg2, CFG_2_MAC_ENTRY_ADDR_TYPE, 0b111, addrType)
	bits.SetN(&cfg2, CFG_2_MAC_ENTRY_ADDR, 0xfff, pgid)
	reg.Write(a.Learn+MAC_ACCESS_CFG_2, cfg2)

	// issue command
	reg.SetN(a.Learn+COMMON_ACCESS_CTRL, CTRL_CPU_ACCESS_CMD, 0xf, cmd)
	reg.Set(a.Learn+COMMON_ACCESS_CTRL, CTRL_MAC_TABLE_ACCESS_SHOT)
	reg.WaitFor(Timeout, a.Learn+COMMON_ACCESS_CTRL, CTRL_MAC_TABLE_ACCESS_SHOT, 1, 0)
}

// Insert adds a new entry in the MAC table.
func (a *ANA) Insert(mac net.HardwareAddr, vid, pgid uint32) {
	a.Learning(mac, TYPE_MULTICAST, vid, pgid, CMD_LEARN)
}

// Delete removes an entry from the MAC table.
func (a *ANA) Delete(mac net.HardwareAddr, vid, pgid uint32) {
	a.Learning(mac, TYPE_MULTICAST, vid, pgid, CMD_FORGET)
}
