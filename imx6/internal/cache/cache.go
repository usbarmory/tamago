// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package cache

import (
	"unsafe"
)

// defined in cache.s
func Clean(addr unsafe.Pointer)
func Enable()
func Disable()
func FlushData()
func FlushInstruction()
