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
	"github.com/usbarmory/tamago/dma"
)

func (hw *CAAM) hash(buf []byte, mode int, init bool, term bool) (sum []byte, err error) {
	sourceBufferAddress := dma.Alloc(buf, len(buf))
	defer dma.Free(sourceBufferAddress)

	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_ALG_CLASS2)
	op.Algorithm(mode)

	switch {
	case init && term:
		op.State(AS_INITIALIZE | AS_FINALIZE)
	case init:
		op.State(AS_INITIALIZE)
	case term:
		op.State(AS_FINALIZE)
	default:
		op.State(AS_UPDATE)
	}

	src := FIFOLoad{}
	src.SetDefaults()
	src.Class(CLASS_2CHA)
	src.DataType(INPUT_DATA_TYPE_MESSAGE_DATA | INPUT_DATA_TYPE_LC2)
	src.Pointer(sourceBufferAddress, len(buf))

	jd := op.Bytes()
	jd = append(jd, src.Bytes()...)

	if term {
		// output is always 32 bytes, regardless of mode
		sum = make([]byte, 32)

		destinationBufferAddress := dma.Alloc(sum, len(sum))
		defer dma.Free(destinationBufferAddress)

		// output sequence start address
		dst := Store{}
		dst.SetDefaults()
		dst.Class(CLASS_2CCB)
		dst.Source(SRC_CTX)
		dst.Pointer(destinationBufferAddress, len(sum))

		jd = append(jd, dst.Bytes()...)

		defer func() {
			dma.Read(destinationBufferAddress, 0, sum)
		}()
	}

	err = hw.job(nil, jd)

	return
}
