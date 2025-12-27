// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

// defined in msr_amd64.s
func ReadMSR(addr uint64) (val uint64)
func WriteMSR(addr uint64, val uint64)
