// BCM2835 SoC timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"time"
)

// WatchdogPeriod is the fixed 16us frequency of the BCM2835 watchdog.
const WatchdogPeriod = uint64(16 * time.Microsecond)

// SysTimerFreq is the frequency (Hz) of the BCM2835 free-running
// timer (fixed at 1Hz).
const SysTimerFreq = 1000000

// defined in timer.s
func read_systimer() int64
