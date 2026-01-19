// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// CommandTimeout is the timeout for commands sent to the admin queue.
var CommandTimeout = 1 * time.Second

// Admin queue status codes
const (
	COMMAND_UNSET                     = 0x0
	COMMAND_PASSED                    = 0x1
	COMMAND_ERROR_ABORTED             = 0xfffffff0
	COMMAND_ERROR_ALREADY_EXISTS      = 0xfffffff1
	COMMAND_ERROR_CANCELLED           = 0xfffffff2
	COMMAND_ERROR_DATALOSS            = 0xfffffff3
	COMMAND_ERROR_DEADLINE_EXCEEDED   = 0xfffffff4
	COMMAND_ERROR_FAILED_PRECONDITION = 0xfffffff5
	COMMAND_ERROR_INTERNAL_ERROR      = 0xfffffff6
	COMMAND_ERROR_INVALID_ARGUMENT    = 0xfffffff7
	COMMAND_ERROR_NOT_FOUND           = 0xfffffff8
	COMMAND_ERROR_OUT_OF_RANGE        = 0xfffffff9
	COMMAND_ERROR_PERMISSION_DENIED   = 0xfffffffa
	COMMAND_ERROR_UNAUTHENTICATED     = 0xfffffffb
	COMMAND_ERROR_RESOURCE_EXHAUSTED  = 0xfffffffc
	COMMAND_ERROR_UNAVAILABLE         = 0xfffffffd
	COMMAND_ERROR_UNIMPLEMENTED       = 0xfffffffe
	COMMAND_ERROR_UNKNOWN_ERROR       = 0xffffffff
)

const (
	pageSize       = 4096
	adminQueueSize = pageSize
	hdrSize        = 8
	cmdSize        = 64
)

type adminCommandHeader struct {
	Opcode uint32
	Status uint32
}

type adminQueue struct {
	// control registers
	Doorbell uint32
	Counter  uint32

	// internal counter
	cnt uint32

	// DMA buffer
	addr uint
	buf  []byte
}

func (aq *adminQueue) Push(opcode uint32, cmd any) (err error) {
	ac := &adminCommandHeader{
		Opcode: opcode,
	}

	low := aq.cnt * cmdSize
	high := low + cmdSize

	if _, err = binary.Encode(aq.buf[low:high], binary.BigEndian, ac); err != nil {
		return
	}

	if _, err = binary.Encode(aq.buf[low+hdrSize:high], binary.BigEndian, cmd); err != nil {
		return
	}

	// zero out remaining buffer
	pad := cmdSize - hdrSize - binary.Size(cmd)
	copy(aq.buf[high-uint32(pad):high], make([]byte, pad))

	// bump counter and ring door bell
	aq.cnt = (aq.cnt + 1) % (adminQueueSize / cmdSize)
	cnt := bits.ReverseBytes32(aq.cnt)
	reg.Write(aq.Doorbell, cnt)

	if !reg.WaitFor(CommandTimeout, aq.Counter, 0, 0xffffffff, cnt) {
		return fmt.Errorf("admin queue timeout")
	}

	if status := binary.BigEndian.Uint32(aq.buf[low+4 : low+8]); status != COMMAND_PASSED {
		return fmt.Errorf("admin command error, status:%#x", status)
	}

	return
}

func (hw *GVE) initAdminQueue() (err error) {
	hw.aq = &adminQueue{
		Doorbell: hw.registers + ADMINQ_DOORBELL,
		Counter:  hw.registers + ADMINQ_EVENT_COUNTER,
	}

	hw.aq.addr, hw.aq.buf = hw.QueueRegion.Reserve(adminQueueSize, 0)

	if hw.Device.Revision < 1 {
		// set admin queue based address to region page frame number
		hw.set(ADMINQ_PFN, uint32(hw.aq.addr/pageSize))
	} else {
		// set admin queue based address and size
		hw.set(ADMINQ_LENGTH, uint16(adminQueueSize))
		hw.set(ADMINQ_BASE_ADDRESS_LOW, uint32(hw.aq.addr))
	}

	return
}
