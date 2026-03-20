// Fisilink FSL91030 cache control
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

// Cache control CSRs (from freeloader.S)
//
// The Nuclei UX600 core uses custom CSRs for cache control:
//   - CSR_MCACHE_CTL (0x7CA): Machine cache control register
//   - CSR_CACHE_ENABLE (0x10001): Enable bit for I/D cache
//
// These are Nuclei-specific extensions not part of standard RISC-V spec.
const (
	CSR_MCACHE_CTL   = 0x7CA   // Machine cache control register
	CSR_CACHE_ENABLE = 0x10001 // I/D cache enable bit
)

// EnableCache enables the instruction and data caches.
//
// Sets the CSR_CACHE_ENABLE bit (0x10001) in the Nuclei-specific
// CSR_MCACHE_CTL register (0x7CA) using the CSRRS instruction.
//
// Based on freeloader.S cache enable sequence:
//
//	li t0, CSR_CACHE_ENABLE
//	csrs CSR_MCACHE_CTL, t0
//
// In the standard boot flow, flashboot.s enables the cache before the Go
// runtime starts. This function is available for explicit cache control,
// for example to re-enable cache after a DisableCache call.
func EnableCache() {
	enableCache()
}

//go:nosplit
func enableCache()

// DisableCache disables the instruction and data caches.
//
// Clears the CSR_CACHE_ENABLE bit (0x10001) in the Nuclei-specific
// CSR_MCACHE_CTL register (0x7CA) using the CSRRC instruction.
//
// Based on freeloader.S cache disable sequence:
//
//	li t0, CSR_CACHE_ENABLE
//	csrc CSR_MCACHE_CTL, t0
//
// Cache must be disabled before DDR controller reconfiguration and may be
// needed for DMA-coherency-sensitive MMIO regions.
func DisableCache() {
	disableCache()
}

//go:nosplit
func disableCache()
