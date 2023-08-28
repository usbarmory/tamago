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
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// Endpoint represents a USB 2.0 endpoint.
type endpoint struct {
	sync.Mutex

	bus  *USB
	desc *EndpointDescriptor

	n   int
	dir int

	res []byte
	err error
}

func (ep *endpoint) rx() {
	var buf []byte

	buf, ep.err = ep.bus.rx(ep.n, ep.res)

	if ep.err == nil && len(buf) != 0 {
		ep.res, ep.err = ep.desc.Function(buf, ep.err)
	}
}

func (ep *endpoint) tx() {
	ep.res, ep.err = ep.desc.Function(nil, ep.err)

	if ep.err == nil && len(ep.res) != 0 {
		ep.err = ep.bus.tx(ep.n, ep.res)
	}
}

// Init initializes an endpoint.
func (ep *endpoint) Init() {
	ep.n = ep.desc.Number()
	ep.dir = ep.desc.Direction()

	ep.bus.set(ep.n, ep.dir, int(ep.desc.MaxPacketSize), ep.desc.Zero, 0)
	ep.bus.enable(ep.n, ep.dir, ep.desc.TransferType())
}

// Flush clears the endpoint receive and transmit buffers.
func (ep *endpoint) Flush() {
	reg.Set(ep.bus.flush, (ep.dir*16)+ep.n)
}

// Start initializes and runs an USB endpoint.
func (ep *endpoint) Start() {
	if ep.desc.Function == nil {
		return
	}

	ep.Lock()

	defer func() {
		ep.Flush()
		ep.bus.wg.Done()
		ep.Unlock()
	}()

	ep.Init()

	for {
		runtime.Gosched()

		if ep.dir == OUT {
			ep.rx()
		} else {
			ep.tx()
		}

		if ep.err != nil {
			ep.bus.stall(ep.n, ep.dir)
		}

		select {
		case <-ep.bus.exit:
			return
		default:
		}
	}
}

func (hw *USB) startEndpoints() {
	if hw.Device.ConfigurationValue == 0 {
		return
	}

	hw.exit = make(chan struct{})

	for _, conf := range hw.Device.Configurations {
		if hw.Device.ConfigurationValue != conf.ConfigurationValue {
			continue
		}

		for _, iface := range conf.Interfaces {
			for _, desc := range iface.Endpoints {
				ep := &endpoint{
					bus:  hw,
					desc: desc,
				}

				hw.wg.Add(1)

				go func(ep *endpoint) {
					ep.Start()
				}(ep)
			}
		}
	}
}
