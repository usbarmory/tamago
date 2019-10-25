// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

// Basic test example for tamago/arm running on the USB armory Mk II.

package main

import (
	"crypto/rand"
	"fmt"
	"time"
	"runtime"

	"github.com/inversepath/tamago/imx6"
	_ "github.com/inversepath/tamago/usbarmory/mark-two"
)

const banner = "Hello from tamago/arm!"

var exit chan bool
var memstats runtime.MemStats

func init() {
	model := imx6.Model()
	_, family, revMajor, revMinor := imx6.SiliconVersion()

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900000000); err != nil {
		fmt.Printf("WARNING: error setting ARM frequency: %v\n", err)
	}

	fmt.Printf("imx6_soc: %#s (%#x, %d.%d) @ freq:%d MHz - native:%v\n", model, family, revMajor, revMinor, imx6.ARMFreq()/1000000, imx6.Native)
}

func dumpMemStats() {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	fmt.Println("-- MemStats dump -----------------------------------------------------")
	fmt.Printf("Alloc:         %d\n", memstats.Alloc)
	fmt.Printf("TotalAlloc:    %d\n", memstats.TotalAlloc)
	fmt.Printf("Sys:           %d\n", memstats.Sys)
	fmt.Printf("Lookups:       %d\n", memstats.Lookups)
	fmt.Printf("Mallocs:       %d\n", memstats.Mallocs)
	fmt.Printf("Frees:         %d\n", memstats.Frees)
	fmt.Printf("HeapAlloc:     %d\n", memstats.HeapAlloc)
	fmt.Printf("HeapSys:       %d\n", memstats.HeapSys)
	fmt.Printf("HeapIdle:      %d\n", memstats.HeapIdle)
	fmt.Printf("HeapInuse:     %d\n", memstats.HeapInuse)
	fmt.Printf("HeapReleased:  %d\n", memstats.HeapReleased)
	fmt.Printf("HeapObjects:   %d\n", memstats.HeapObjects)
	fmt.Printf("StackInuse:    %d\n", memstats.StackInuse)
	fmt.Printf("StackSys:      %d\n", memstats.StackSys)
	fmt.Printf("MSpanInuse:    %d\n", memstats.MSpanInuse)
	fmt.Printf("MSpanSys:      %d\n", memstats.MSpanSys)
	fmt.Printf("MCacheInuse:   %d\n", memstats.MCacheInuse)
	fmt.Printf("MCacheSys:     %d\n", memstats.MCacheSys)
	fmt.Printf("BuckHashSys:   %d\n", memstats.BuckHashSys)
	fmt.Printf("GCSys:         %d\n", memstats.GCSys)
	fmt.Printf("OtherSys:      %d\n", memstats.OtherSys)
	fmt.Printf("NextGC:        %d\n", memstats.NextGC)
	fmt.Printf("PauseTotalNs:  %d\n", memstats.PauseTotalNs)
	fmt.Printf("NumGC:         %d\n", memstats.NumGC)
	fmt.Printf("NumForcedGC:   %d\n", memstats.NumForcedGC)
	fmt.Printf("GCCPUFraction: %f\n", memstats.GCCPUFraction)
	fmt.Printf("EnableGC:      %t\n", memstats.EnableGC)
	fmt.Printf("DebugGC:       %t\n", memstats.DebugGC)
}

func main() {
	start := time.Now()
	exit = make(chan bool)
	n := 0

	fmt.Printf("%s (epoch %d)\n", banner, start.UnixNano())

	n += 1
	go func() {
		fmt.Println("-- file --------------------------------------------------------------")
		TestFile()
		exit <- true
	}()

	n += 1
	go func() {
		sleep := 100 * time.Millisecond
		fmt.Println("-- sleep -------------------------------------------------------------")
		fmt.Printf("sleeping %s @ %d\n", sleep, time.Now().Nanosecond())
		time.Sleep(sleep)
		fmt.Printf("   slept %s @ %d\n", sleep, time.Now().Nanosecond())
		exit <- true
	}()

	n += 1
	go func() {
		fmt.Println("-- rng ---------------------------------------------------------------")
		for i := 0; i < 10; i++ {
			rng := make([]byte, 32)
			rand.Read(rng)
			fmt.Printf("   %x\n", rng)
		}
		exit <- true
	}()

	n += 1
	go func() {
		fmt.Println("-- ecdsa -------------------------------------------------------------")
		TestSignAndVerify()
		exit <- true
	}()

	n += 1
	go func() {
		fmt.Println("-- btc ---------------------------------------------------------------")
		ExamplePayToAddrScript()
		ExampleExtractPkScriptAddrs()
		ExampleSignTxOutput()
		exit <- true
	}()

	if imx6.Family == imx6.IMX6ULL && imx6.Native {
		n += 1
		go func() {
			fmt.Println("-- i.mx6 dcp ---------------------------------------------------------")
			TestDCP()
			exit <- true
		}()

		// TODO
		//n += 2 // account for eth_rx goroutine
		//go func() {
		//	fmt.Println("-- u-boot net --------------------------------------------------------")
		//	TestNet()
		//	exit <- true
		//}()
	}

	n += 1
	go func() {
		fmt.Println("-- memory allocation -------------------------------------------------")

		dumpMemStats()

		n := 64
		m := 3
		mem := make([][]byte, n)

		fmt.Printf("allocating %d MB chunks\n", m)

		for i := 0; i <= n-1; i++ {
			fmt.Printf("*")
			mem[i] = make([]byte, m*1024*1024)
		}
		fmt.Printf("\ndone (%d MB)\n", n*m)

		dumpMemStats()

		exit <- true
	}()

	fmt.Printf("launched %d test goroutines\n", n)

	for i := 1; i <= n; i++ {
		<-exit
	}

	fmt.Printf("----------------------------------------------------------------------\n")
	fmt.Printf("completed %d goroutines\n", n)
	fmt.Printf("Goodbye from tamago/arm (%s)\n", time.Since(start))
}
