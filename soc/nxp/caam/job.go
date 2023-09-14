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
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

// CAAM Job Ring registers
const (
	CAAM_JR0_MIDR_MS = 0x10
	CAAM_JR1_MIDR_MS = 0x18
	CAAM_JR2_MIDR_MS = 0x20

	JRxMIDR_MS_JROWN_NS = 3

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
	jobWords         = 1
	jobResultWords   = 2
)

var once sync.Once

type jobRing struct {
	sync.Mutex

	// base register
	base uint32

	// control registers
	irjar uint32
	orjrr uint32
	orsfr uint32

	// input/output ring size
	size int
	// job queue
	input uint32
	// results queue
	output uint32
}

func (ring *jobRing) initQueue(words int, size int) uint32 {
	buf := make([]byte, size*words*4)
	return uint32(dma.Alloc(buf, 4))
}

func (ring *jobRing) init(base uint32, size int) {
	ring.base = base
	ring.irjar = ring.base + CAAM_IRJAR_JRx
	ring.orjrr = ring.base + CAAM_ORJRR_JRx
	ring.orsfr = ring.base + CAAM_ORSFR_JRx

	if ring.size > 0 {
		dma.Free(uint(ring.input))
		dma.Free(uint(ring.output))
	}

	ring.size = size
	ring.input = ring.initQueue(jobWords, ring.size)
	ring.output = ring.initQueue(jobResultWords, ring.size)

	reg.Write(ring.base+CAAM_IRBAR_JRx, ring.input)
	reg.Write(ring.base+CAAM_IRSR_JRx, uint32(ring.size))

	reg.Write(ring.base+CAAM_ORBAR_JRx, ring.output)
	reg.Write(ring.base+CAAM_ORSFR_JRx, uint32(ring.size))
}

func (ring *jobRing) add(hdr *Header, jd []byte) (err error) {
	if hdr == nil {
		hdr = &Header{}
		hdr.SetDefaults()
		hdr.Length(1 + len(jd)/4)
	}

	jd = append(hdr.Bytes(), jd...)

	ptr := dma.Alloc(jd, 4)
	defer dma.Free(ptr)

	ring.Lock()
	defer ring.Unlock()

	// add job descriptor to input ring
	reg.Write(ring.input, uint32(ptr))

	// signal job addition
	reg.Write(ring.irjar, 1)
	defer reg.Write(ring.orjrr, 1)

	// wait for job completion
	reg.Wait(ring.orsfr, 0, 0x3ff, 1)

	if res := reg.Read(ring.output); res != uint32(ptr) {
		return fmt.Errorf("CAAM job error, invalid output descriptor")
	}

	if status := reg.Read(ring.output + 4); status != 0 {
		return fmt.Errorf("CAAM job error, status:%#x", status)
	}

	return
}

func (hw *CAAM) initJobRing() {
	// start is required to enable access to job ring registers
	startJRx := (jobRingInterface >> 12) - 1

	jrstart := hw.Base + CAAM_JRSTART
	reg.Clear(jrstart, startJRx)
	reg.Set(jrstart, startJRx)

	hw.jr = &jobRing{}
	hw.jr.init(hw.Base+jobRingInterface, jobRingSize)

	// initialize internal RNG access, required for certain CAAM commands
	hw.initRNG()
}

func (hw *CAAM) job(hdr *Header, jd []byte) (err error) {
	once.Do(hw.initJobRing)
	return hw.jr.add(hdr, jd)
}

// SetOwner defines the bus master that is permitted to access CAAM job ring
// registers. The argument defines either secure (e.g. TrustZone Secure World)
// or non-secure (e.g. TrustZone Normal World) ownership.
func (hw *CAAM) SetOwner(secure bool) {
	reg.SetTo(hw.Base+CAAM_JR0_MIDR_MS, JRxMIDR_MS_JROWN_NS, !secure)
	hw.initJobRing()
}
