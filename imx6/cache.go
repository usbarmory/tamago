// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"unsafe"
)

// defined in cache.s
func cache_clean(addr unsafe.Pointer)
func cache_enable()
func cache_disable()
func v7_flush_dcache_all()
func v7_flush_icache_all()
