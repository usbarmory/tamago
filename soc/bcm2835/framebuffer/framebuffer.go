// BCM2835 SoC FrameBuffer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package framebuffer

import (
	"encoding/binary"
	"fmt"

	"github.com/usbarmory/tamago/soc/bcm2835"
)

// EDID gets the raw EDID information from the attached monitor
func EDID() []byte {
	outBuf := []byte{}

	for block := uint32(0); true; block++ {
		buf := make([]byte, bcm2835.VC_MEM_GET_EDID_BLOCK_LEN)
		binary.LittleEndian.PutUint32(buf[0:], block)

		msg := &bcm2835.MailboxMessage{
			Tags: []bcm2835.MailboxTag{
				{
					ID:     bcm2835.VC_MEM_GET_EDID_BLOCK,
					Buffer: buf,
				},
			},
		}

		bcm2835.Mailbox.Call(bcm2835.VC_CH_PROPERTYTAGS_A_TO_VC, msg)

		tag := msg.Tag(bcm2835.VC_MEM_GET_EDID_BLOCK)

		if tag == nil || len(tag.Buffer) < 8 {
			return outBuf
		}

		if binary.LittleEndian.Uint32(tag.Buffer[0:]) != block {
			panic("Got EDID data for wrong block")
		}

		status := binary.LittleEndian.Uint32(tag.Buffer[4:])
		if status != 0 {
			break
		}

		outBuf = append(outBuf, tag.Buffer[8:]...)
	}

	return outBuf
}

// PhysicalSize is the dimensions of the current framebuffer in pixels
func PhysicalSize() (width uint32, height uint32) {
	buf := make([]byte, bcm2835.VC_FB_GET_PHYSICAL_SIZE_LEN)

	msg := &bcm2835.MailboxMessage{
		Tags: []bcm2835.MailboxTag{
			{
				ID:     bcm2835.VC_FB_GET_PHYSICAL_SIZE,
				Buffer: buf,
			},
		},
	}

	bcm2835.Mailbox.Call(bcm2835.VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(bcm2835.VC_FB_GET_PHYSICAL_SIZE)

	if tag == nil || len(tag.Buffer) < 8 {
		return 0, 0
	}

	return binary.LittleEndian.Uint32(tag.Buffer[0:]), binary.LittleEndian.Uint32(tag.Buffer[4:])
}

// SetPhysicalSize changes the display resolution (in pixels)
func SetPhysicalSize(width uint32, height uint32) error {
	buf := make([]byte, bcm2835.VC_FB_SET_PHYSICAL_SIZE_LEN)
	binary.LittleEndian.PutUint32(buf[0:], width)
	binary.LittleEndian.PutUint32(buf[4:], height)

	msg := &bcm2835.MailboxMessage{
		Tags: []bcm2835.MailboxTag{
			{
				ID:     bcm2835.VC_FB_SET_PHYSICAL_SIZE,
				Buffer: buf,
			},
		},
	}

	bcm2835.Mailbox.Call(bcm2835.VC_CH_PROPERTYTAGS_A_TO_VC, msg)

	tag := msg.Tag(bcm2835.VC_FB_SET_PHYSICAL_SIZE)

	if tag == nil || len(tag.Buffer) < 8 {
		return fmt.Errorf("failed to set desired size")
	}

	newWidth := binary.LittleEndian.Uint32(tag.Buffer[0:])
	newHeight := binary.LittleEndian.Uint32(tag.Buffer[4:])

	if newWidth != width || newHeight != height {
		return fmt.Errorf("failed to set desired size")
	}

	return nil
}
