TamaGo example application
==========================

This example Go application is part of the
[TamaGo](https://github.com/f-secure-foundry/tamago) project and targets the NXP
i.MX6ULZ SoC of the USB armory Mk II.

The example application performs a variety of simple test procedures each in
its separate goroutine:

  1. Directory and file writt/read from an in-memory filesystem.

  2. Timer operation.

  3. Sleep operation.

  4. Random bytes collection (gathered from SoC TRNG on non-emulated runs).

  5. ECDSA signing and verification.

  6. Test BTC transaction creation and signing.

  7. Key derivation with DCP HSM (only on non-emulated runs).

  8. Large memory allocation.

Once all tests are completed, and only on non-emulated hardware, the following
network services are started on Ethernet over USB (ECM protocol, only supported
on Linux hosts).

  * UDP echo server on 10.0.0.1:1234
  * HTTP server on 10.0.0.1:80
  * HTTPS server on 10.0.0.1:443

The HTTP/HTTPS servers expose the following routes:

  * `/`: a welcome message
  * `/dir`: in-memory filesystem
  * `/debug/pprof`: Go runtime profiling data through [pprof](https://golang.org/pkg/net/http/pprof/)
  * `/debug/charts`: Go runtime profiling data through [debugchargs](https://github.com/mkevac/debugcharts)
