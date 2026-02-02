// Package goos describes required, as well as optional, runtime
// functions/variables for custom GOOS implementations as supported by the
// GOOSPKG variable.
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
// This package is a stub and is only used for documentation purposes,
// applications need to define the described functions/variables or import them
// from external packages (such as the ones provided by [tamago]) relevant to
// the target environment.
//
// [tamago]: https://github.com/usbarmory/tamago
// [usbarmory]: https://github.com/usbarmory/tamago/tree/master/board/usbarmory
// [uefi]: https://github.com/usbarmory/go-boot/tree/main/uefi
// [microvm]: https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm
// [linux]: https://github.com/usbarmory/tamago/tree/master/user/linux
// [applet]: https://github.com/usbarmory/GoTEE/tree/master/applet
package goos

import "unsafe"

// Required variables.
var (
	// RamStart defines the start address of the physical or virtual memory
	// available to the runtime for allocation (including the code segment
	// which must be mapped within).
	RamStart uint

	// RamSize defines the total size of the physical or virtual memory
	// available to the runtime for allocation (including the code segment
	// which must be mapped within).
	RamSize uint

	// RamStackOffset, defines the negative offset from the end of the
	// available memory for stack allocation.
	RamStackOffset uint
)

// CPUInit handles immediate startup CPU initialization as it represents the
// first instruction set executed.
func CPUinit()

// Hwinit0 takes care of the lower level initialization triggered before
// runtime setup (pre World start).
//
// It must be defined using Go's Assembler to retain Go's commitment to
// backward compatibility, otherwise extreme care must be taken as the lack of
// World start does not allow memory allocation.
func Hwinit0()

// InitRNG initializes random number generation.
func InitRNG()

// GetRandomData generates len(b) random bytes and writes them into b.
func GetRandomData(b []byte)

// Nanotime returns the system time in nanoseconds.
//
// Before [Hwinit1] it must be defined using Go's Assembler to retain Go's
// commitment to backward compatibility, otherwise extreme care must be taken
// as the lack of World start does not allow memory allocation.
func Nanotime() int64

// Printk handles character printing to standard output.
//
// Before [Hwinit1] it must be defined using Go's Assembler to retain Go's
// commitment to backward compatibility, otherwise extreme care must be taken
// as the lack of World start does not allow memory allocation.
func Printk(c byte)

// Hwinit1 takes care of the lower level initialization triggered early in
// runtime setup (post World start).
func Hwinit1()

// Optional variables/functions.
var (
	// Bloc is an optional variable which can be set to redefine the heap
	// memory start address, this is typically only required on OS
	// supported environments.
	Bloc uintptr

	// Exit is an optional function which can be set to override default
	// runtime termination.
	Exit func(code int32)

	// Idle is an optional function which can be set to implement CPU idle
	// time management.
	Idle func(until int64)

	// ProcID is an optional function which can be set to provide the
	// processor identifier for tracing purposes.
	ProcID func() uint64

	// Task is an optional function which can be set to provide an
	// implementation for HW/OS threading (e.g. [runtime.newosproc]).
	Task func(sp, mp, gp, fn unsafe.Pointer)
)
