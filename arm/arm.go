// ARM processor
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

type Arm struct {
	features features
}

func (a *Arm) Init() {
	a.features.init()
}
