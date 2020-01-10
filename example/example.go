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
	"io/ioutil"
	"log"
	"math"
	"math/big"
	mathrand "math/rand"
	"os"
	"time"

	"github.com/inversepath/tamago/imx6"
	_ "github.com/inversepath/tamago/usbarmory/mark-two"
)

const banner = "Hello from tamago/arm!"
const verbose = true

var exit chan bool

func init() {
	log.SetFlags(0)

	// imx6 package debugging
	if verbose {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	model := imx6.Model()
	_, family, revMajor, revMinor := imx6.SiliconVersion()

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900000000); err != nil {
		fmt.Printf("WARNING: error setting ARM frequency: %v\n", err)
	}

	fmt.Printf("imx6_soc: %#s (%#x, %d.%d) @ freq:%d MHz - native:%v\n",
		model, family, revMajor, revMinor, imx6.ARMFreq()/1000000, imx6.Native)
}

func main() {
	start := time.Now()
	exit = make(chan bool)
	n := 0

	fmt.Println("-- main --------------------------------------------------------------")
	fmt.Printf("%s (epoch %d)\n", banner, start.UnixNano())

	n += 1
	go func() {
		fmt.Println("-- fs ----------------------------------------------------------------")
		TestFile()
		TestDir()
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
		fmt.Printf("slept %s (%v)\n", sleep, time.Now().Sub(start))

		exit <- true
	}()

	n += 1
	go func() {
		fmt.Println("-- rng ---------------------------------------------------------------")

		size := 32

		for i := 0; i < 10; i++ {
			rng := make([]byte, size)
			rand.Read(rng)
			fmt.Printf("%x\n", rng)
		}

		count := 1000
		start := time.Now()

		for i := 0; i < count; i++ {
			rng := make([]byte, size)
			rand.Read(rng)
		}

		fmt.Printf("retrieved %d random bytes in %s\n", size*count, time.Since(start))

		seed, _ := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
		mathrand.Seed(seed.Int64())

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

	if imx6.Native && imx6.Family == imx6.IMX6ULL {
		n += 1
		go func() {
			fmt.Println("-- i.mx6 dcp ---------------------------------------------------------")
			TestDCP()
			exit <- true
		}()
	}

	fmt.Printf("launched %d test goroutines\n", n)

	for i := 1; i <= n; i++ {
		<-exit
	}

	fmt.Printf("----------------------------------------------------------------------\n")
	fmt.Printf("completed %d goroutines (%s)\n", n, time.Since(start))

	runs := 9
	chunksMax := 50
	fillSize := 160 * 1024 * 1024
	chunks := mathrand.Intn(chunksMax) + 1
	chunkSize := fillSize / chunks

	fmt.Printf("-- memory allocation (%d runs) ----------------------------------------\n", runs)
	testAlloc(runs, chunks, chunkSize)

	if imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL) {
		fmt.Println("-- i.mx6 usb ---------------------------------------------------------")
		StartUSBEthernet()
	}

	fmt.Printf("Goodbye from tamago/arm (%s)\n", time.Since(start))
}
