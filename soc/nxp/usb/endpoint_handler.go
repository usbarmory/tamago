// USB device mode support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"log"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// Endpoint represents a USB 2.0 endpoint.
type Endpoint struct {
	sync.Mutex

	bus  *USB
	wg   *sync.WaitGroup
	desc *EndpointDescriptor

	n   int
	dir int
}

// Init initializes an endpoint.
func (ep *Endpoint) Init() {
	ep.n = ep.desc.Number()
	ep.dir = ep.desc.Direction()

	ep.bus.set(ep.n, ep.dir, int(ep.desc.MaxPacketSize), ep.desc.Zero, 0)
	ep.bus.enable(ep.n, ep.dir, ep.desc.TransferType())
}

// Flush clears the endpoint receive and transmit buffers.
func (ep *Endpoint) Flush() {
	reg.Set(ep.bus.flush, (ep.dir*16)+ep.n)
}

// Start initializes and runs an USB endpoint.
func (ep *Endpoint) Start() {
	var err error
	var buf []byte
	var res []byte

	if ep.desc.Function == nil {
		return
	}

	ep.Lock()

	defer func() {
		ep.Flush()
		ep.wg.Done()
		ep.Unlock()
	}()

	ep.Init()

	for {
		runtime.Gosched()

		if ep.dir == OUT {
			buf, err = ep.bus.rx(ep.n, false, res)

			if err == nil && len(buf) != 0 {
				res, err = ep.desc.Function(buf, err)
			}
		} else {
			res, err = ep.desc.Function(nil, err)

			if err == nil && len(res) != 0 {
				err = ep.bus.tx(ep.n, false, res)
			}
		}

		if err != nil {
			ep.Flush()
			log.Printf("usb: EP%d.%d transfer error, %v", ep.n, ep.dir, err)
		}

		select {
		case <-ep.bus.done:
			return
		default:
		}
	}
}

func (hw *USB) startEndpoints(wg *sync.WaitGroup, dev *Device, configurationValue uint8) {
	if configurationValue == 0 {
		return
	}

	hw.done = make(chan bool)

	for _, conf := range dev.Configurations {
		if configurationValue != conf.ConfigurationValue {
			continue
		}

		for _, iface := range conf.Interfaces {
			for _, desc := range iface.Endpoints {
				ep := &Endpoint{
					wg:   wg,
					bus:  hw,
					desc: desc,
				}

				wg.Add(1)

				go func(ep *Endpoint) {
					ep.Start()
				}(ep)
			}
		}
	}
}
