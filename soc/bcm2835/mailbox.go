// BCM2835 SoC Mailbox support
// https://github.com/usbarmory/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//
// Mailboxes are used for inter-processor communication, in particular
// the VideoCore processor.
//

package bcm2835

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// We reserve the 'gap' above excStack and below TEXT segment start for
// mailbox usage.  There is nothing requiring use of this region, but it
// is convenient since it is always available.  Other regions could be
// used by adjusting ramSize to provide space at top of address range.
const (
	MAILBOX_REGION_BASE = 0xC000
	MAILBOX_REGION_SIZE = 0x4000
)

// Registers for using mailbox
const (
	MAILBOX_BASE       = 0xB880
	MAILBOX_READ_REG   = MAILBOX_BASE + 0x00
	MAILBOX_STATUS_REG = MAILBOX_BASE + 0x18
	MAILBOX_WRITE_REG  = MAILBOX_BASE + 0x20
	MAILBOX_FULL       = 0x80000000
	MAILBOX_EMPTY      = 0x40000000
)

type mailbox struct {
	Region *dma.Region
	Mutex  sync.Mutex
}

// Mailbox provides access to the BCM2835 mailbox used to communicate with
// the VideoCore CPU
var Mailbox = mailbox{}

func init() {
	// We don't use this region for DMA, but dma package provides a convenient
	// block allocation system.
	Mailbox.Region, _ = dma.NewRegion(MAILBOX_REGION_BASE|DRAM_FLAG_NOCACHE, MAILBOX_REGION_SIZE)
}

type MailboxTag struct {
	ID     uint32
	Buffer []byte
}

type MailboxMessage struct {
	MinSize int
	Code    uint32
	Tags    []MailboxTag
}

func (m *MailboxMessage) Error() bool {
	return m.Code == 0x80000001
}

func (m *MailboxMessage) Tag(code uint32) *MailboxTag {
	for i, tag := range m.Tags {
		if tag.ID&0x7FFFFFFF == code&0x7FFFFFFF {
			return &m.Tags[i]
		}
	}

	return nil
}

// Call exchanges message via a mailbox channel
//
// The caller is responsible for ensuring the 'tags' in the
// message have sufficient buffer allocated for the response
// expected.  The response replaces the input message.
func (mb *mailbox) Call(channel int, message *MailboxMessage) {
	size := 8 // Message Header
	for _, tag := range message.Tags {
		// 3 word tag header + tag data (padded to 32-bits)
		size += int(12 + (uint32(len(tag.Buffer)+3) & 0xFFFFFFFC))
	}

	size += 4 // null tag

	// Allow client to request bigger buffer for response
	if size < message.MinSize {
		size = message.MinSize
	}

	// Allocate temporary location-fixed buffer
	addr, buf := mb.Region.Reserve(size, 16)
	defer mb.Region.Release(addr)

	binary.LittleEndian.PutUint32(buf[0:], uint32(size))
	binary.LittleEndian.PutUint32(buf[4:], 0)

	offset := 8
	for _, tag := range message.Tags {
		binary.LittleEndian.PutUint32(buf[offset:], tag.ID)
		binary.LittleEndian.PutUint32(buf[offset+4:], uint32(len(tag.Buffer)))
		binary.LittleEndian.PutUint32(buf[offset+8:], 0)
		copy(buf[offset+12:], tag.Buffer)

		offset += int(12 + (uint32(len(tag.Buffer)+3) & 0xFFFFFFFC))
	}

	// terminating null tag
	binary.LittleEndian.PutUint32(buf[offset:], 0x0)

	mb.exchangeMessage(channel, addr)

	message.Tags = make([]MailboxTag, 0, len(message.Tags))
	message.Code = binary.LittleEndian.Uint32(buf[4:])
	offset = 8

	for offset < len(buf) {
		tag := MailboxTag{}
		tag.ID = binary.LittleEndian.Uint32(buf[offset:])

		// Terminating null tag
		if tag.ID == 0 {
			break
		}

		len := binary.LittleEndian.Uint32(buf[offset+4:])

		if len > uint32(size-offset) {
			panic("malformed mailbox response, over-sized tag")
		}

		tag.Buffer = make([]byte, len)
		copy(tag.Buffer, buf[offset+12:])

		// Move to next tag
		offset += int(12 + (len+3)&0xFFFFFFFC)
		message.Tags = append(message.Tags, tag)
	}
}

func (mb *mailbox) exchangeMessage(channel int, addr uint32) {
	if (addr & 0xF) != 0 {
		panic("Mailbox message must be 16-byte aligned")
	}

	// For now, hold a global lock so only 1 outstanding mailbox
	// message at any time.
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	// Wait for space to send
	for (reg.Read(peripheralBase+MAILBOX_STATUS_REG) & MAILBOX_FULL) != 0 {
		runtime.Gosched()
	}

	// Send
	reg.Write(peripheralBase+MAILBOX_WRITE_REG, uint32(channel&0xF)|uint32(addr&0xFFFFFFF0))

	// Wait for response
	for (reg.Read(peripheralBase+MAILBOX_STATUS_REG) & MAILBOX_EMPTY) != 0 {
		runtime.Gosched()
	}

	// Read response
	data := reg.Read(peripheralBase + MAILBOX_READ_REG)

	// Ensure response corresponds to request (note response data over-writes request data)
	if (data & 0xF) != uint32(channel&0xF) {
		panic(fmt.Sprintf("overlapping messages, got response for channel %d, expecting %d", data&0xF, channel&0xF))
	}
	if (data & 0xFFFFFFF0) != (addr & 0xFFFFFFF0) {
		panic(fmt.Sprintf("overlapping messages, got response for channel %d, expecting %d", data&0xFFFFFFF0, addr&0xFFFFFFF0))
	}
}
