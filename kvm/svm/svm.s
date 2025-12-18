// AMD virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func vmgexit()
TEXT ·vmgexit(SB),$0
	BYTE $0x0f
	BYTE $0x01
	BYTE $0xd9
	RET
