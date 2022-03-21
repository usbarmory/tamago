// BCM2835 SoC VideoCore support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import "encoding/binary"

const (
	GPU_MEMORY_FLAG_DISCARDABLE      = 1 << 0
	GPU_MEMORY_FLAG_NORMAL           = 0 << 2
	GPU_MEMORY_FLAG_DIRECT           = 1 << 2
	GPU_MEMORY_FLAG_COHERENT         = 2 << 2
	GPU_MEMORY_FLAG_L1_NONALLOCATING = 3 << 2
	GPU_MEMORY_FLAG_ZERO             = 1 << 4
	GPU_MEMORY_FLAG_NO_INIT          = 1 << 5
	GPU_MEMORY_FLAG_HINT_PERMALOCK   = 1 << 6
)

// FirmwareRevision gets the firmware rev of the VideoCore GPU
func FirmwareRevision() uint32 {
	buf := exchangeSingleTagMessage(VC_BOARD_GET_REV, make([]byte, VC_BOARD_GET_REV_LEN))

	if len(buf) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buf)
}

// BoardModel gets the board model
func BoardModel() uint32 {
	buf := exchangeSingleTagMessage(VC_BOARD_GET_MODEL, make([]byte, VC_BOARD_GET_MODEL_LEN))

	if len(buf) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buf)
}

// MACAddress gets the board's MAC address
func MACAddress() []byte {
	return exchangeSingleTagMessage(VC_BOARD_GET_MAC, make([]byte, VC_BOARD_GET_MAC_LEN))
}

// Serial gets the board's serial number
func Serial() uint32 {
	buf := exchangeSingleTagMessage(VC_BOARD_GET_SERIAL, make([]byte, VC_BOARD_GET_SERIAL_LEN))

	if len(buf) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buf)
}

// CPUMemory gets the memory ranges allocated to the ARM core(s)
func CPUMemory() (start uint32, size uint32) {
	buf := exchangeSingleTagMessage(VC_BOARD_GET_ARM_MEMORY, make([]byte, VC_BOARD_GET_ARM_MEMORY_LEN))

	if len(buf) < 8 {
		return 0, 0
	}

	return binary.LittleEndian.Uint32(buf[0:]), binary.LittleEndian.Uint32(buf[4:])
}

// GPUMemory gets the memory ranges allocated to VideoCore
func GPUMemory() (start uint32, size uint32) {
	buf := exchangeSingleTagMessage(VC_BOARD_GET_VC_MEMORY, make([]byte, VC_BOARD_GET_VC_MEMORY_LEN))

	if len(buf) < 8 {
		return 0, 0
	}

	return binary.LittleEndian.Uint32(buf[0:]), binary.LittleEndian.Uint32(buf[4:])
}

// CPUAvailableDMAChannels gets the DMA channels available to the ARM core(s)
func CPUAvailableDMAChannels() (bitmask uint32) {
	buf := exchangeSingleTagMessage(VC_RES_GET_DMACHANNELS, make([]byte, VC_RES_GET_DMACHANNELS_LEN))

	if len(buf) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buf)
}

// AllocateGPUMemory allocates space from the GPU address space
//
// The returned value is a handle, use LockMemory to convert
// to an address.
func AllocateGPUMemory(size uint32, alignment uint32, flags uint32) (handle uint32) {
	buf := make([]byte, VC_MEM_ALLOCATE_LEN)
	binary.LittleEndian.PutUint32(buf[0:], size)
	binary.LittleEndian.PutUint32(buf[4:], alignment)
	binary.LittleEndian.PutUint32(buf[8:], uint32(flags))

	buf = exchangeSingleTagMessage(VC_MEM_ALLOCATE, buf)

	if len(buf) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buf)
}

// LockGPUMemory provides the address of previously allocated memory
func LockGPUMemory(handle uint32) (addr uint32) {
	buf := make([]byte, VC_MEM_LOCK_LEN)
	binary.LittleEndian.PutUint32(buf[0:], handle)

	buf = exchangeSingleTagMessage(VC_MEM_LOCK, buf)

	if len(buf) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(buf)
}

func exchangeSingleTagMessage(code uint32, buf []byte) []byte {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     code,
				Buffer: buf,
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(code)

	if tag == nil {
		return nil
	}

	return tag.Buffer
}
