// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// stub for pkg.go.dev coverage
//go:build !tamago

// Package doc describes required, as well as optional, runtime
// functions/variables for target `GOOS=tamago` as supported by the TamaGo
// framework for bare metal Go, see [tamago].
//
// These hooks act as a "Rosetta Stone" for integration of a freestanding Go
// runtime within an arbitrary environment, whether bare metal or OS supported.
//
// For bare metal examples see the following packages: [usbarmory], [uefi],
// [microvm].
//
// For OS supported examples see the following tamago packages: [linux],
// [applet].
//
// This package is only used for documentation purposes, applications need to
// define the described functions/variables or import them from external
// packages (such as the ones provided by [tamago]) relevant to the target
// environment.
//
// [tamago]: https://github.com/usbarmory/tamago
// [usbarmory]: https://github.com/usbarmory/tamago/tree/master/board/usbarmory
// [uefi]: https://github.com/usbarmory/go-boot/tree/main/uefi
// [microvm]: https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm
// [linux]: https://github.com/usbarmory/tamago/tree/master/user/linux
// [applet]: https://github.com/usbarmory/GoTEE/tree/master/applet
package doc

// cpuinit handles pre-runtime CPU initialization.
//
// It must be defined using Go's Assembler to retain Go's commitment to
// backward compatibility, otherwise extreme care must be taken as the lack of
// World start does not allow memory allocation.
//
// For an example see package [arm CPU initialization].
//
// [arm CPU initialization]: https://github.com/usbarmory/tamago/blob/master/arm/init.s
func cpuinit()

// Hwinit0, which must be linked as [runtime.hwinit0]¹, takes care of the lower
// level initialization triggered before runtime setup (pre World start).
//
// It must be defined using Go's Assembler to retain Go's commitment to
// backward compatibility, otherwise extreme care must be taken as the lack of
// World start does not allow memory allocation.
//
// For an example see package [amd64 initialization].
//
//  ¹ //go:linkname Hwinit0 runtime.hwinit0
//
// [amd64 initialization]: https://github.com/usbarmory/tamago/blob/master/amd64/mem.go
//
//go:linkname Hwinit0 runtime.hwinit0
func Hwinit0()

// Hwinit1, which must be linked as [runtime.hwinit1]¹, takes care of the lower
// level initialization triggered early in runtime setup (post World start).
//
// For an example see package [microvm platform initialization].
//
//  ¹ //go:linkname Hwinit1 runtime.hwinit1
//
// [microvm platform initialization]: https://github.com/usbarmory/tamago/blob/master/board/firecracker/microvm/microvm.go
//
//go:linkname Hwinit1 runtime.hwinit1
func Hwinit1()

// Printk, which must be linked as [runtime.printk]¹, handles character printing
// to standard output.
//
// It must be defined using Go's Assembler to retain Go's commitment to
// backward compatibility, otherwise extreme care must be taken as the lack of
// World start does not allow memory allocation.
//
// For an example see package [usbarmory console handling].
//
//  ¹ //go:linkname Printk runtime.printk
//
// [usbarmory console handling]: https://github.com/usbarmory/tamago/blob/master/board/usbarmory/mk2/console.go
//
//go:linkname Printk runtime.printk
func Printk(c byte)

// InitRNG, which must be linked as [runtime.initRNG]¹, initializes random
// number generation.
//
// For an example see package [amd64 randon number generation].
//
//  ¹ //go:linkname InitRNG runtime.initRNG
//
// [amd64 randon number generation]: https://github.com/usbarmory/tamago/blob/master/amd64/rng.go
//
//go:linkname InitRNG runtime.initRNG
func InitRNG()

// GetRandomData, which must be linked as [runtime.GetRandomData]¹, generates
// len(b) random bytes and writes them into b.
//
// For an example see package [amd64 random number generation].
//
//  ¹ //go:linkname GetRandomData runtime.getRandomData
//
// [amd64 random number generation]: https://github.com/usbarmory/tamago/blob/master/amd64/rng.go
//
//go:linkname GetRandomData runtime.getRandomData
func GetRandomData(b []byte) 

// Nanotime, which must be linked as [runtime.nanotime1]¹, returns the system
// time in nanoseconds.
//
// It must be defined using Go's Assembler to retain Go's commitment to
// backward compatibility, otherwise extreme care must be taken as the lack of
// World start does not allow memory allocation.
//
// For an example see package [fu540 initialization].
//
//  ¹ //go:linkname Nanotime runtime.nanotime1
//
// [fu540 initialization]: https://github.com/usbarmory/tamago/blob/master/soc/sifive/fu540/init.go
//
//go:linkname Nanotime runtime.nanotime1
func Nanotime() int64

// RamStart, which must be linked as [runtime.ramStart]¹, defines the start
// address of the physical or virtual memory available to the runtime for
// allocation (including the code segment which must be mapped within).
//
// For an example see package [amd64 memory layout].
//
//  ¹ //go:linkname RamStart runtime.ramStart
//
// [amd64 memory layout]: https://github.com/usbarmory/tamago/blob/master/amd64/mem.go
//
//go:linkname RamStart runtime.ramStart
var RamStart uint

// RamSize, which must be linked as [runtime.ramSize]¹, defines the total size
// of the physical or virtual memory available to the runtime for allocation
// (including the code segment which must be mapped within).
//
// For an example see package [microvm memory layout].
//
//  ¹ //go:linkname RamSize runtime.ramSize
//
// [microvm memory layout]: https://github.com/usbarmory/tamago/blob/master/board/firecracker/microvm/mem.go
//
//go:linkname RamSize runtime.ramSize
var RamSize uint

// RamStackOffset, which must be linked as [runtime.ramStackOffset]¹, defines
// the negative offset from the end of the available memory for stack
// allocation.
//
// For an example see package [amd64].
//
//  ¹ //go:linkname RamStackOffset runtime.ramStackOffset
//
// [amd64]: https://github.com/usbarmory/tamago/blob/master/amd64/amd64.go
//
//go:linkname ramStackOffset runtime.ramStackOffset
var RamStackOffset uint

// Bloc describes the optional override of [runtime.Bloc] to redefine the heap
// memory start address, this is typically only required on OS supported
// environments.
//
// For an example see package [linux].
//
// [linux]: https://github.com/usbarmory/tamago/blob/master/user/linux/runtime.go
var Bloc uintptr

// Exit describes the optional set of [runtime.Exit] to define a runtime
// termination function.
//
// For an example see package [microvm].
//
// [microvm]: https://github.com/usbarmory/tamago/blob/master/board/qemu/microvm/microvm.go
var Exit func(int32)

// Idle describes the optional set of [runtime.Idle] to define a CPU
// idle function.
//
// For a basic example see package [amd64], a more advanced example involving a
// physical countdown timer such as [arm.CPU.SetAlarm] is implemented in the [tamago example].
//
// [amd64]: https://github.com/usbarmory/tamago/blob/master/amd64/amd64.go
// [tamago example]: https://github.com/usbarmory/tamago-example/blob/master/network/imx.go
var Idle func(until int64)

// SocketFunc describes the optional override of [net.SocketFunc] to
// provide the network socket implementation. The returned interface
// must match the requested socket and be either [net.Conn],
// [net.PacketConn] or [net.Listen].
//
// For an example see package [vnet].
//
// [vnet]: https://github.com/usbarmory/virtio-net/blob/master/runtime.go
var SocketFunc func(ctx context.Context, net string, family, sotype int, laddr, raddr Addr) (interface{}, error)
