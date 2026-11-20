// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !el1secure

package arm64

// security state of EL1 when dropping from EL3 (SCR_EL3.NS)
const scrNS = 1
