// NXP i.MX6 CCM driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

const (
	// base addresses

	CCM_BASE = 0x020c4000

	// Register definitions

	CCR      uint32 = 0x0000
	CCDR            = 0x0004
	CSR             = 0x0008
	CCSR            = 0x000c
	CACRR           = 0x0010
	CBCRR           = 0x0014
	CBCMR           = 0x0018
	CSCMR1          = 0x001c
	CSCMR2          = 0x0020
	CSCDR1          = 0x0024
	CS1CDR          = 0x0028
	CS2CDR          = 0x002c
	CDCDR           = 0x0030
	CHSCCDR         = 0x0034
	CHSCCDR2        = 0x0038
	CHSCCDR3        = 0x003c
	CDHIPR          = 0x0048
	CLPCR           = 0x0054
	CISR            = 0x0058
	CIMR            = 0x005c
	CCOSR           = 0x0060
	CGPR            = 0x0064
	CCGR0           = 0x0068
	CCGR1           = 0x006c
	CCGR2           = 0x0070
	CCGR3           = 0x0074
	CCGR4           = 0x0078
	CCGR5           = 0x007c
	CCGR6           = 0x0080
	CMEOR           = 0x0084

	// bit positions
	CSCDR1_CLK_PODF = 0
	CSCDR1_UART_CLK_SEL = 6
	// misc
	OSC_CLK = 24000000
)

type ccm struct {
	ccr      uint32
	ccdr     uint32
	csr      uint32
	ccsr     uint32
	cacrr    uint32
	cbcrr    uint32
	cbcmr    uint32
	cscmr1   uint32
	cscmr2   uint32
	cscdr1   uint32
	cs1cdr   uint32
	cs2cdr   uint32
	cdcdr    uint32
	chsccdr  uint32
	chsccdr2 uint32
	chsccdr3 uint32
	cdhipr   uint32
	clpcr    uint32
	cisr     uint32
	cimr     uint32
	ccosr    uint32
	cgpr     uint32
	ccgr0    uint32
	ccgr1    uint32
	ccgr2    uint32
	ccgr3    uint32
	ccgr4    uint32
	ccgr5    uint32
	ccgr6    uint32
	cmeor    uint32
}

func (c *ccm) Init(base uint32) {
	c.ccr      = base + CCR
	c.ccdr     = base + CCDR
	c.csr      = base + CSR
	c.ccsr     = base + CCSR
	c.cacrr    = base + CACRR
	c.cbcrr    = base + CBCRR
	c.cbcmr    = base + CBCMR
	c.cscmr1   = base + CSCMR1
	c.cscmr2   = base + CSCMR2
	c.cscdr1   = base + CSCDR1
	c.cs1cdr   = base + CS1CDR
	c.cs2cdr   = base + CS2CDR
	c.cdcdr    = base + CDCDR
	c.chsccdr  = base + CHSCCDR
	c.chsccdr2 = base + CHSCCDR2
	c.chsccdr3 = base + CHSCCDR3
	c.cdhipr   = base + CDHIPR
	c.clpcr    = base + CLPCR
	c.cisr     = base + CISR
	c.cimr     = base + CIMR
	c.ccosr    = base + CCOSR
	c.cgpr     = base + CGPR
	c.ccgr0    = base + CCGR0
	c.ccgr1    = base + CCGR1
	c.ccgr2    = base + CCGR2
	c.ccgr3    = base + CCGR3
	c.ccgr4    = base + CCGR4
	c.ccgr5    = base + CCGR5
	c.ccgr6    = base + CCGR6
	c.cmeor    = base + CMEOR
}
