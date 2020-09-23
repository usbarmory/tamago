// BCM2835 SoC VideoCore support
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
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_BOARD_GET_REV,
				Buffer: make([]byte, 4),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_BOARD_GET_REV)

	if tag == nil || len(tag.Buffer) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:])
}

// BoardModel gets the board model
func BoardModel() uint32 {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_BOARD_GET_MODEL,
				Buffer: make([]byte, 4),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_BOARD_GET_MODEL)

	if tag == nil || len(tag.Buffer) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:])
}

// MACAddress gets the board's MAC address
func MACAddress() []byte {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_BOARD_GET_MAC,
				Buffer: make([]byte, 6),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_BOARD_GET_MAC)

	if tag == nil || len(tag.Buffer) < 4 {
		return []byte{}
	}

	return tag.Buffer
}

// Serial gets the board's serial number
func Serial() uint32 {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_BOARD_GET_SERIAL,
				Buffer: make([]byte, 4),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_BOARD_GET_SERIAL)

	if tag == nil || len(tag.Buffer) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:])
}

// CPUMemory gets the memory ranges allocated to the ARM core(s)
func CPUMemory() (uint32, uint32) {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_BOARD_GET_ARM_MEMORY,
				Buffer: make([]byte, 8),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_BOARD_GET_ARM_MEMORY)

	if tag == nil || len(tag.Buffer) < 8 {
		return 0, 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:]), binary.LittleEndian.Uint32(tag.Buffer[4:])
}

// GPUMemory gets the memory ranges allocated to VideoCore
func GPUMemory() (uint32, uint32) {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_BOARD_GET_VC_MEMORY,
				Buffer: make([]byte, 8),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_BOARD_GET_VC_MEMORY)

	if tag == nil || len(tag.Buffer) < 8 {
		return 0, 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:]), binary.LittleEndian.Uint32(tag.Buffer[4:])
}

// CPUAvailableDMAChannels gets the DMA channels available to the ARM core(s)
func CPUAvailableDMAChannels() uint32 {
	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_RES_GET_DMACHANNELS,
				Buffer: make([]byte, 4),
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_RES_GET_DMACHANNELS)

	if tag == nil || len(tag.Buffer) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:])
}

// AllocateGPUMemory allocates space from the GPU address space
//
// The returned value is a handle, use LockMemory to convert
// to an address.
//
func AllocateGPUMemory(size uint32, alignment uint32, flags uint32) uint32 {

	buf := make([]byte, 12)
	binary.LittleEndian.PutUint32(buf[0:], size)
	binary.LittleEndian.PutUint32(buf[4:], alignment)
	binary.LittleEndian.PutUint32(buf[8:], uint32(flags))

	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_MEM_ALLOCATE,
				Buffer: buf,
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_MEM_ALLOCATE)

	if tag == nil || len(tag.Buffer) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:])
}

// LockGPUMemory provides the address of previously allocated memory
//
func LockGPUMemory(handle uint32) uint32 {

	buf := make([]byte, 12)
	binary.LittleEndian.PutUint32(buf[0:], handle)

	msg := &MailboxMessage{
		Tags: []MailboxTag{
			{
				ID:     VC_MEM_LOCK,
				Buffer: buf,
			},
		},
	}

	Mailbox.Call(VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(VC_MEM_LOCK)

	if tag == nil || len(tag.Buffer) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:])
}
