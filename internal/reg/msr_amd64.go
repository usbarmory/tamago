// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

// defined in msr_amd64.s
func ReadMSR(addr uint32) (val uint32)
func WriteMSR(addr uint32, val uint32)
