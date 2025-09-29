TamaGo - bare metal Go - Userspace support
==========================================

tamago | https://github.com/usbarmory/tamago  

Copyright (c) The TamaGo Authors. All Rights Reserved.  

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Andrej Rosano  
andrej@inversepath.com  

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal processors.

The execution of programs compiled with `GOOS=tamago` can also take place in
user space by importing any package that implements the required
[runtime changes](https://pkg.go.dev/github.com/usbarmory/tamago/doc)
with OS supervision instead of bare metal drivers.

Compiling and running Go programs in user space as `GOOS=tamago` provides the
benefit of system call isolation as the executable cannot leverage on the Go
runtime to directly access OS resources, this results in:

  * isolation from OS file system, through in-memory emulated disk
  * isolation from OS networking, see [net.SocketFunc](https://github.com/usbarmory/tamago-go/blob/latest/src/net/net_tamago.go)
  * API for custom networking, rng, time handlers

Currently supported `GOOS` are `amd64`, `arm`, `arm64`, `riscv64`.

Example
=======

The following example code

```go
package main

import (
	"fmt"
	"net"
	"os"

	_ "github.com/usbarmory/tamago/user/linux"
)

func main() {
	_, err := net.Dial("tcp", "8.8.8.8:53")
	fmt.Printf("** I can't get out!    ;-( ** %s\n", err)

	_, err = os.ReadFile("/etc/passwd")
	fmt.Printf("** I can't get out!    ;-( ** %s\n", err)
}
```

can be executed as follows:

```
GOOS=tamago GOARCH=amd64 $TAMAGO run -ldflags '-X runtime.testBinary=true' test.go
** I can't get out!    ;-( ** dial tcp 8.8.8.8:53: net.SocketFunc is nil
** I can't get out!    ;-( ** open /etc/passwd: No such file or directory
```

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For TamaGo see its [repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki) for information.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/tamago).

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
