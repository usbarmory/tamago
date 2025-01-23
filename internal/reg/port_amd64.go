// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

// defined in port_amd64.s
func In8(port uint16) (val uint8)
func Out8(port uint16, val uint8)
func In32(port uint32) (val uint32)
func Out32(port uint32, val uint32)
