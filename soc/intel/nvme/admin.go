// NVM Express driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package nvme

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/usbarmory/tamago/bits"
)

// Admin command opcodes
const (
	ADMIN_DELETE_IO_SQ = 0x00
	ADMIN_CREATE_IO_SQ = 0x01
	ADMIN_DELETE_IO_CQ = 0x04
	ADMIN_CREATE_IO_CQ = 0x05
	ADMIN_IDENTIFY     = 0x06
	ADMIN_SET_FEATURES = 0x09
	ADMIN_GET_FEATURES = 0x0a
)

// Identify Controller or Namespace Structure (CNS) values
const (
	CNS_NAMESPACE        = 0x00
	CNS_CONTROLLER       = 0x01
	CNS_ACTIVE_NAMESPACE = 0x02
)

// Feature identifiers
const (
	FEATURE_VOLATILE_WRITE_CACHE = 0x06
	FEATURE_NUMBER_OF_QUEUES     = 0x07
)

// Admin command dwords
// (NVM Express Base Specification 2.0 - 5 Admin Command Set).
const (
	// Create I/O Completion/Submission Queue - Command Dword 10
	CREATE_QUEUE_QSIZE = 16
	CREATE_QUEUE_QID   = 0

	// Create I/O Completion Queue - Command Dword 11
	CREATE_CQ_IV  = 16
	CREATE_CQ_IEN = 1
	CREATE_CQ_PC  = 0

	// Create I/O Submission Queue - Command Dword 11
	CREATE_SQ_CQID  = 16
	CREATE_SQ_QPRIO = 1
	CREATE_SQ_PC    = 0

	// Number of Queues - Command Dword 11 and Dword 0 of the completion
	NUMBER_OF_QUEUES_NCQ = 16
	NUMBER_OF_QUEUES_NSQ = 0
)

// Identify Controller data structure offsets
// (NVM Express Base Specification 2.0 - Figure 275).
const (
	ID_CTRL_VID    = 0
	ID_CTRL_SN     = 4
	ID_CTRL_MN     = 24
	ID_CTRL_FR     = 64
	ID_CTRL_MDTS   = 77
	ID_CTRL_CNTLID = 78
	ID_CTRL_NN     = 516
	ID_CTRL_ONCS   = 520
	ID_CTRL_VWC    = 525
	VWC_PRESENT    = 0
)

// Identify Namespace data structure offsets
// (NVM Express NVM Command Set Specification 1.0 - Figure 97).
const (
	ID_NS_NSZE = 0
	ID_NS_NCAP = 8

	// Formatted LBA Size
	ID_NS_FLBAS  = 26
	FLBAS_LBAF_L = 0
	FLBAS_LBAF_H = 5

	// LBA Format Data Structure
	ID_NS_LBAF = 128
	LBAF_RP    = 24
	LBAF_LBADS = 16
	LBAF_MS    = 0
)

// ControllerInfo represents a subset of the Identify Controller data
// structure.
type ControllerInfo struct {
	VendorID     uint16
	Serial       string
	Model        string
	Firmware     string
	ControllerID uint16
	// MDTS is the Maximum Data Transfer Size as a power of two multiple of
	// the minimum memory page size (0 means no limit).
	MDTS uint8
	// NN is the maximum namespace identifier.
	NN uint32
	// ONCS is the Optional NVM Command Support bitmap.
	ONCS uint16
	// VolatileWriteCache reports whether a volatile write cache is
	// present, in which case [Namespace.Flush] (or FUA writes) are
	// required for durability.
	VolatileWriteCache bool
}

func trim(b []byte) string {
	return strings.TrimRight(string(b), " \x00")
}

func (hw *NVMe) identify(cns uint32, nsid uint32) (data []byte, err error) {
	data = make([]byte, pageSize)

	addr := hw.Region.Alloc(data, pageSize)
	defer hw.Region.Free(addr)

	cmd := command{
		opcode: ADMIN_IDENTIFY,
		nsid:   nsid,
		addr:   addr,
		length: pageSize,
	}
	cmd.cdw[0] = cns

	if _, err = hw.adminCommand(cmd); err != nil {
		return nil, err
	}

	hw.Region.Read(addr, 0, data)

	return
}

func (hw *NVMe) identifyController() (err error) {
	buf, err := hw.identify(CNS_CONTROLLER, 0)

	if err != nil {
		return
	}

	hw.Info = &ControllerInfo{
		VendorID:     binary.LittleEndian.Uint16(buf[ID_CTRL_VID:]),
		Serial:       trim(buf[ID_CTRL_SN:ID_CTRL_MN]),
		Model:        trim(buf[ID_CTRL_MN:ID_CTRL_FR]),
		Firmware:     trim(buf[ID_CTRL_FR : ID_CTRL_FR+8]),
		MDTS:         buf[ID_CTRL_MDTS],
		ControllerID: binary.LittleEndian.Uint16(buf[ID_CTRL_CNTLID:]),
		NN:           binary.LittleEndian.Uint32(buf[ID_CTRL_NN:]),
		ONCS:         binary.LittleEndian.Uint16(buf[ID_CTRL_ONCS:]),
	}

	vwc := uint32(buf[ID_CTRL_VWC])
	hw.Info.VolatileWriteCache = bits.Get(&vwc, VWC_PRESENT)

	// largest power of two transfer within MDTS and the PRP list limit
	hw.maxXfer = (maxTransferPages / 2) * pageSize // 1 MiB

	if mdts := hw.Info.MDTS; mdts > 0 && mdts < 32 {
		hw.maxXfer = min(hw.maxXfer, pageSize<<mdts)
	}

	return
}

func (hw *NVMe) initIOQueues() (err error) {
	// queue counts are 0's based values
	cmd := command{opcode: ADMIN_SET_FEATURES}
	cmd.cdw[0] = FEATURE_NUMBER_OF_QUEUES
	bits.SetN(&cmd.cdw[1], NUMBER_OF_QUEUES_NCQ, 0xffff, uint32(hw.nq-1))
	bits.SetN(&cmd.cdw[1], NUMBER_OF_QUEUES_NSQ, 0xffff, uint32(hw.nq-1))

	res, err := hw.adminCommand(cmd)

	if err != nil {
		return
	}

	ncq := bits.GetN(&res, NUMBER_OF_QUEUES_NCQ, 0xffff)
	nsq := bits.GetN(&res, NUMBER_OF_QUEUES_NSQ, 0xffff)
	granted := int(min(ncq, nsq)) + 1
	count := min(hw.nq, granted)

	for i := range count {
		qid := uint16(i + 1)

		q, err := hw.newQueue(qid, hw.QueueDepth)

		if err != nil {
			return err
		}

		// queue size is a 0's based value
		var cdw10 uint32
		bits.SetN(&cdw10, CREATE_QUEUE_QSIZE, 0xffff, uint32(q.depth-1))
		bits.SetN(&cdw10, CREATE_QUEUE_QID, 0xffff, uint32(qid))

		// physically contiguous, interrupts enabled on MSI-X entry qid
		cq := command{opcode: ADMIN_CREATE_IO_CQ, addr: q.cqAddr}
		cq.cdw[0] = cdw10
		bits.SetN(&cq.cdw[1], CREATE_CQ_IV, 0xffff, uint32(qid))
		bits.Set(&cq.cdw[1], CREATE_CQ_IEN)
		bits.Set(&cq.cdw[1], CREATE_CQ_PC)

		if _, err = hw.adminCommand(cq); err != nil {
			q.release()
			return fmt.Errorf("create CQ %d, %v", qid, err)
		}

		// physically contiguous, bound to the matching CQ, QPRIO is
		// ignored with round robin arbitration
		sq := command{opcode: ADMIN_CREATE_IO_SQ, addr: q.sqAddr}
		sq.cdw[0] = cdw10
		bits.SetN(&sq.cdw[1], CREATE_SQ_CQID, 0xffff, uint32(qid))
		bits.Set(&sq.cdw[1], CREATE_SQ_PC)

		if _, err = hw.adminCommand(sq); err != nil {
			q.release()
			return fmt.Errorf("create SQ %d, %v", qid, err)
		}

		hw.io = append(hw.io, q)
	}

	return
}

func (hw *NVMe) identifyNamespaces() (err error) {
	list, err := hw.identify(CNS_ACTIVE_NAMESPACE, 0)

	if err != nil {
		return
	}

	hw.Namespaces = nil

	for off := 0; off < len(list); off += 4 {
		nsid := binary.LittleEndian.Uint32(list[off:])

		if nsid == 0 {
			break
		}

		buf, err := hw.identify(CNS_NAMESPACE, nsid)

		if err != nil {
			return err
		}

		blocks := binary.LittleEndian.Uint64(buf[ID_NS_NSZE:])

		if blocks == 0 {
			continue
		}

		// current LBA format index
		flbas := uint32(buf[ID_NS_FLBAS])
		format := bits.GetN(&flbas, FLBAS_LBAF_H, 0b11)<<4 | bits.GetN(&flbas, FLBAS_LBAF_L, 0xf)
		lbaf := binary.LittleEndian.Uint32(buf[ID_NS_LBAF+format*4:])

		ns := &Namespace{
			hw:       hw,
			ID:       nsid,
			Blocks:   blocks,
			Capacity: binary.LittleEndian.Uint64(buf[ID_NS_NCAP:]),
			Metadata: int(bits.GetN(&lbaf, LBAF_MS, 0xffff)),
		}

		// LBA data size is reported as a power of two
		if lbads := bits.GetN(&lbaf, LBAF_LBADS, 0xff); lbads >= 9 && lbads < 32 {
			ns.BlockSize = 1 << lbads
		} else {
			continue
		}

		ns.maxBlocks = min(hw.maxXfer/ns.BlockSize, maxBlocksPerCommand)
		hw.Namespaces = append(hw.Namespaces, ns)
	}

	return
}
