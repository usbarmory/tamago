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

func (hw *CAAM) hash(buf []byte, mode int, size int, init bool, term bool) (sum []byte, err error) {
	sourceBufferAddress := dma.Alloc(buf, 4)
	defer dma.Free(sourceBufferAddress)

	op := Operation{}
	op.SetDefaults()
	op.OpType(OPTYPE_ALG_CLASS2)
	op.Algorithm(mode, 0)

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
	src.Class(2)
	src.DataType(INPUT_DATA_TYPE_MESSAGE_DATA | INPUT_DATA_TYPE_LC2)
	src.Pointer(sourceBufferAddress, len(buf))

	jd := op.Bytes()
	jd = append(jd, src.Bytes()...)

	if term {
		sum = make([]byte, size)

		destinationBufferAddress := dma.Alloc(sum, 4)
		defer dma.Free(destinationBufferAddress)

		dst := Store{}
		dst.SetDefaults()
		dst.Class(2)
		dst.Source(CTX)
		dst.Pointer(destinationBufferAddress, len(sum))

		jd = append(jd, dst.Bytes()...)

		defer func() {
			dma.Read(destinationBufferAddress, 0, sum)
		}()
	}

	err = hw.job(nil, jd)

	return
}
