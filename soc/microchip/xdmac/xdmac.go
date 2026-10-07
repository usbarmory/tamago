// Microchip Extensible DMA Controller (XDMAC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package xdmac implements a driver for the Microchip Extensible DMA
// Controller (XDMAC) adopting the following specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package xdmac

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// XDMAC registers
const (
	XDMAC_GIE = 0x0c
	XDMAC_GID = 0x10
	XDMAC_GIS = 0x18
	XDMAC_GE  = 0x1c
	XDMAC_GD  = 0x20
	XDMAC_GS  = 0x24

	XDMAC_CH      = 0x60
	XDMAC_CH_SIZE = 0x40

	XDMAC_CIE = 0x00
	XDMAC_CID = 0x04

	XDMAC_CIS  = 0x0c
	CIS_BIS    = 0
	CIS_RBEIS  = 4
	CIS_WBEIS  = 5
	CIS_ROIS   = 6
	CIS_TCIS   = 7
	CIS_ERRORS = 1<<CIS_RBEIS | 1<<CIS_WBEIS | 1<<CIS_ROIS | 1<<CIS_TCIS

	XDMAC_CSA  = 0x10
	XDMAC_CDA  = 0x14
	XDMAC_CNDA = 0x18
	XDMAC_CNDC = 0x1c
	XDMAC_CUBC = 0x20
	XDMAC_CBC  = 0x24

	XDMAC_CC  = 0x28
	CC_PERID  = 24
	CC_DAM    = 18
	CC_SAM    = 16
	CC_DWIDTH = 11
	CC_CSIZE  = 8
	CC_PROT   = 5
	CC_MBSIZE = 1

	XDMAC_CDS_MSP = 0x2c
	XDMAC_CSUS    = 0x30
	XDMAC_CDUS    = 0x34
)

const (
	// Channels is the number of XDMAC channels.
	Channels = 16

	// maximum microblock length in data units
	maxUnits = 0xffffff

	// memory burst and chunk of sixteen data units
	burst16 = 3
	chunk16 = 4

	// peripheral identifier for memory-to-memory transfers
	noPeripheral = 0x7f

	// polling time before waiting for the channel interrupt
	interruptWaitThreshold = 100 * time.Microsecond
)

// Timeout is the default transfer timeout.
const Timeout = 1 * time.Second

// InterruptPollInterval is the maximum time between channel status checks
// when waiting for channel interrupts.
var InterruptPollInterval = 1 * time.Millisecond

// XDMAC represents an XDMAC controller instance.
type XDMAC struct {
	// Base register
	Base uint32
	// Interrupt ID
	IRQ int
	// Transfer timeout (default: Timeout)
	Timeout time.Duration

	channels [Channels]sync.Mutex
	events   [Channels]chan struct{}
}

// EnableInterrupt enables waiting for channel interrupts during transfers,
// [XDMAC.ServiceInterrupts] must then be called on each XDMAC interrupt.
func (hw *XDMAC) EnableInterrupt() {
	for ch := range Channels {
		hw.channels[ch].Lock()
		hw.events[ch] = make(chan struct{}, 1)
		hw.channels[ch].Unlock()
	}
}

// ServiceInterrupts handles an XDMAC interrupt by waking up the transfers
// waiting on its channels.
func (hw *XDMAC) ServiceInterrupts() {
	pending := reg.Read(hw.Base + XDMAC_GIS)
	reg.Write(hw.Base+XDMAC_GID, pending)

	for ch := range Channels {
		if pending&(1<<ch) == 0 {
			continue
		}

		select {
		case hw.events[ch] <- struct{}{}:
		default:
		}
	}
}

func (hw *XDMAC) channel(ch int) uint32 {
	return hw.Base + XDMAC_CH + uint32(ch)*XDMAC_CH_SIZE
}

func (hw *XDMAC) quiesce(ch int, timeout time.Duration) (err error) {
	reg.Write(hw.Base+XDMAC_GD, 1<<ch)
	reg.Write(hw.Base+XDMAC_GID, 1<<ch)

	if err = hw.wait(ch, timeout); err != nil {
		return fmt.Errorf("disable channel %d, %w", ch, err)
	}

	// channel registers are writable only once the channel is idle
	reg.Write(hw.channel(ch)+XDMAC_CID, 0xff)
	reg.Read(hw.channel(ch) + XDMAC_CIS)

	return
}

func (hw *XDMAC) wait(ch int, timeout time.Duration) error {
	if !reg.WaitFor(timeout, hw.Base+XDMAC_GS, ch, 1, 0) {
		return errors.New("channel busy timeout")
	}

	return nil
}

func (hw *XDMAC) waitTransfer(ch int, timeout time.Duration) (status uint32, err error) {
	started := time.Now()
	deadline := started.Add(timeout)

	for reg.Read(hw.Base+XDMAC_GS)&(1<<ch) != 0 {
		// status is read-to-clear, accumulate it
		status |= reg.Read(hw.channel(ch) + XDMAC_CIS)

		if status&CIS_ERRORS != 0 {
			return
		}

		now := time.Now()

		if now.After(deadline) {
			return status, errors.New("channel busy timeout")
		}

		if hw.events[ch] == nil || now.Sub(started) < interruptWaitThreshold {
			runtime.Gosched()
			continue
		}

		// servicing the interrupt masks the channel until the next wait
		reg.Write(hw.Base+XDMAC_GIE, 1<<ch)

		timer := time.NewTimer(max(min(time.Until(deadline), InterruptPollInterval), 0))

		select {
		case <-hw.events[ch]:
		case <-timer.C:
		}

		timer.Stop()
	}

	status |= reg.Read(hw.channel(ch) + XDMAC_CIS)

	return
}

// Copy copies size bytes from src to dst on channel ch and waits for its
// completion, the caller is responsible for cache maintenance.
func (hw *XDMAC) Copy(ch int, dst uint32, src uint32, size int) (err error) {
	if err = hw.check(ch, dst, src, size); err != nil || size == 0 {
		return
	}

	// widest data unit aligned to addresses and size
	width := uint32(3)

	for (uint32(size)|src|dst)&(1<<width-1) != 0 {
		width--
	}

	// incrementing source and destination addresses
	var config uint32
	bits.SetN(&config, CC_PERID, 0x7f, noPeripheral)
	bits.SetN(&config, CC_DAM, 0b11, 1)
	bits.SetN(&config, CC_SAM, 0b11, 1)
	bits.SetN(&config, CC_DWIDTH, 0b11, width)
	bits.SetN(&config, CC_CSIZE, 0b111, chunk16)
	bits.SetN(&config, CC_MBSIZE, 0b11, burst16)

	return hw.transfer(ch, config, dst, src, uint32(size)>>width, hw.Timeout)
}

func (hw *XDMAC) check(ch int, dst uint32, src uint32, size int) error {
	if hw.Base == 0 || ch < 0 || ch >= Channels {
		return errors.New("invalid instance or channel")
	}

	if size < 0 || uint64(src)+uint64(size) > 1<<32 || uint64(dst)+uint64(size) > 1<<32 {
		return errors.New("transfer exceeds the 32-bit bus")
	}

	return nil
}

func (hw *XDMAC) transfer(ch int, config uint32, dst uint32, src uint32, units uint32, timeout time.Duration) (err error) {
	hw.channels[ch].Lock()
	defer hw.channels[ch].Unlock()

	if units > maxUnits {
		return errors.New("transfer exceeds one microblock")
	}

	if timeout == 0 {
		timeout = Timeout
	}

	if err = hw.quiesce(ch, timeout); err != nil {
		return
	}

	// discard a wakeup left by an earlier transfer
	select {
	case <-hw.events[ch]:
	default:
	}

	// Non-secure transactions
	bits.Set(&config, CC_PROT)

	channel := hw.channel(ch)

	reg.Write(channel+XDMAC_CSA, src)
	reg.Write(channel+XDMAC_CDA, dst)
	reg.Write(channel+XDMAC_CNDA, 0)
	reg.Write(channel+XDMAC_CNDC, 0)
	reg.Write(channel+XDMAC_CBC, 0)
	reg.Write(channel+XDMAC_CDS_MSP, 0)
	reg.Write(channel+XDMAC_CSUS, 0)
	reg.Write(channel+XDMAC_CDUS, 0)
	reg.Write(channel+XDMAC_CUBC, units)
	reg.Write(channel+XDMAC_CC, config)
	reg.Write(channel+XDMAC_CIE, 1<<CIS_BIS|CIS_ERRORS)

	reg.Write(hw.Base+XDMAC_GE, 1<<ch)

	status, err := hw.waitTransfer(ch, timeout)

	if err != nil {
		remaining := reg.Read(channel + XDMAC_CUBC)
		hw.quiesce(ch, timeout)
		return fmt.Errorf("channel %d transfer, %w (%d units remaining, status %#x)", ch, err, remaining, status)
	}

	if err = hw.quiesce(ch, timeout); err != nil {
		return
	}

	switch {
	case status&CIS_ERRORS != 0:
		return fmt.Errorf("channel %d transfer error, status %#x", ch, status)
	case status&(1<<CIS_BIS) == 0:
		return fmt.Errorf("channel %d transfer incomplete, status %#x", ch, status)
	}

	return
}
