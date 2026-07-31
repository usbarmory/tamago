// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

// defined in reg_*.s
func Read8(addr uint32) uint8
func Write8(addr uint32, val uint8)
