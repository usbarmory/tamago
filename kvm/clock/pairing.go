// KVM clock pairing driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package kvmclock implements a driver for the KVM specific paravirtualized
// clocksources following the KVM_HC_CLOCK_PAIRING hypercall as described at:
//
//	https://docs.kernel.org/virt/kvm/x86/hypercalls.html
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package kvmclock

import (
	"time"
)

// Pairing() returns the KVM host clock information.
func Pairing() (sec int64, nsec int64, tsc uint64)

// Now() returns the time corresponding to the KVM host clock.
func Now() (t time.Time) {
	sec, nsec, _ := Pairing()
	return time.Unix(sec, nsec)
}
