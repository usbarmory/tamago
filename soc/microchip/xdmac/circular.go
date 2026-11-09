// Microchip Extensible DMA Controller (XDMAC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package xdmac

import (
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// linked list descriptor microblock control member
const (
	MBR_UBC_NVIEW = 27
	MBR_UBC_NDEN  = 26
	MBR_UBC_NSEN  = 25
	MBR_UBC_NDE   = 24
)

const (
	// next descriptor view 1: next descriptor, microblock control, source
	// and destination addresses
	descriptorView1 = 1
	descriptorSize  = 16

	// attempts at reading a consistent transfer position
	positionRetries = 5
)

type circular struct {
	region     *dma.Region
	descriptor uint
	size       int
}

// StartCircular starts a transfer on channel ch from the peripheral register
// src into the size bytes at dst, one byte for each peripheral request. After
// its last byte the transfer starts over at dst, until [XDMAC.Stop].
//
// The channel stays reserved for the transfer and [XDMAC.Position] reports
// its progress, the caller is responsible for cache maintenance of dst.
func (hw *XDMAC) StartCircular(ch, perid int, dst, src uint32, size int) (err error) {
	if perid < 0 || perid >= noPeripheral {
		return errors.New("invalid peripheral request line")
	}

	if err = hw.check(ch, dst, src, size); err != nil {
		return
	}

	if size == 0 || size > maxUnits {
		return errors.New("invalid circular transfer size")
	}

	if hw.Region == nil {
		hw.Region = dma.Default()
	}

	if hw.Region == nil {
		return errors.New("invalid DMA region")
	}

	hw.channels[ch].Lock()

	defer func() {
		if err != nil {
			hw.channels[ch].Unlock()
		}
	}()

	if err = hw.quiesce(ch, hw.timeout()); err != nil {
		return
	}

	// one descriptor that links to itself, for a block without end
	addr := hw.Region.Alloc(make([]byte, descriptorSize), 8)

	if uint64(addr)+descriptorSize > 1<<32 {
		hw.Region.Free(addr)
		return errors.New("descriptor exceeds the 32-bit bus")
	}

	var ubc uint32
	bits.SetN(&ubc, 0, maxUnits, uint32(size))
	bits.Set(&ubc, MBR_UBC_NDE)
	bits.Set(&ubc, MBR_UBC_NSEN)
	bits.Set(&ubc, MBR_UBC_NDEN)
	bits.SetN(&ubc, MBR_UBC_NVIEW, 0b11, descriptorView1)

	descriptor := make([]byte, descriptorSize)
	binary.LittleEndian.PutUint32(descriptor[0:], uint32(addr))
	binary.LittleEndian.PutUint32(descriptor[4:], ubc)
	binary.LittleEndian.PutUint32(descriptor[8:], src)
	binary.LittleEndian.PutUint32(descriptor[12:], dst)
	hw.Region.Write(addr, 0, descriptor)

	// fixed source and incrementing destination addresses, peripheral
	// synchronized transfer, Non-secure transactions
	var config uint32
	bits.SetN(&config, CC_PERID, 0x7f, uint32(perid))
	bits.SetN(&config, CC_DAM, 0b11, 1)
	bits.Set(&config, CC_PROT)
	bits.Set(&config, CC_TYPE)

	var cndc uint32
	bits.SetN(&cndc, CNDC_NDVIEW, 0b11, descriptorView1)
	bits.Set(&cndc, CNDC_NDDUP)
	bits.Set(&cndc, CNDC_NDSUP)
	bits.Set(&cndc, CNDC_NDE)

	channel := hw.channel(ch)
	reg.Write(channel+XDMAC_CBC, 0)
	reg.Write(channel+XDMAC_CDS_MSP, 0)
	reg.Write(channel+XDMAC_CSUS, 0)
	reg.Write(channel+XDMAC_CDUS, 0)
	reg.Write(channel+XDMAC_CNDA, uint32(addr))
	reg.Write(channel+XDMAC_CC, config)
	reg.Write(channel+XDMAC_CNDC, cndc)
	reg.Write(channel+XDMAC_CID, 0xff)
	reg.Read(channel + XDMAC_CIS)

	hw.circular[ch] = circular{region: hw.Region, descriptor: addr, size: size}
	reg.Write(hw.Base+XDMAC_GE, 1<<ch)

	return
}

// Position returns the offset at the buffer of the circular transfer on
// channel ch of the next byte that it writes, the bytes before it have reached
// memory.
func (hw *XDMAC) Position(ch int) (off int, err error) {
	if ch < 0 || ch >= Channels || hw.circular[ch].size == 0 {
		return 0, errors.New("invalid circular transfer")
	}

	size := hw.circular[ch].size
	channel := hw.channel(ch)

	// the controller counts bytes it still holds as transferred
	if err = hw.flush(ch); err != nil {
		return
	}

	for range positionRetries {
		// the remaining length reads as complete until the descriptor
		// fetch at the end of the buffer terminates
		ubc := reg.Read(channel+XDMAC_CUBC) & maxUnits

		if !reg.Get(channel+XDMAC_CC, CC_INITD) {
			continue
		}

		if err = hw.flush(ch); err != nil {
			return
		}

		return (size - int(ubc)) % size, nil
	}

	return 0, fmt.Errorf("channel %d position unavailable", ch)
}

// Stop stops the circular transfer on channel ch.
func (hw *XDMAC) Stop(ch int) (err error) {
	if ch < 0 || ch >= Channels || hw.circular[ch].size == 0 {
		return errors.New("invalid circular transfer")
	}

	if err = hw.quiesce(ch, hw.timeout()); err != nil {
		// the channel can still access its descriptor
		return
	}

	hw.circular[ch].region.Free(hw.circular[ch].descriptor)
	hw.circular[ch] = circular{}
	hw.channels[ch].Unlock()

	return
}

// flush writes the bytes channel ch holds to memory.
func (hw *XDMAC) flush(ch int) error {
	channel := hw.channel(ch)
	deadline := time.Now().Add(hw.timeout())

	reg.Write(hw.Base+XDMAC_GSWF, 1<<ch)

	for {
		status := reg.Read(channel + XDMAC_CIS)

		switch {
		case status&CIS_ERRORS != 0:
			return fmt.Errorf("channel %d transfer error, status %#x", ch, status)
		case status&(1<<CIS_FIS) != 0:
			return nil
		case time.Now().After(deadline):
			return fmt.Errorf("channel %d flush timeout", ch)
		}

		runtime.Gosched()
	}
}

func (hw *XDMAC) timeout() time.Duration {
	if hw.Timeout == 0 {
		return Timeout
	}

	return hw.Timeout
}
