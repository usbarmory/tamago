// BCM2835 SoC timer support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

// SysTimerFreq is the frequency (Hz) of the BCM2835 free-running
// timer (fixed at 1Hz)
const SysTimerFreq = 1000000

// defined in timer.s
func read_systimer() int64
