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

	// Instead of forcing runtime.GC() as shown in the loop, gcpercent can
	// be tuned to a value sufficiently low to prevent the next GC target
	// being set beyond the end of available RAM. A lower than default
	// (100) value (such as 80 for this example) triggers GC more
	// frequently and avoids forced GC runs.
	//
	// This is not something unique to `GOOS=tamago` but more evident as,
	// when running on bare metal, there is no swap or OS virtual memory.
	//
	//gcpercent := 80
	//fmt.Printf("setting garbage collection target: %d\n", gcpercent)
	//debug.SetGCPercent(gcpercent)

	for run := 1; run <= runs; run++ {
		fmt.Printf("allocating %d * %d MB chunks (%d/%d) ", chunks, chunkSize/(1024*1024), run, runs)

		mem := make([][]byte, chunks)

		for i := 0; i <= chunks-1; i++ {
			fmt.Printf(".")
			mem[i] = make([]byte, chunkSize)
		}

		fmt.Printf("\n")

		// When getting close to the end of available RAM, the next GC
		// target might be set beyond it. Therfore in this specific
		// test it is best to force a GC run.
		//
		// This is not something unique to `GOOS=tamago` but more
		// evident as when running bare metal we have no swap or OS
		// virtual memory.
		runtime.GC()
	}

	runtime.ReadMemStats(&memstats)
	totalAllocated := uint64(runs) * uint64(chunks) * uint64(chunkSize)
	fmt.Printf("%d MB allocated (Mallocs: %d Frees: %d HeapSys: %d NumGC:%d)\n",
		totalAllocated/(1024*1024), memstats.Mallocs, memstats.Frees, memstats.HeapSys, memstats.NumGC)
}
