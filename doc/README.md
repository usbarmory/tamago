# Runtime API for `GOOS=tamago`

Package doc describes required, as well as optional, runtime functions/variables for target \`GOOS=tamago\` as supported by the TamaGo framework for bare metal Go, see [tamago](<https://github.com/usbarmory/tamago>).

These hooks act as a "Rosetta Stone" for integration of a freestanding Go runtime within an arbitrary environment, whether bare metal or OS supported.

For bare metal examples see the following packages: [usbarmory](<https://github.com/usbarmory/tamago/tree/master/board/usbarmory>), [uefi](<https://github.com/usbarmory/go-boot/tree/main/uefi>), [microvm](<https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm>).

For OS supported examples see the following tamago packages: [linux](<https://github.com/usbarmory/tamago/tree/master/user/linux>), [applet](<https://github.com/usbarmory/GoTEE/tree/master/applet>).

This package is only used for documentation purposes, applications need to define the described functions/variables or import them from external packages \(such as the ones provided by [tamago](<https://github.com/usbarmory/tamago>)\) relevant to the target environment.

## Index

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago/doc.svg)](https://pkg.go.dev/github.com/usbarmory/tamago/doc)

- [Variables](<#variables>)
- [func GetRandomData\(b \[\]byte\)](<#GetRandomData>)
- [func Hwinit0\(\)](<#Hwinit0>)
- [func Hwinit1\(\)](<#Hwinit1>)
- [func InitRNG\(\)](<#InitRNG>)
- [func Nanotime\(\) int64](<#Nanotime>)
- [func Printk\(c byte\)](<#Printk>)


## Variables

<a name="Bloc"></a>Bloc describes the optional override of [runtime.Bloc](<https://pkg.go.dev/runtime/#Bloc>) to redefine the heap memory start address, this is typically only required on OS supported environments.

For an example see package [linux](<https://github.com/usbarmory/tamago/blob/master/user/linux/runtime.go>).

```go
var Bloc uintptr
```

<a name="Exit"></a>Exit describes the optional [runtime.Exit](<https://pkg.go.dev/runtime/#Exit>) function, which can be set to override default runtime termination.

For an example see package [microvm](<https://github.com/usbarmory/tamago/blob/master/board/qemu/microvm/microvm.go>).

```go
var Exit func(int32)
```

<a name="Idle"></a>Idle describes the [runtime.Idle](<https://pkg.go.dev/runtime/#Idle>) function, which can be set to implement CPU idle time management.

For a basic example see package [amd64](<https://github.com/usbarmory/tamago/blob/master/amd64/amd64.go>), a more advanced example involving a physical countdown timer such as \[arm.CPU.SetAlarm\] is implemented in the [tamago example](<https://github.com/usbarmory/tamago-example/blob/master/network/imx.go>).

```go
var Idle func(until int64)
```

<a name="RamSize"></a>RamSize, which must be linked as \[runtime.ramSize\]¹, defines the total size of the physical or virtual memory available to the runtime for allocation \(including the code segment which must be mapped within\).

For an example see package [microvm memory layout](<https://github.com/usbarmory/tamago/blob/master/board/firecracker/microvm/mem.go>).

```
¹ //go:linkname RamSize runtime.ramSize
```

```go
var RamSize uint
```

<a name="RamStackOffset"></a>RamStackOffset, which must be linked as \[runtime.ramStackOffset\]¹, defines the negative offset from the end of the available memory for stack allocation.

For an example see package [amd64](<https://github.com/usbarmory/tamago/blob/master/amd64/amd64.go>).

```
¹ //go:linkname RamStackOffset runtime.ramStackOffset
```

```go
var RamStackOffset uint
```

<a name="RamStart"></a>RamStart, which must be linked as \[runtime.ramStart\]¹, defines the start address of the physical or virtual memory available to the runtime for allocation \(including the code segment which must be mapped within\).

For an example see package [amd64 memory layout](<https://github.com/usbarmory/tamago/blob/master/amd64/mem.go>).

```
¹ //go:linkname RamStart runtime.ramStart
```

```go
var RamStart uint
```

<a name="SocketFunc"></a>SocketFunc describes the optional override of [net.SocketFunc](<https://pkg.go.dev/net/#SocketFunc>) to provide the network socket implementation. The returned interface must match the requested socket and be either [net.Conn](<https://pkg.go.dev/net/#Conn>), [net.PacketConn](<https://pkg.go.dev/net/#PacketConn>) or [net.Listen](<https://pkg.go.dev/net/#Listen>).

For an example see package [vnet](<https://github.com/usbarmory/virtio-net/blob/master/runtime.go>).

```go
var SocketFunc func(ctx context.Context, net string, family, sotype int, laddr, raddr Addr) (interface{}, error)
```

<a name="Task"></a>Task describes the optional [runtime.Task](<https://pkg.go.dev/runtime/#Task>) function, which can be set to provide an implementation for HW/OS threading \(see \[runtime.newosproc\]\).

The call takes effect only when [runtime.NumCPU](<https://pkg.go.dev/runtime/#NumCPU>) is greater than 1 \(see [runtime.SetNumCPU](<https://pkg.go.dev/runtime/#SetNumCPU>)\).

```go
var Task func(stk, mp, g0, fn unsafe.Pointer)
```

<a name="GetRandomData"></a>
## func [GetRandomData](<https://github.com/usbarmory/tamago/blob/master/doc/api_doc_stub.go#L112>)

```go
func GetRandomData(b []byte)
```

GetRandomData, which must be linked as \[runtime.GetRandomData\]¹, generates len\(b\) random bytes and writes them into b.

For an example see package [amd64 random number generation](<https://github.com/usbarmory/tamago/blob/master/amd64/rng.go>).

```
¹ //go:linkname GetRandomData runtime.getRandomData
```

<a name="Hwinit0"></a>
## func [Hwinit0](<https://github.com/usbarmory/tamago/blob/master/doc/api_doc_stub.go#L60>)

```go
func Hwinit0()
```

Hwinit0, which must be linked as \[runtime.hwinit0\]¹, takes care of the lower level initialization triggered before runtime setup \(pre World start\).

It must be defined using Go's Assembler to retain Go's commitment to backward compatibility, otherwise extreme care must be taken as the lack of World start does not allow memory allocation.

For an example see package [amd64 initialization](<https://github.com/usbarmory/tamago/blob/master/amd64/mem.go>).

```
¹ //go:linkname Hwinit0 runtime.hwinit0
```

<a name="Hwinit1"></a>
## func [Hwinit1](<https://github.com/usbarmory/tamago/blob/master/doc/api_doc_stub.go#L72>)

```go
func Hwinit1()
```

Hwinit1, which must be linked as \[runtime.hwinit1\]¹, takes care of the lower level initialization triggered early in runtime setup \(post World start\).

For an example see package [microvm platform initialization](<https://github.com/usbarmory/tamago/blob/master/board/firecracker/microvm/microvm.go>).

```
¹ //go:linkname Hwinit1 runtime.hwinit1
```

<a name="InitRNG"></a>
## func [InitRNG](<https://github.com/usbarmory/tamago/blob/master/doc/api_doc_stub.go#L100>)

```go
func InitRNG()
```

InitRNG, which must be linked as \[runtime.initRNG\]¹, initializes random number generation.

For an example see package [amd64 randon number generation](<https://github.com/usbarmory/tamago/blob/master/amd64/rng.go>).

```
¹ //go:linkname InitRNG runtime.initRNG
```

<a name="Nanotime"></a>
## func [Nanotime](<https://github.com/usbarmory/tamago/blob/master/doc/api_doc_stub.go#L128>)

```go
func Nanotime() int64
```

Nanotime, which must be linked as \[runtime.nanotime1\]¹, returns the system time in nanoseconds.

It must be defined using Go's Assembler to retain Go's commitment to backward compatibility, otherwise extreme care must be taken as the lack of World start does not allow memory allocation.

For an example see package [fu540 initialization](<https://github.com/usbarmory/tamago/blob/master/soc/sifive/fu540/init.go>).

```
¹ //go:linkname Nanotime runtime.nanotime1
```

<a name="Printk"></a>
## func [Printk](<https://github.com/usbarmory/tamago/blob/master/doc/api_doc_stub.go#L88>)

```go
func Printk(c byte)
```

Printk, which must be linked as \[runtime.printk\]¹, handles character printing to standard output.

It must be defined using Go's Assembler to retain Go's commitment to backward compatibility, otherwise extreme care must be taken as the lack of World start does not allow memory allocation.

For an example see package [usbarmory console handling](<https://github.com/usbarmory/tamago/blob/master/board/usbarmory/mk2/console.go>).

```
¹ //go:linkname Printk runtime.printk
```
