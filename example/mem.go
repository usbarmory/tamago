// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package main

import (
	"fmt"
	"runtime"
)

func testAlloc(runs int, chunks int, chunkSize int) {
	var memstats runtime.MemStats

	for run := 1; run <= runs; run++ {
		fmt.Printf("allocating %d MB chunks...", chunkSize)

		mem := make([][]byte, chunks)

		for i := 0; i <= chunks-1; i++ {
			mem[i] = make([]byte, chunkSize*1024*1024)
		}

		runtime.ReadMemStats(&memstats)
		fmt.Printf("done %d/%d (%d MB) - Mallocs: %d Frees: %d HeapSys: %d\n",
			run, runs, chunks*chunkSize,
			memstats.Mallocs, memstats.Frees, memstats.HeapSys)

		// Forced GC runs hangs forever as runtime.bgscavenge is
		// affected by the FIXME we currently have in lock_tamago.go.
		// runtime.GC()
	}
}
