// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago

package qemu

// The Go runtime does not allow running certain functions before the bootstrap
// sequence has completed, for instance doing fmt.Printf under
// runtime.schedinit doesn't work as runtime.procPin fails due to nil pointer
// mp.p.
//
// To handle this we initialize this boolean to selectively use such functions
// in packages that are used both before and after bootstrap (e.g.
// getRandomData, triggered during schedinit but also used later).

var bootstrapped bool

func init() {
	bootstrapped = true
}
