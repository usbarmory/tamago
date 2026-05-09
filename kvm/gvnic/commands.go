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
)

// Additional AdminQ opcodes from gve_adminq.h.
const (
	ADMINQ_UNREGISTER_PAGE_LIST        = 0x4
	ADMINQ_DESTROY_TX_QUEUE            = 0x7
	ADMINQ_DESTROY_RX_QUEUE            = 0x8
	ADMINQ_DECONFIGURE_DEVICE_RESOURCE = 0x9
	ADMINQ_SET_DRIVER_PARAMETER        = 0xB
	ADMINQ_REPORT_STATS                = 0xC
	ADMINQ_REPORT_LINK_SPEED           = 0xD
)

type destroyQueueCommand struct {
	QueueID uint32
}

type unregisterPageListCommand struct {
	PageListID uint32
}

// destroyTxQueue issues ADMINQ_DESTROY_TX_QUEUE.
func (hw *GVE) destroyTxQueue(queueID uint32) error {
	cmd := &destroyQueueCommand{QueueID: queueID}
	return hw.aq.Push(ADMINQ_DESTROY_TX_QUEUE, cmd)
}

// destroyRxQueue issues ADMINQ_DESTROY_RX_QUEUE.
func (hw *GVE) destroyRxQueue(queueID uint32) error {
	cmd := &destroyQueueCommand{QueueID: queueID}
	return hw.aq.Push(ADMINQ_DESTROY_RX_QUEUE, cmd)
}

// unregisterPageList issues ADMINQ_UNREGISTER_PAGE_LIST.
func (hw *GVE) unregisterPageList(pageListID uint32) error {
	cmd := &unregisterPageListCommand{PageListID: pageListID}
	return hw.aq.Push(ADMINQ_UNREGISTER_PAGE_LIST, cmd)
}

// deconfigureDeviceResources issues ADMINQ_DECONFIGURE_DEVICE_RESOURCE.
func (hw *GVE) deconfigureDeviceResources() error {
	return hw.aq.Push(ADMINQ_DECONFIGURE_DEVICE_RESOURCE, nil)
}

// parseDeviceOptions reads device options following the 40-byte descriptor
// in a DESCRIBE_DEVICE response buffer.
func parseDeviceOptions(buf []byte, count int) []DeviceOption {
	var opts []DeviceOption
	off := 0
	for i := 0; i < count && off+8 <= len(buf); i++ {
		opt := DeviceOption{
			OptionID:             binary.BigEndian.Uint16(buf[off : off+2]),
			OptionLength:         binary.BigEndian.Uint16(buf[off+2 : off+4]),
			RequiredFeaturesMask: binary.BigEndian.Uint32(buf[off+4 : off+8]),
		}
		opts = append(opts, opt)

		// Advance past header + payload (aligned to 8 bytes)
		payloadLen := int(opt.OptionLength)
		off += 8 + payloadLen

		// Align to 8-byte boundary
		if off%8 != 0 {
			off += 8 - (off % 8)
		}
	}
	return opts
}

// supportsGQI_QPL returns true if the device advertises the GQI-QPL datapath
// as a device option.
func supportsGQI_QPL(opts []DeviceOption) bool {
	for _, opt := range opts {
		if opt.OptionID == DevOptGqiQPL {
			return true
		}
	}
	return false
}

// statusString returns a human-readable string for an AdminQ status code.
func statusString(status uint32) string {
	switch status {
	case COMMAND_PASSED:
		return "passed"
	case COMMAND_ERROR_ABORTED:
		return "aborted"
	case COMMAND_ERROR_ALREADY_EXISTS:
		return "already exists"
	case COMMAND_ERROR_CANCELLED:
		return "cancelled"
	case COMMAND_ERROR_DATALOSS:
		return "data loss"
	case COMMAND_ERROR_DEADLINE_EXCEEDED:
		return "deadline exceeded"
	case COMMAND_ERROR_FAILED_PRECONDITION:
		return "failed precondition"
	case COMMAND_ERROR_INTERNAL_ERROR:
		return "internal error"
	case COMMAND_ERROR_INVALID_ARGUMENT:
		return "invalid argument"
	case COMMAND_ERROR_NOT_FOUND:
		return "not found"
	case COMMAND_ERROR_OUT_OF_RANGE:
		return "out of range"
	case COMMAND_ERROR_PERMISSION_DENIED:
		return "permission denied"
	case COMMAND_ERROR_UNAUTHENTICATED:
		return "unauthenticated"
	case COMMAND_ERROR_RESOURCE_EXHAUSTED:
		return "resource exhausted"
	case COMMAND_ERROR_UNAVAILABLE:
		return "unavailable"
	case COMMAND_ERROR_UNIMPLEMENTED:
		return "unimplemented"
	case COMMAND_ERROR_UNKNOWN_ERROR:
		return "unknown error"
	default:
		return fmt.Sprintf("unknown status %#x", status)
	}
}
