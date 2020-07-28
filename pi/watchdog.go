// Raspberry Pi Watchdog Timer Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi

import (
	"time"

	"github.com/f-secure-foundry/tamago/bcm2835"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	// Power Management, Reset controller and Watchdog registers
	PM_BASE                  uint32 = (0x100000)
	PM_RSTC                  uint32 = (PM_BASE + 0x1c)
	PM_WDOG                  uint32 = (PM_BASE + 0x24)
	PM_WDOG_RESET            uint32 = 0000000000
	PM_PASSWORD              uint32 = 0x5a000000
	PM_WDOG_TIME_SET         uint32 = 0x000fffff
	PM_RSTC_WRCFG_CLR        uint32 = 0xffffffcf
	PM_RSTC_WRCFG_SET        uint32 = 0x00000030
	PM_RSTC_WRCFG_FULL_RESET uint32 = 0x00000020
	PM_RSTC_RESET            uint32 = 0x00000102
)

// watchdogPeriod is the fixed 16us frequency of BCM2835 watchdog
const watchdogPeriod = uint64(16 * time.Microsecond)

type watchdog struct {
	timeout uint32
}

// Watchdog can automatically reset the board due on lock-up.
//
// A typical example might be to reset the board on an OOM
// (Out-Of-Memory) condition.  In Go out-of-memory is not
// recoverable, and halts the CPU - automatic reset of the
// board can be an appropriate action to take.
//
// To use, start the watchdog with a timeout.  Periodically
// call Reset from your logic (within the timeout).  If you
// fail to call Reset within the timeout, the watchdog
// interrupt will fire, resetting the board.
var Watchdog = watchdog{}

// Start the watchdog timer, with a given timeout.
func (w *watchdog) Start(timeout time.Duration) {
	t := uint64(timeout) / watchdogPeriod

	// Exceeding the watchdog timeout is indicative of a major logic issue, so
	// panic rather than returning error.
	if (t & ^uint64(PM_WDOG_TIME_SET)) != 0 {
		panic("excess timeout for watchdog")
	}

	w.timeout = uint32(t)

	w.Reset()
}

// Reset the count-down on the watchdog
func (w *watchdog) Reset() {
	// Reset could probably be done more efficiently, there should be a way to reset
	// without a full re-initialization
	pm_rstc := reg.Read(bcm2835.PeripheralBase + PM_RSTC)
	pm_wdog := PM_PASSWORD | (w.timeout & PM_WDOG_TIME_SET)
	pm_rstc = PM_PASSWORD | (pm_rstc & PM_RSTC_WRCFG_CLR) | PM_RSTC_WRCFG_FULL_RESET
	reg.Write(bcm2835.PeripheralBase+PM_WDOG, pm_wdog)
	reg.Write(bcm2835.PeripheralBase+PM_RSTC, pm_rstc)
}

// Stop the watchdog
func (w *watchdog) Stop() {
	reg.Write(bcm2835.PeripheralBase+PM_RSTC, PM_PASSWORD|PM_RSTC_RESET)
}

// Remaining gets the remaining duration on the watchdog
func (w *watchdog) Remaining() time.Duration {
	t := reg.Read(bcm2835.PeripheralBase+PM_WDOG) & PM_WDOG_TIME_SET
	return time.Duration(uint64(t) * watchdogPeriod)
}
