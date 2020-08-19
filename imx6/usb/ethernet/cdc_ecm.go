// Ethernet over USB driver - CDC ECM
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ethernet

import (
	"encoding/binary"
	"net"

	"github.com/f-secure-foundry/tamago/imx6/usb"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// NIC represents a virtual Ethernet instance.
type NIC struct {
	// Host MAC address
	Host net.HardwareAddr
	// Device MAC address
	Device net.HardwareAddr

	// Link is a gVisor channel endpoint
	Link *channel.Endpoint

	// Rx is tendpoint 1 OUT function, set at initialization to ECMRx.
	Rx func([]byte, error) ([]byte, error)
	// Tx is endpoint 1 IN function, set at initialization to ECMTx.
	Tx func([]byte, error) ([]byte, error)
	// Control is endpoint 2 IN function, set at initialization to ECMControl.
	Control func([]byte, error) ([]byte, error)

	// maximum packet size for EP1
	maxPacketSize int
	// incoming Ethernet frames buffer
	buf []byte
}

// Init initializes a virtual Ethernet instance on a specific USB device and
// configuration index.
func (eth *NIC) Init(device *usb.Device, configurationIndex int) {
	eth.Rx = eth.ECMRx
	eth.Tx = eth.ECMTx
	eth.Control = eth.ECMControl

	controlInterface := eth.buildControlInterface(device)
	device.Configurations[configurationIndex].AddInterface(controlInterface)

	dataInterface := eth.buildDataInterface(device)
	device.Configurations[configurationIndex].AddInterface(dataInterface)

	eth.maxPacketSize = int(dataInterface.Endpoints[0].MaxPacketSize)
}

// ECMControl implements the endpoint 2 IN function.
func (eth *NIC) ECMControl(_ []byte, lastErr error) (in []byte, err error) {
	// ignore for now
	return
}

// ECMTx implements the endpoint 1 IN function, used to transmit Ethernet
// packet from device to host.
func (eth *NIC) ECMTx(_ []byte, lastErr error) (in []byte, err error) {
	info, valid := eth.Link.Read()

	if !valid {
		return
	}

	hdr := info.Pkt.Header.View()
	payload := info.Pkt.Data.ToView()

	proto := make([]byte, 2)
	binary.BigEndian.PutUint16(proto, uint16(info.Proto))

	// Ethernet frame header
	in = append(in, eth.Host...)
	in = append(in, eth.Device...)
	in = append(in, proto...)
	// packet header
	in = append(in, hdr...)
	// payload
	in = append(in, payload...)

	return
}

// ECMRx implements the endpoint 1 OUT function, used to receive ethernet
// packet from host to device.
func (eth *NIC) ECMRx(out []byte, lastErr error) (_ []byte, err error) {
	if len(eth.buf) == 0 && len(out) < 14 {
		return
	}

	eth.buf = append(eth.buf, out...)

	// more data expected or zero length packet
	if len(out) == eth.maxPacketSize {
		return
	}

	hdr := buffer.NewViewFromBytes(eth.buf[0:14])
	proto := tcpip.NetworkProtocolNumber(binary.BigEndian.Uint16(eth.buf[12:14]))
	payload := buffer.NewViewFromBytes(eth.buf[14:])

	pkt := &stack.PacketBuffer{
		LinkHeader: hdr,
		Data:       payload.ToVectorisedView(),
	}

	eth.Link.InjectInbound(proto, pkt)
	eth.buf = []byte{}

	return
}
