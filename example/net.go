// U-Boot networking hooks
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

// Interim hack which piggybacks on U-Boot net_send_udp_packet() and
// udp_packet_handler to send and receive packets.

package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type in_addr struct {
	s_addr uint32
}

type rxhand_f func([]byte, uint, in_addr, uint, uint)

// u-boot.map addresses + bdinfo->reloc off
var reloc_off uint32 = 0x18789000
var net_init_ptr uint32 = 0x878494ac + reloc_off            // 0x9ffd24ac
var eth_init_ptr uint32 = 0x87848fc4 + reloc_off            // 0x9ffd1fc4
var eth_rx_ptr uint32 = 0x878490a4 + reloc_off              // 0x9ffd20a4
var net_set_udp_handler_ptr uint32 = 0x87849684 + reloc_off // 0x9ffd2684
var net_send_udp_packet_ptr uint32 = 0x87849e8c + reloc_off // 0x9ffd2e8c

// defined in net.s
func net_init()
func eth_init()
func eth_rx()
func net_set_udp_handler()
func net_send_udp_packet(*byte, in_addr, int, int, int) int

func udp_handler(pkt []byte, dport uint, sip in_addr, sport uint, len uint) {
	ip := make([]byte, 4)
	binary.LittleEndian.PutUint32(ip, sip.s_addr)

	fmt.Printf("received packet from %s:%d -> tamago:%d %x\n", net.IP(ip), sport, dport, pkt)

	exit <- true
}

func TestNet() {
	fmt.Printf("net_init (%x)\n", net_init_ptr)
	net_init()

	fmt.Printf("eth_init (%x)\n", eth_init_ptr)
	eth_init()

	fmt.Printf("net_set_udp_handler (%x)\n", net_set_udp_handler_ptr)
	net_set_udp_handler()

	pkt := []byte{0xaa, 0xbb, 0xcc, 0xdd}
	dest := in_addr{s_addr: 0x11223344}

	fmt.Printf("starting eth_rx (%x) loop\n", eth_rx_ptr)

	go func() {
		for {
			fmt.Printf("eth_rx (%x)\n", eth_rx_ptr)
			eth_rx()

			fmt.Printf("net_send_udp_packet (%x)\n", net_send_udp_packet_ptr)
			net_send_udp_packet(&pkt[0], dest, 80, 8080, len(pkt))

			time.Sleep(1000 * time.Millisecond)
		}
	}()
}
