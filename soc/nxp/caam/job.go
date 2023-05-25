// NXP Cryptographic Acceleration and Assurance Module (CAAM) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package caam

import (
	"fmt"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// CAAM Job Ring registers
const (
	CAAM_JRSTART = 0x5c

	CAAM_JR0_BASE = 0x1000
	CAAM_JR1_BASE = 0x2000
	CAAM_JR2_BASE = 0x3000

	CAAM_IRBAR_JRx = 0x04
	CAAM_IRSR_JRx  = 0x0c
	CAAM_IRJAR_JRx = 0x1c
	CAAM_ORBAR_JRx = 0x24
	CAAM_ORSR_JRx  = 0x2c
	CAAM_ORJRR_JRx = 0x34
	CAAM_ORSFR_JRx = 0x3c
)

const (
	jobRingInterface = CAAM_JR0_BASE
	jobRingSize      = 1

	inputRingWords  = 1
	outputRingWords = 3
)

type jobRing struct {
	buf  []byte
	addr uint32
}

func (ring *jobRing) init(words int, size int) uint32 {
	addr := dma.Alloc(make([]byte, size*words*4), 4)
	ring.addr = uint32(addr)
	return ring.addr
}

func (hw *CAAM) initJobRing(off int, size uint32) {
	hw.jrstart = hw.Base + CAAM_JRSTART
	hw.jr = hw.Base + uint32(off)

	// start is required before accessing the following registers
	n := (off >> 12) - 1
	reg.Clear(hw.jrstart, n)
	reg.Set(hw.jrstart, n)

	// input ring
	reg.Write(hw.jr+CAAM_IRBAR_JRx, hw.input.init(inputRingWords, jobRingSize))
	reg.Write(hw.jr+CAAM_IRSR_JRx, size)

	// output ring
	reg.Write(hw.jr+CAAM_ORBAR_JRx, hw.output.init(outputRingWords, jobRingSize))
	reg.Write(hw.jr+CAAM_ORSR_JRx, size)
}

func (hw *CAAM) job(hdr *Header, jd []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.jr == 0 {
		hw.initJobRing(jobRingInterface, jobRingSize)
	}

	if hdr == nil {
		hdr = &Header{}
		hdr.SetDefaults()
		hdr.Length(1 + len(jd)/4)
	}

	jd = append(hdr.Bytes(), jd...)

	ptr := dma.Alloc(jd, 4)
	defer dma.Free(ptr)

	// add job descriptor to input ring
	reg.Write(hw.input.addr, uint32(ptr))

	// signal job addition
	reg.Write(hw.jr+CAAM_IRJAR_JRx, 1)
	// remove job upon completion
	defer reg.Write(hw.jr+CAAM_ORJRR_JRx, 1)

	// wait for job completion
	reg.Wait(hw.jr+CAAM_ORSFR_JRx, 0, 0x3ff, 1)

	if reg.Read(hw.output.addr) != uint32(ptr) {
		return fmt.Errorf("CAAM job error, invalid output descriptor (%#x != %#x)", ptr, hw.output.addr)
	}

	if status := reg.Read(hw.output.addr + 4); status != 0 {
		return fmt.Errorf("CAAM job error, status:%#x", status)
	}

	return
}
