// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"bytes"
	"encoding/binary"
	"math"
	"runtime"
	"time"

	"github.com/usbarmory/tamago/dma"
)

// Interrupt Gate Descriptor Attributes
const (
	InterruptGate = 0b10001110
	TrapGate      = 0b10001111
)

// IRQ handling jump table constants
const (
	callSize = 5
	vectors  = 256
)

// IRQ handling jump table variables
var (
	idtAddr        uintptr
	irqHandlerAddr uintptr
)

// IRQ handling goroutine
var (
	irqHandlerG   uint64
	currentVector uintptr
)

// defined in irq.s
func load_idt() (idt uintptr, irqHandler uintptr)
func irq_enable()
func irq_disable()
func wait_interrupt()

//go:nosplit
func irqHandler()

var throwing bool

func currentInterrupt() (id int) {
	id = int(currentVector - irqHandlerAddr)

	if id > 0 {
		id = id / callSize
	}

	return
}

// DefaultExceptionHandler handles an exception by printing its vector and
// processor mode before panicking.
func DefaultExceptionHandler() {
	if throwing {
		halt(0)
	}

	// TODO: implement runtime.CallOnG0 for a cleaner approach
	throwing = true

	print("exception: vector ", currentInterrupt(), " \n")
	panic("unhandled exception")
}

// GateDescriptor represents an IDT Gate descriptor
// (Intel® 64 and IA-32 Architectures Software Developer’s Manual
// Volume 3A - 6.14.1 64-Bit Mode IDT).
type GateDescriptor struct {
	Offset1         uint16
	SegmentSelector uint16
	IST             uint8
	Attributes      uint8
	Offset2         uint16
	Offset3         uint32
	Reserved        uint32
}

// Bytes converts the descriptor structure to byte array format.
func (d *GateDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// SetOffset sets the address of the handling procedure entry point.
func (d *GateDescriptor) SetOffset(addr uintptr) {
	d.Offset1 = uint16(addr & 0xffff)
	d.Offset2 = uint16(addr >> 16 & 0xffff)
	d.Offset3 = uint32(addr >> 32)
}

func setIDT(start int, end int) {
	if idtAddr == 0 || irqHandlerAddr == 0 {
		idtAddr, irqHandlerAddr = load_idt()
	}

	desc := &GateDescriptor{
		SegmentSelector: 1 << 3,
		Attributes:      InterruptGate,
	}

	gateSize := len(desc.Bytes())
	idtSize := gateSize * vectors

	r, err := dma.NewRegion(uint(idtAddr), idtSize, true)

	if err != nil {
		panic(err)
	}

	addr, idt := r.Reserve(idtSize, 0)
	defer r.Release(addr)

	for i := start; i < end; i++ {
		if i == vectors {
			break
		}

		off := irqHandlerAddr + uintptr(i*callSize)
		// set ISR to irqHandler.abi0 + vector offset
		desc.SetOffset(off)
		copy(idt[i*gateSize:], desc.Bytes())
	}
}

// EnableExceptions initializes handling of processor exceptions through
// DefaultExceptionHandler().
func (cpu *CPU) EnableExceptions() {
	// processor exceptions
	setIDT(0, 31)
}

// EnableInterrupts unmasks external interrupts.
// status.
func (cpu *CPU) EnableInterrupts() {
	irq_enable()
}

// DisableInterrupts masks external interrupts.
func (cpu *CPU) DisableInterrupts() {
	irq_disable()
}

// Waitnterrupt suspends execution until an interrupt is received.
func (cpu *CPU) Waitnterrupt() {
	wait_interrupt()
}

// ServiceInterrupts puts the calling goroutine in wait state, its execution is
// resumed when a user defined interrupt is received, an argument function can
// be set for servicing.
func ServiceInterrupts(isr func(id int)) {
	irqHandlerG, _ = runtime.GetG()

	if isr == nil {
		isr = func(_ int) { return }
	}

	// user defined interrupts
	setIDT(32, 255)

	for {
		// To avoid losing interrupts, re-enabling must happen only after we
		// are sleeping.
		go irq_enable()

		// Sleep indefinitely until woken up by runtime.WakeG
		// (see irqHandler).
		time.Sleep(math.MaxInt64)

		isr(currentInterrupt())
	}
}
