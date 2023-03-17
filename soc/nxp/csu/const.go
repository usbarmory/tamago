// NXP i.MX Central Security Unit (CSU) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package csu

// p383, 11.2.3.1 CSU Peripheral Access Policy, IMX6ULLRM
const (
	// NonSecure User: RW, NonSecure Supervisor: RW, Secure User: RW, Secure Supervisor: RW
	SEC_LEVEL_0 = 0b11111111
	// NonSecure User: NA, NonSecure Supervisor: RW, Secure User: RW, Secure Supervisor: RW
	SEC_LEVEL_1 = 0b10111011
	// NonSecure User: RO, NonSecure Supervisor: RO, Secure User: RW, Secure Supervisor: RW
	SEC_LEVEL_2 = 0b00111111
	// NonSecure User: NA, NonSecure Supervisor: RO, Secure User: RW, Secure Supervisor: RW
	SEC_LEVEL_3 = 0b00111011
	// NonSecure User: NA, NonSecure Supervisor: NA, Secure User: RW, Secure Supervisor: RW
	SEC_LEVEL_4 = 0b00110011
	// NonSecure User: NA, NonSecure Supervisor: NA, Secure User: NA, Secure Supervisor: RW
	SEC_LEVEL_5 = 0b00100010
	// NonSecure User: NA, NonSecure Supervisor: NA, Secure User: RO, Secure Supervisor: RO
	SEC_LEVEL_6 = 0b00000011
	// NonSecure User: NA, NonSecure Supervisor: NA, Secure User: NA, Secure Supervisor: NA
	SEC_LEVEL_7 = 0b00000000

	// NonSecure Supervisor Write Access bit
	CSL_NW_SUP_WR = 7
	// NonSecure User Write Access bit
	CSL_NW_USR_WR = 6
	// Secure Supervisor Write Access bit
	CSL_SW_SUP_WR = 5
	// Secure User Write Access bit
	CSL_SW_USR_WR = 4
	// NonSecure Supervisor Read Access bit
	CSL_NW_SUP_RD = 3
	// NonSecure User Read Access bit
	CSL_NW_USR_RD = 2
	// Secure Supervisor Read Access bit
	CSL_SW_SUP_RD = 1
	// Secure User Read Access bit
	CSL_SW_USR_RD = 0
)

// The following peripheral, slave identifiers can be used to set the CSL using
// SetSecurityLevel on i.MX6 P/Ns. Note that peripherals presence depends on
// the specific P/Ns:
//
//	|    ID | Blocks (¹UL/ULL/ULZ, ²SX, ³SL, ⁴S/D/DL/Q)                               |
//	|-------|-------------------------------------------------------------------------|
//	|  0, 0 | PWM                                                                     |
//	|  0, 1 | CAN1¹²⁴, DBGMON³                                                        |
//	|  1, 0 | CAN2¹²⁴, QOS³                                                           |
//	|  1, 1 | GPT1, EPIT                                                              |
//	|  2, 0 | GPIO1, GPIO2                                                            |
//	|  2, 1 | GPIO3, GPIO4                                                            |
//	|  3, 0 | GPIO5, GPIO6²⁴                                                          |
//	|  3, 1 | GPIO7²⁴, SNVS_LP¹                                                       |
//	|  4, 0 | KPP                                                                     |
//	|  4, 1 | WDOG1                                                                   |
//	|  5, 0 | WDOG2                                                                   |
//	|  5, 1 | CCM, SNVS_HP, SRC, GPC                                                  |
//	|  6, 0 | ANATOP                                                                  |
//	|  6, 1 | IOMUXC                                                                  |
//	|  7, 0 | IOMUXC_GPR¹², CSI³, TCON³, DCIC⁴                                        |
//	|  7, 1 | SDMA, EPDC⁴⁽ˢ⁄ᴰᴸ⁾,  LCDIF⁴⁽ˢ⁄ᴰᴸ⁾, PXP⁴⁽ˢ⁄ᴰᴸ⁾                            |
//	|  8, 0 | USB                                                                     |
//	|  8, 1 | ENET¹²⁴, FEC³                                                           |
//	|  9, 0 | GPT2¹, MLB²⁴, MSHC³                                                     |
//	|  9, 1 | USDHC1                                                                  |
//	| 10, 0 | USDHC2                                                                  |
//	| 10, 1 | USDHC3²³⁴, SIM1¹⁽ᵁᴸ⁾                                                    |
//	| 11, 0 | USDHC4²³⁴, SIM2¹⁽ᵁᴸ⁾                                                    |
//	| 11, 1 | I2C1                                                                    |
//	| 12, 0 | I2C2                                                                    |
//	| 12, 1 | I2C3                                                                    |
//	| 13, 0 | ROMCP                                                                   |
//	| 13, 1 | MMDC, DCP³, VPU⁴                                                        |
//	| 14, 0 | WEIM¹², EIM³⁴                                                           |
//	| 14, 1 | OCOTP_CTRL                                                              |
//	| 15, 0 | SCTR¹, RDC²                                                             |
//	| 15, 1 | SCTR¹, PERFMON²³⁴                                                       |
//	| 16, 0 | SCTR¹, DBGMON², TZASC1³⁴                                                |
//	| 16, 1 | TZASC1¹, TZASC2²⁴, RNGB³                                                |
//	| 17, 0 | AUDMUX²³⁴, SAI¹²                                                        |
//	| 17, 1 | QSPI¹², ASRC¹, CAAM⁴                                                    |
//	| 18, 0 | SPDIF                                                                   |
//	| 18, 1 | eCSPI1                                                                  |
//	| 19, 0 | eCSPI2                                                                  |
//	| 19, 1 | eCSPI3                                                                  |
//	| 20, 0 | eCSPI4, I2C4¹²                                                          |
//	| 20, 1 | ecSPI5²⁴⁽ˢ⁄ᴰᴸ⁾, IPS_1TO4_MUX¹, UART5³                                   |
//	| 21, 0 | UART1                                                                   |
//	| 21, 1 | UART7¹, UART2³, ESAI²⁴                                                  |
//	| 22, 0 | UART8¹, ESAI¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, SSI1²³⁴                                         |
//	| 22, 1 | SSI2²³⁴                                                                 |
//	| 23, 0 | SSI3²³⁴                                                                 |
//	| 23, 1 | ASRC²⁴, UART3³                                                          |
//	| 24, 0 | CANFD²                                                                  |
//	| 24, 1 | RDC_SEMA4², ROMCP³⁴                                                     |
//	| 25, 0 | WDOG3¹²                                                                 |
//	| 25, 1 | ADC1¹²                                                                  |
//	| 26, 0 | ADC2¹², OCRAM³                                                          |
//	| 27, 0 | APBH_DMA⁴                                                               |
//	| 27, 1 | SEMA4², HDMI⁴                                                           |
//	| 28, 0 | IOMUXC_SNVS¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, MU(A9)², GPU3D⁴                                  |
//	| 28, 1 | IOMUXC_SNVS_GPR¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, CANFD_MEM², PXP³, SATA⁴⁽ˢ⁄ᴰᴸ⁾                |
//	| 29, 0 | UART8¹, MU(M4)², OPENVG³⁴⁽ˢ⁄ᴰᴸ⁾                                         |
//	| 29, 1 | UART6¹², ARM³⁴                                                          |
//	| 30, 0 | UART2¹², EPDC³, HSI⁴                                                    |
//	| 30, 1 | UART3¹², IPU1⁴                                                          |
//	| 31, 0 | UART4¹², LCDIF³, IPU2⁴⁽ˢ⁄ᴰᴸ⁾                                            |
//	| 31, 1 | UART5¹², EIM³, WEIM⁴                                                    |
//	| 32, 0 | LCDIF¹², CSI¹², PXP¹², EPDC¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, VDEC², VADC², DCIC², GIS², PCIE⁴ |
//	| 32, 1 | SPBA², GPU2D⁴                                                           |
//	| 33, 0 | SPBA¹², MIPI_CORE_CSI⁴                                                  |
//	| 33, 1 | TSC¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, MIPI_CORE_HIS⁴                                           |
//	| 34, 0 | DCP¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, VDOA⁴                                                    |
//	| 34, 1 | RNGB¹⁽ᵁᴸᴸ⁄ᵁᴸᶻ⁾, OCRAM², UART2⁴                                          |
//	| 35, 0 | UART3⁴                                                                  |
//	| 35, 1 | UART4⁴                                                                  |
//	| 36, 0 | UART5⁴⁽ˢ⁄ᴰᴸ⁾, I2C4⁴                                                     |
//	| 36, 1 | DTCP⁴                                                                   |
//	| 38, 1 | UART4³                                                                  |
//	| 39, 0 | SPBA³                                                                   |
//	| 39, 1 | OCRAM¹                                                                  |
//
//	¹UL/ULL/ULZ, ²SX, ³SL, ⁴S/D/DL/Q
const (
	CSL_MIN = 0
	CSL_MAX = 39

	// Second slave
	CSL_S2_LOCK = 24
	CSL_S2      = 16
	// First slave
	CSL_S1_LOCK = 8
	CSL_S1      = 0
)

// The following master identifiers can be used to set the SA using
// SetAccess on i.MX6 P/Ns. Note that peripherals presence depends on
// the specific P/Ns:
//
//	|  ID | Blocks (¹UL/ULL/ULZ, ²SX, ³SL, ⁴S/D/DL/Q)                      |
//	|-----|----------------------------------------------------------------|
//	|   0 | CA7¹, CP15⁴                                                    |
//	|   1 | M4², DCP³, SATA⁴                                               |
//	|   2 | SDMA                                                           |
//	|   3 | PXP, CSI²,  LCDIF²³⁴, GPU²³⁴, EPDC³⁴, TCON³, VDOA⁴, IPU⁴, VPU⁴ |
//	|   4 | USB, MLB⁴                                                      |
//	|   5 | TEST, PCIE²⁴                                                   |
//	|   6 | MLB², CSI³                                                     |
//	|   7 | RAWNAND_DMA¹²⁴, MSHC³                                          |
//	|   8 | RAWNAND_APBH_DMA¹², FEC³, ENET⁴                                |
//	|   9 | ENET¹², DAP³⁴                                                  |
//	|  10 | USDHC1                                                         |
//	|  11 | USDHC2                                                         |
//	|  12 | USDHC3²³⁴                                                      |
//	|  13 | USDHC4²³⁴                                                      |
//	|  14 | DCP¹^, DAP¹, HDMI⁴, HSI⁴                                       |
//
// ^ULL/ULZ only, undocumented and found through testing, confirmed by NXP R&D.
const (
	SA_MIN = 0
	SA_MAX = 15
)
