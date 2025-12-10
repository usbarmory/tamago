// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

// defined in port_amd64.s
func In8(port uint16) (val uint8)
func Out8(port uint16, val uint8)
func In16(port uint16) (val uint16)
func Out16(port uint16, val uint16)
func In32(port uint16) (val uint32)
func Out32(port uint16, val uint32)
