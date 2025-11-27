// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exception

import "runtime"

func Throw(pc uintptr) {
	fn := runtime.FuncForPC(pc)
	file, line := fn.FileLine(pc)

	print("\t", file, ":", line, "\n")
	panic("unhandled exception")
}
