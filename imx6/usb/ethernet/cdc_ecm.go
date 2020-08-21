// Ethernet over USB driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ethernet implements a driver for Ethernet over USB emulation on
// i.MX6 SoCs.
//
// It currently implements CDC-ECM networking and for this reason the Ethernet
// device is only supported on Linux hosts. Applications are meant to use the
// driver in combination with gVisor tcpip package to expose TCP/IP networking
// stack through Ethernet over USB.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package ethernet

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/f-secure-foundry/tamago/imx6/usb"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// NIC represents an virtual Ethernet instance.
type NIC struct {
	// Host MAC address
	Host net.HardwareAddr

	// Device MAC address
	Device net.HardwareAddr

	// Link is a gVisor channel endpoint
	Link *channel.Endpoint

	// Rx is tendpoint 1 OUT function, set by Init() to ECMRx if not
	// already defined.
	Rx func([]byte, error) ([]byte, error)

	// Tx is endpoint 1 IN function, set by Init() to ECMTx if not alread
	// defined.
	Tx func([]byte, error) ([]byte, error)

	// Control is endpoint 2 IN function, set by Init() to ECMControl if
	// not already defined.
	Control func([]byte, error) ([]byte, error)

	maxPacketSize int
	buf           []byte
}

// Init initializes a virtual Ethernet instance on a specific USB device and
// configuration index.
func (eth *NIC) Init(device *usb.Device, configurationIndex int) (err error) {
	if eth.Link == nil {
		return errors.New("missing link endpoint")
	}

	if len(eth.Host) != 6 || len(eth.Device) != 6 {
		return errors.New("invalid MAC address")
	}

	if eth.Rx == nil {
		eth.Rx = eth.ECMRx
	}

	if eth.Tx == nil {
		eth.Tx = eth.ECMTx
	}

	if eth.Control == nil {
		eth.Control = eth.ECMControl
	}

	addControlInterface(device, configurationIndex, eth)
	dataInterface := addDataInterface(device, configurationIndex, eth)

	eth.maxPacketSize = int(dataInterface.Endpoints[0].MaxPacketSize)

	return
}

// ECMControl implements the endpoint 2 IN function.
func (eth *NIC) ECMControl(_ []byte, lastErr error) (in []byte, err error) {
	// ignore for now
	return
}

// ECMRx implements the endpoint 1 OUT function, used to receive Ethernet
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
