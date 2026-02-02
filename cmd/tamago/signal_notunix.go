// tamago-go installer and runner (non-UNIX signals)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build plan9 || windows

package main

import "os"

var signalsToIgnore = []os.Signal{os.Interrupt}
