// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Interim hack which piggybacks on U-Boot net_send_udp_packet() and
// udp_packet_handler to send and receive packets.
//
// We should in principle preserve R10/g when calling external code, but we
// know that U-Boot is never using them.

// func net_init() int
TEXT main·net_init(SB),$0
	MOVW	main·net_init_ptr(SB), R11
	BL	(R11)
	RET

// func eth_init() int
TEXT main·eth_init(SB),$0
	MOVW	main·eth_init_ptr(SB), R11
	BL	(R11)
	RET

// func eth_rx()
TEXT main·eth_rx(SB),$0
	MOVW	main·eth_rx_ptr(SB), R11
	BL	(R11)
	RET

// func udp_handler_entry([]byte, uint, in_addr, uint, uint)
TEXT main·udp_handler_entry(SB),$0
	MOVW	R0, pkt+0(FP)
	MOVW	R1, dport+4(FP)
	MOVW	R2, source+8(FP)
	MOVW	R3, sport+12(FP)

	WORD	$0xe49d4004  // pop {r4}
	MOVW	R4, len+16(FP)

	BL main·udp_handler(SB)

// func net_set_udp_handler(handler *rxhand_f)
TEXT main·net_set_udp_handler(SB),$0
	MOVW	main·udp_handler_entry(SB), R0

	MOVW	main·net_set_udp_handler_ptr(SB), R11
	BL	(R11)

	RET

// func net_send_udp_packet([]byte, in_addr, int, int, int) int
TEXT main·net_send_udp_packet(SB),$0
	MOVW	pkt+0(FP), R0
	MOVW	dest+4(FP), R1
	MOVW	dport+8(FP), R2
	MOVW	sport+12(FP), R3

	MOVW	len+16(FP), R4
	WORD	$0xe52d4004  // push {r4}

	MOVW	main·net_send_udp_packet_ptr(SB), R11
	BL	(R11)

	WORD	$0xe49d4004  // pop {r4}

	MOVW R0, ret+20(FP)
	RET
