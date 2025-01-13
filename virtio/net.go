// VirtIO network card driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package virtio

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"sync"
)

// Net represents a VirtIO network device instance.
type Net struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32

	// MAC address
	MAC net.HardwareAddr

	// VirtIO instance
	io *VirtIO
}

// Init initializes the VirtIO network device.
func (hw *Net) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.io = &VirtIO{
		Base: hw.Base,
	}

	if hw.MAC == nil {
		hw.MAC = make([]byte, 6)
		rand.Read(hw.MAC)
		// flag address as unicast and locally administered
		hw.MAC[0] &= 0xfe
		hw.MAC[0] |= 0x02
	} else if len(hw.MAC) != 6 {
		return errors.New("invalid MAC")
	}

	if err := hw.io.Init(); err != nil {
		return err
	}

	if id := hw.io.DeviceID(); id != NetworkCard {
		return fmt.Errorf("incompatible device ID (%x)", id)
	}

	hw.io.SelectQueue(0)

	return
}
