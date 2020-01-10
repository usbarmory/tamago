// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package main

import (
	"encoding/binary"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/inversepath/tamago/imx6/usb"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const maxPacketSize = 512
const hostMAC = "1a:55:89:a2:69:42"
const deviceMAC = "1a:55:89:a2:69:41"
const IP = "10.0.0.1"
const MTU = 1500

// set to true to enable packet sniffing
const sniff = false

// populated by setupStack()
var hostMACBytes []byte
var deviceMACBytes []byte
var link *channel.Endpoint

// Ethernet frame buffers
var rx []byte

func configureEthernetDevice(device *usb.Device) {
	// Supported Language Code Zero: English
	device.SetLanguageCodes([]uint16{0x0409})

	// device descriptor
	device.Descriptor = &usb.DeviceDescriptor{}
	device.Descriptor.SetDefaults()
	device.Descriptor.DeviceClass = 0x2
	device.Descriptor.VendorId = 0x0525
	device.Descriptor.ProductId = 0xa4a2
	device.Descriptor.Device = 0x0001
	device.Descriptor.NumConfigurations = 1

	iManufacturer, _ := device.AddString(`TamaGo`)
	device.Descriptor.Manufacturer = iManufacturer

	iProduct, _ := device.AddString(`RNDIS/Ethernet Gadget`)
	device.Descriptor.Product = iProduct

	iSerial, _ := device.AddString(`0.1`)
	device.Descriptor.SerialNumber = iSerial

	// device qualifier
	device.Qualifier = &usb.DeviceQualifierDescriptor{}
	device.Qualifier.SetDefaults()
	device.Qualifier.DeviceClass = 2
	device.Qualifier.NumConfigurations = 2
}

func configureECM(device *usb.Device) {
	// source and sink configuration
	conf := &usb.ConfigurationDescriptor{}
	conf.SetDefaults()
	conf.TotalLength = 71
	conf.NumInterfaces = 1
	conf.ConfigurationValue = 1

	device.Configurations = append(device.Configurations, conf)

	// CDC communication interface
	iface := &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.NumEndpoints = 1
	iface.InterfaceClass = 2
	iface.InterfaceSubClass = 6

	iInterface, _ := device.AddString(`CDC Ethernet Control Model (ECM)`)
	iface.Interface = iInterface

	header := &usb.CDCHeaderDescriptor{}
	header.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, header.Bytes())

	union := &usb.CDCUnionDescriptor{}
	union.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, union.Bytes())

	ethernet := &usb.CDCEthernetDescriptor{}
	ethernet.SetDefaults()

	iMacAddress, _ := device.AddString(strings.ReplaceAll(hostMAC, ":", ""))
	ethernet.MacAddress = iMacAddress

	iface.ClassDescriptors = append(iface.ClassDescriptors, ethernet.Bytes())

	conf.Interfaces = append(conf.Interfaces, iface)

	ep2IN := &usb.EndpointDescriptor{}
	ep2IN.SetDefaults()
	ep2IN.EndpointAddress = 0x82
	ep2IN.Attributes = 3
	ep2IN.MaxPacketSize = 16
	ep2IN.Interval = 9
	ep2IN.Function = ECMControl

	iface.Endpoints = append(iface.Endpoints, ep2IN)

	// CDC data interface
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.AlternateSetting = 1
	iface.NumEndpoints = 2
	iface.InterfaceClass = 10

	iInterface, _ = device.AddString(`CDC Data`)
	iface.Interface = iInterface

	conf.Interfaces = append(conf.Interfaces, iface)

	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = maxPacketSize
	ep1IN.Function = ECMTx

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = maxPacketSize
	ep1OUT.Function = ECMRx

	iface.Endpoints = append(iface.Endpoints, ep1OUT)
}

func configureNetworkStack(addr tcpip.Address, nic tcpip.NICID, sniff bool) (s *stack.Stack) {
	var err error

	hostMACBytes, err = net.ParseMAC(hostMAC)

	if err != nil {
		log.Fatal(err)
	}

	deviceMACBytes, err = net.ParseMAC(deviceMAC)

	if err != nil {
		log.Fatal(err)
	}

	s = stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocol{
			ipv4.NewProtocol(),
			arp.NewProtocol()},
		TransportProtocols: []stack.TransportProtocol{
			tcp.NewProtocol(),
			udp.NewProtocol(),
			icmp.NewProtocol4()},
	})

	linkAddr, err := tcpip.ParseMACAddress(deviceMAC)

	if err != nil {
		log.Fatal(err)
	}

	link = channel.New(256, MTU, linkAddr)
	linkEP := stack.LinkEndpoint(link)

	if sniff {
		linkEP = sniffer.New(linkEP)
	}

	if err := s.CreateNIC(nic, linkEP); err != nil {
		log.Fatal(err)
	}

	if err := s.AddAddress(nic, arp.ProtocolNumber, arp.ProtocolAddress); err != nil {
		log.Fatal(err)
	}

	if err := s.AddAddress(nic, ipv4.ProtocolNumber, addr); err != nil {
		log.Fatal(err)
	}

	subnet, err := tcpip.NewSubnet("\x00\x00\x00\x00", "\x00\x00\x00\x00")

	if err != nil {
		log.Fatal(err)
	}

	s.SetRouteTable([]tcpip.Route{{
		Destination: subnet,
		NIC:         nic,
	}})

	return
}

func startICMPEndpoint(s *stack.Stack, addr tcpip.Address, port uint16, nic tcpip.NICID) {
	var wq waiter.Queue

	fullAddr := tcpip.FullAddress{Addr: addr, Port: port, NIC: nic}
	ep, err := s.NewEndpoint(icmp.ProtocolNumber4, ipv4.ProtocolNumber, &wq)

	if err != nil {
		log.Fatalf("endpoint error (icmp): %v\n", err)
	}

	if err := ep.Bind(fullAddr); err != nil {
		log.Fatal("bind error (icmp endpoint): ", err)
	}
}

func startUDPListener(s *stack.Stack, addr tcpip.Address, port uint16, nic tcpip.NICID) (conn *gonet.PacketConn) {
	var err error

	fullAddr := tcpip.FullAddress{Addr: addr, Port: port, NIC: nic}
	conn, err = gonet.DialUDP(s, &fullAddr, nil, ipv4.ProtocolNumber)

	if err != nil {
		log.Fatal("listener error: ", err)
	}

	return
}

func startEchoServer(s *stack.Stack, addr tcpip.Address, port uint16, nic tcpip.NICID) {
	c := startUDPListener(s, addr, port, nic)

	for {
		runtime.Gosched()

		buf := make([]byte, 1024)
		n, addr, err := c.ReadFrom(buf)

		if err != nil {
			log.Printf("udp recv error, %v\n", err)
			continue
		}

		_, err = c.WriteTo(buf[0:n], addr)

		if err != nil {
			log.Printf("udp send error, %v\n", err)
		}
	}
}

// ECMControl implements the endpoint 2 IN function.
func ECMControl(_ []byte, lastErr error) (in []byte, err error) {
	// ignore for now
	return
}

// ECMTx implements the endpoint 1 IN function, used to transmit Ethernet
// packet from device to host.
func ECMTx(_ []byte, lastErr error) (in []byte, err error) {
	select {
	case info := <-link.C:
		hdr := info.Pkt.Header.View()
		payload := info.Pkt.Data.ToView()

		proto := make([]byte, 2)
		binary.BigEndian.PutUint16(proto, uint16(info.Proto))

		// Ethernet frame header
		in = append(in, hostMACBytes...)
		in = append(in, deviceMACBytes...)
		in = append(in, proto...)
		// packet header
		in = append(in, hdr...)
		// payload
		in = append(in, payload...)
	default:
	}

	return
}

// ECMRx implements the endpoint 1 OUT function, used to receive ethernet
// packet from host to device.
func ECMRx(out []byte, lastErr error) (_ []byte, err error) {
	if len(rx) == 0 && len(out) < 14 {
		return
	}

	rx = append(rx, out...)

	// more data expected or zero length packet
	if len(out) == 512 {
		return
	}

	hdr := buffer.NewViewFromBytes(rx[0:14])
	proto := tcpip.NetworkProtocolNumber(binary.BigEndian.Uint16(rx[12:14]))
	payload := buffer.NewViewFromBytes(rx[14:])

	pkt := tcpip.PacketBuffer{
		LinkHeader: hdr,
		Data:       payload.ToVectorisedView(),
	}

	link.InjectInbound(proto, pkt)

	rx = []byte{}

	return
}

// StartUSBEthernet starts an emulated Ethernet over USB device (ECM protocol,
// only supported on Linux hosts) with a test UDP echo service on port 1234.
func StartUSBEthernet() {
	addr := tcpip.Address(net.ParseIP(IP)).To4()

	s := configureNetworkStack(addr, 1, sniff)

	// handle pings
	startICMPEndpoint(s, addr, 0, 1)

	go func() {
		// UDP echo server
		startEchoServer(s, addr, 1234, 1)
	}()

	go func() {
		// HTTP web server (see web_server.go)
		startWebServer(s, addr, 80, 1, false)
	}()

	go func() {
		file, _ := os.OpenFile("/index.html", os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)
		file.WriteString("<html><body>")
		file.WriteString(banner)
		file.WriteString("</body></html>")
		file.Close()

		staticHandler := http.FileServer(http.Dir("/"))
		http.Handle("/", http.StripPrefix("/", staticHandler))

		// HTTPS web server (see web_server.go)
		startWebServer(s, addr, 443, 1, true)
	}()

	device := &usb.Device{}

	configureEthernetDevice(device)
	configureECM(device)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}
