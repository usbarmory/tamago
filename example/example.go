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

	"github.com/inversepath/tamago/imx6"
	_ "github.com/inversepath/tamago/usbarmory/mark-two"
)

const banner = "Hello from tamago/arm!"

var exit chan bool

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

	sleep := 100 * time.Millisecond

	n += 1
	go func() {
		fmt.Println("-- timer -------------------------------------------------------------")

		t := time.NewTimer(sleep)
		fmt.Printf("waking up timer after %v\n", sleep)

		start := time.Now()

		for now := range t.C {
			fmt.Printf("woke up at %d (%v)\n", now.Nanosecond(), now.Sub(start))
			break
		}

		exit <- true
	}()

	n += 1
	go func() {
		fmt.Println("-- sleep -------------------------------------------------------------")


		fmt.Printf("sleeping %s\n", sleep)
		start := time.Now()
		time.Sleep(sleep)
		fmt.Printf("   slept %s (%v)\n", sleep, time.Now().Sub(start))

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

	if imx6.Native {
		if imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL {
			n += 1
			go func() {
				fmt.Println("-- i.mx6 usb ---------------------------------------------------------")
				TestUSB()
				exit <- true
			}()
		}

		if imx6.Family == imx6.IMX6ULL {
			n += 1
			go func() {
				fmt.Println("-- i.mx6 dcp ---------------------------------------------------------")
				TestDCP()
				exit <- true
			}()
		}
	}

	fmt.Printf("launched %d test goroutines\n", n)

	for i := 1; i <= n; i++ {
		<-exit
	}

	fmt.Printf("----------------------------------------------------------------------\n")
	fmt.Printf("completed %d goroutines\n", n)

	runs := 8
	chunks := 40
	chunkSize := 4

	fmt.Printf("-- memory allocation (%d runs) ---------------------------------------\n", runs)
	testAlloc(runs, chunks, chunkSize)
	fmt.Printf("total %d MB allocated\n", runs * chunks * chunkSize)

	fmt.Printf("Goodbye from tamago/arm (%s)\n", time.Since(start))
}
