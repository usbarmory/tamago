// NVM Express driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package nvme

import (
	"fmt"
	"time"

	"github.com/usbarmory/tamago/amd64"
)

// LAPIC vector range usable for device interrupts on amd64
const (
	MinVector = 32
	MaxVector = amd64.IRQ_MAX
)

// recheck is the interval at which a waiting queue re-checks its completion
// queue in the absence of interrupts, it only applies while commands are
// outstanding and protects against lost interrupts.
var recheck = 100 * time.Millisecond

// ServiceInterrupt services an interrupt, it must be invoked by the
// application interrupt service routine (see [amd64.CPU.ServiceInterrupts])
// for every vector in the [NVMe.Vector] range. It returns false when the
// vector does not belong to this controller.
func (hw *NVMe) ServiceInterrupt(vector int) (handled bool) {
	irqs := hw.irqs.Load()

	if irqs == nil {
		return false
	}

	n := vector - hw.Vector

	if n < 0 || n >= len(*irqs) {
		return false
	}

	select {
	case (*irqs)[n] <- struct{}{}:
	default:
		// a wake up is already pending
	}

	return true
}

// Vectors returns the number of interrupt vectors in use, starting from
// [NVMe.Vector].
func (hw *NVMe) Vectors() int {
	if irqs := hw.irqs.Load(); irqs != nil {
		return len(*irqs)
	}

	return 0
}

func (hw *NVMe) initMSIX() (err error) {
	if hw.Vector < MinVector || hw.Vector > MaxVector {
		return fmt.Errorf("invalid interrupt vector %d, must be within %d-%d", hw.Vector, MinVector, MaxVector)
	}

	// one vector for the admin queue plus one for each I/O queue
	entries := min(1+hw.nq, hw.msix.TableSize(), MaxVector-hw.Vector+1)

	if entries < 2 {
		return fmt.Errorf("insufficient MSI-X vectors (vector %d)", hw.Vector)
	}

	hw.nq = entries - 1

	irqs := make([]chan struct{}, entries)

	for i := range irqs {
		irqs[i] = make(chan struct{}, 1)
	}

	hw.irqs.Store(&irqs)

	for i := range entries {
		if err = hw.msix.EnableInterrupt(i, uint64(amd64.LAPIC_BASE), uint32(hw.Vector+i)); err != nil {
			return
		}
	}

	return
}
