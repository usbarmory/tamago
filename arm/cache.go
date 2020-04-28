// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package arm

// defined in cache.s
func CacheEnable()
func CacheDisable()
func CacheFlushData()
func CacheFlushInstruction()
