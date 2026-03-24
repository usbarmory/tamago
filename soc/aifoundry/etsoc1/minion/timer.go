// AI Foundry ET-SoC-1 Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package minion

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

// Oscillator frequencies
const (
	CoreFreq        = 200e6 // 200 MHz
	TimerMultiplier = 50000
)

// Machine Timer
const ESR_MTIME = 0x01_ff80_0000

// Counter returns the CPU Machine Timer Register.
func Counter() uint64 {
	return reg.Read64(ESR_MTIME)
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return RV64.GetTime()
}
