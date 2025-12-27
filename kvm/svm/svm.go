// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package svm implements a driver for AMD specific hypervisor calls, issued by
// a Secure Virtual Machine, following reference specifications:
//
//   - AMD64 Architecture Programmer’s Manual, Volume 2
//   - SEV-ES Guest-Hypervisor Communication Block Standardization
//   - SEV Secure Nested Paging Firmware ABI Specification
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package svm

// defined in svm.s
func vmgexit()
func pvalidate(addr uint64, validate bool) uint32
