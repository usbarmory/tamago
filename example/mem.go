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
		fmt.Printf("allocating %d * %d MB chunks (%d/%d) ", chunks, chunkSize / (1024*1024), run, runs)

		mem := make([][]byte, chunks)

		for i := 0; i <= chunks-1; i++ {
			fmt.Printf(".")
			mem[i] = make([]byte, chunkSize)
		}

		fmt.Printf("\n")

		// FIXME: ideally we shouldn't need to force a GC, this might
		// be a side effect of being single-threaded or some issues in
		// credit allocation, pending investigation.
		runtime.GC()
	}

	runtime.ReadMemStats(&memstats)
	fmt.Printf("Mallocs: %d Frees: %d HeapSys: %d NumGC:%d\n", memstats.Mallocs, memstats.Frees, memstats.HeapSys, memstats.NumGC)
}
