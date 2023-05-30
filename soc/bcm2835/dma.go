// BCM2835 SoC DMA support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

//
// There are 16 DMA channels with identical register layout.
//
// The 16th channel (channel 15) is in a non-contiguous register
// space.
//
// See BCM2835-ARM-Periperals.pdf for detailed register usage
//

// Channel base addresses
const (
	// Base address of the first channel.
	DMA_CH_BASE0 = 0x7000

	// Base address of the 16th channel (ch 15).
	//
	// Channel 15 is register space is non-contiguous with other
	// channels.
	DMA_CH_BASE15 = 0x5000

	// Offset of each channel's registers from previous.
	DMA_CH_SPAN = 0x100
)

// Layout of registers for each channel
const (
	// Control and Status
	DMA_CH_REG_CS = 0x00

	// Control Block Address
	DMA_CH_REG_CONBLK_AD = 0x04

	// Debug
	DMA_CH_REG_DEBUG = 0x20
)

// DMA Control and Status flags
//
// These are a mix of read-only, write-only, read/write
const (
	// Activate (start) the DMA transfer
	DMA_CS_ACTIVE = 1 << 0

	// Indicates the DMA transfer is complete
	DMA_CS_END = 1 << 1

	// Interrupt Status
	DMA_CS_INT = 1 << 2

	// DREQ Status
	DMA_CS_DREQ = 1 << 3

	// Indicates the DMA transfer has been paused
	DMA_CS_PAUSED = 1 << 4

	// Indicates DMA is paused due to DREQ
	DMA_CS_DREQ_STOPS_DMA = 1 << 5

	// Indicates if the DMA is waiting for outstanding writes
	DMA_CS_WAITING_FOR_OUTSTANDING_WRITES = 1 << 6

	// Indicates if a DMA error occured
	DMA_CS_ERROR = 1 << 8

	// Sets Priority on the AXI bus
	DMA_CS_PRIORITY_SHIFT = 16
	DMA_CS_PRIORITY_MASK  = 0xF

	// Sets the Panic priority on the AXI bus
	DMA_CS_PANIC_PRIORITY_SHIFT = 20
	DMA_CS_PANIC_PRIORITY_MASK  = 0xF

	// Limits outstanding writes and waits for completion at end of transfer
	DMA_CS_WAIT_FOR_OUTSTANDING_WRITES = 1 << 28

	// Disables debug pause
	DMA_CS_DISDEBUG = 1 << 29

	// Aborts the current control block (transfer segment)
	DMA_CS_ABORT = 1 << 30

	// Resets the DMA channel
	DMA_CS_RESET = 1 << 31
)

// DMA Transfer Information control flags
const (
	// Interrupt Enable (signal on transfer complete)
	DMA_TI_INTEN = 0x1 << 0

	// Use two-dimensional (striped) mode
	DMA_TI_TDMODE = 0x1 << 1

	// Wait for write response
	DMA_TI_WAITRESP = 0x1 << 3

	// Increment destination address
	DMA_TI_DEST_INC = 0x1 << 4

	// Destination transfer width (1 = 128bit, 0 = 32bit)
	DMA_TI_DEST_WIDTH = 0x1 << 5

	// Control writes with DREQ (allow peripheral to control speed)
	DMA_TI_DEST_DREQ = 0x1 << 6

	// Ignore writes (do not perform writes)
	DMA_TI_DEST_IGNORE = 0x1 << 7

	// Increment source address
	DMA_TI_SRC_INC = 0x1 << 8

	// Source transfer width (1 = 128bit, 0 = 32bit)
	DMA_TI_SRC_WIDTH = 0x1 << 9

	// Control reads with DREQ (allow peripheral to control speed)
	DMA_TI_SRC_DREQ = 0x1 << 10

	// Ignore reads (do not perform reads)
	DMA_TI_SRC_IGNORE = 0x1 << 11

	// 4-bit burst length requested
	DMA_TI_BURST_LENGTH_SHIFT = 12
	DMA_TI_BURST_LENGTH_MASK  = 0xF

	// Peripheral ID for DREQ rate control
	DMA_TI_BURST_PERMAP_SHIFT = 16
	DMA_TI_BURST_PREMAP_MASK  = 0x1F

	// Add wait cycles between each DMA read or write
	DMA_TI_WAITS_SHIFT = 21
	DMA_TI_WAITS_MASK  = 0x1F

	// Don't do wide bursts
	DMA_TI_NO_WIDE_BURSTS = 0x1 << 26
)

// DMA Debug flags
const (
	// Indicates AXI read last signal was not set when expected
	DMA_DEBUG_READ_LAST_NOT_SET_ERROR = 0x1 << 0

	// Indicates FIFO error
	DMA_DEBUG_FIFO_ERROR = 0x1 << 1

	// Indicates read error
	DMA_DEBUG_READ_ERROR = 0x1 << 2

	// Indicates outstanding writes
	DMA_DEBUG_OUTSTANDING_WRITES_SHIFT = 4
	DMA_DEBUG_OUTSTANDING_WRITE_MASK   = 0xF

	// Gets the AXI ID of this channel
	DMA_DEBUG_DMA_ID_SHIFT = 8
	DMA_DEBUG_DMA_ID_MASK  = 0xFF

	// Gets the DMA engine state
	DMA_DEBUG_DMA_STATE_SHIFT = 16
	DMA_DEBUG_DMA_STATE_MASK  = 0xFF

	// Gets the DMA version
	DMA_DEBUG_VERSION_SHIFT = 25
	DMA_DEBUG_VERSION_MASK  = 0x7

	// Indicates if this is a reduced performance channel
	DMA_DEBUG_LITE = 1 << 28
)

type DMAController struct {
	sync.Mutex
	channels []*DMAChannel
	region   *dma.Region
}

// DMAChannel controls a specific DMA channel on the DMA controller
type DMAChannel struct {
	index     int
	allocated bool

	// base is the absolute base address of this channel's registers,
	// already offset by the peripheral base address.
	base uint32

	ctrlr *DMAController
}

// Access to the debug flag bitfield
type DMADebugInfo uint32

// Access to the status flag bitfield
type DMAStatus uint32

// DMA provides access to the DMA controller
var DMA = DMAController{}

// Initialize the controller.
//
// The given region provides a memory region dedicated for
// DMA transfer purposes
func (c *DMAController) Init(rgn *dma.Region) {
	c.region = rgn
	c.channels = make([]*DMAChannel, 16)

	// VideoCore reserves some DMA channels, only mark available
	// channels the VideoCore is making available to the CPU
	available := CPUAvailableDMAChannels()

	for i := range c.channels {
		c.channels[i] = &DMAChannel{
			index:     i,
			allocated: (available & (1 << i)) != 0,
			base:      peripheralBase + DMA_CH_BASE0 + uint32(i)*DMA_CH_SPAN,
			ctrlr:     c,
		}
	}

	// Channel 15 is a special case
	c.channels[15].base = peripheralBase + DMA_CH_BASE15
}

// AllocChannel provides exclusive use of a DMA channel
func (c *DMAController) AllocChannel() (*DMAChannel, error) {
	c.Lock()
	defer c.Unlock()

	for _, ch := range c.channels {
		if !ch.allocated {
			ch.allocated = true
			return ch, nil
		}
	}

	return nil, fmt.Errorf("no DMA channels available")
}

// FreeChannel surrenders exclusive use of a DMA channel
func (c *DMAController) FreeChannel(ch *DMAChannel) error {
	c.Lock()
	defer c.Unlock()

	if !ch.allocated {
		return fmt.Errorf("attempt to free unallocated channel %d", ch.index)
	}

	ch.allocated = false
	return nil
}

// Debug state for the channel
func (ch *DMAChannel) DebugInfo() DMADebugInfo {
	return DMADebugInfo(reg.Read(ch.base + DMA_CH_REG_DEBUG))
}

// Status of the channel
func (ch *DMAChannel) Status() DMAStatus {
	return DMAStatus(reg.Read(ch.base + DMA_CH_REG_CS))
}

// Do a RAM to RAM transfer.
//
// This method blocks until the transfer is complete, but does allow the
// Go scheduler to schedule other Go routines.
func (ch *DMAChannel) Copy(from uint32, size int, to uint32) {
	cbAddr, cb := ch.ctrlr.region.Reserve(8*4, 64)
	defer ch.ctrlr.region.Release(cbAddr)

	conv := binary.LittleEndian

	conv.PutUint32(cb[0:], DMA_TI_SRC_INC|DMA_TI_DEST_INC)
	conv.PutUint32(cb[4:], from)
	conv.PutUint32(cb[8:], to)
	conv.PutUint32(cb[12:], uint32(size))
	conv.PutUint32(cb[16:], 0)
	conv.PutUint32(cb[20:], 0)
	conv.PutUint32(cb[24:], 0)
	conv.PutUint32(cb[28:], 0)

	reg.Write(ch.base+DMA_CH_REG_CS, DMA_CS_RESET)
	reg.Write(ch.base+DMA_CH_REG_DEBUG, 0x7) // Clear Errors
	reg.Write(ch.base+DMA_CH_REG_CONBLK_AD, uint32(cbAddr))
	reg.Write(ch.base+DMA_CH_REG_CS, DMA_CS_ACTIVE)

	for (reg.Read(ch.base+DMA_CH_REG_CS) & DMA_CS_END) == 0 {
		runtime.Gosched()
	}
}

// Indicates the last AXI read signal was not set when expected
func (i DMADebugInfo) ReadLastNotSetError() bool {
	return (i & DMA_DEBUG_READ_LAST_NOT_SET_ERROR) != 0
}

// Indicates a FIFO error condition
func (i DMADebugInfo) FIFOError() bool {
	return (i & DMA_DEBUG_FIFO_ERROR) != 0
}

// Indicates a read error
func (i DMADebugInfo) ReadError() bool {
	return (i & DMA_DEBUG_READ_ERROR) != 0
}

// Currently outstanding writes
func (i DMADebugInfo) OutstandingWrites() int {
	return int(i>>DMA_DEBUG_OUTSTANDING_WRITES_SHIFT) & DMA_DEBUG_OUTSTANDING_WRITE_MASK
}

// AXI ID of this channel
func (i DMADebugInfo) ID() int {
	return int(i>>DMA_DEBUG_DMA_ID_SHIFT) & DMA_DEBUG_DMA_ID_MASK
}

// State Machine State
func (i DMADebugInfo) State() int {
	return int(i>>DMA_DEBUG_DMA_STATE_SHIFT) & DMA_DEBUG_DMA_STATE_MASK
}

// Version
func (i DMADebugInfo) Version() int {
	return int(i>>DMA_DEBUG_VERSION_SHIFT) & DMA_DEBUG_VERSION_MASK
}

// Indicates if reduced performance 'lite' engine
func (i DMADebugInfo) Lite() bool {
	return (i & DMA_DEBUG_LITE) != 0
}

// Compat representation of the DMA debug status for debugging
func (i DMADebugInfo) String() string {
	return fmt.Sprintf(
		"[E:%s%s%s, O:%d, ID:%d, S:%x, V:%d, L:%s]",
		boolToStr(i.ReadLastNotSetError(), "L", "l"),
		boolToStr(i.FIFOError(), "F", "f"),
		boolToStr(i.ReadError(), "R", "r"),
		i.OutstandingWrites(),
		i.ID(),
		i.State(),
		i.Version(),
		boolToStr(i.Lite(), "T", "F"),
	)
}

// Indicates if a transfer is active
func (s DMAStatus) Active() bool {
	return (uint32(s) & DMA_CS_ACTIVE) != 0
}

// Indicates if the transfer is complete
func (s DMAStatus) End() bool {
	return (uint32(s) & DMA_CS_END) != 0
}

// Gets the interrupt status
func (s DMAStatus) Int() bool {
	return (uint32(s) & DMA_CS_INT) != 0
}

// Gets the DREQ state
func (s DMAStatus) DReq() bool {
	return (uint32(s) & DMA_CS_DREQ) != 0
}

// Indicates if the transfer is currently paused
func (s DMAStatus) Paused() bool {
	return (uint32(s) & DMA_CS_PAUSED) != 0
}

// Indicates if the transfer is currently paused due to DREQ state
func (s DMAStatus) DReqStopsDMA() bool {
	return (uint32(s) & DMA_CS_DREQ_STOPS_DMA) != 0
}

// Indicates if the transfer is waiting for last write to complete
func (s DMAStatus) WaitingForOutstandingWrites() bool {
	return (uint32(s) & DMA_CS_WAITING_FOR_OUTSTANDING_WRITES) != 0
}

// Indicates if there is an error state on the channel
func (s DMAStatus) Error() bool {
	return (uint32(s) & DMA_CS_ERROR) != 0
}

// Gets the AXI priority level of the channel
func (s DMAStatus) Priority() int {
	return int((uint32(s) & DMA_CS_PRIORITY_MASK) >> DMA_CS_PRIORITY_SHIFT)
}

// Gets the AXI panic priority level of the channel
func (s DMAStatus) PanicPriority() int {
	return int((uint32(s) & DMA_CS_PANIC_PRIORITY_MASK) >> DMA_CS_PANIC_PRIORITY_SHIFT)
}

// Indicates if the channel will wait for all writes to complete
func (s DMAStatus) WaitForOutstandingWrites() bool {
	return (uint32(s) & DMA_CS_WAIT_FOR_OUTSTANDING_WRITES) != 0
}

// Indicates if the debug pause signal is disabled
func (s DMAStatus) DisableDebug() bool {
	return (uint32(s) & DMA_CS_DISDEBUG) != 0
}

// Compat representation of the DMA channel status for debugging
func (s DMAStatus) String() string {
	return fmt.Sprintf(
		"[F:%s%s%s%s%s%s%s%s%s%s, P:%d, PP:%d]",
		boolToStr(s.Active(), "A", "a"),
		boolToStr(s.End(), "E", "e"),
		boolToStr(s.Int(), "I", "i"),
		boolToStr(s.DReq(), "D", "d"),
		boolToStr(s.Paused(), "P", "p"),
		boolToStr(s.DReqStopsDMA(), "S", "s"),
		boolToStr(s.WaitingForOutstandingWrites(), "W", "w"),
		boolToStr(s.Error(), "E", "E"),
		boolToStr(s.WaitForOutstandingWrites(), "W", "w"),
		boolToStr(s.DisableDebug(), "D", "d"),
		s.Priority(),
		s.PanicPriority(),
	)
}

func boolToStr(val bool, ifTrue string, ifFalse string) string {
	if val {
		return ifTrue
	}

	return ifFalse
}
