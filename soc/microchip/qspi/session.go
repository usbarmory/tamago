// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package qspi

// Session represents exclusive access to a QSPI controller, valid only within
// its [QSPI.Serialize] function.
type Session struct {
	hw *QSPI
}

// Serialize runs operation with exclusive access to the controller, to issue
// a sequence of commands (e.g. write enable, page program, status poll) which
// must not be interleaved. QSPI methods must not be called within operation.
func (hw *QSPI) Serialize(operation func(*Session) error) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if !hw.ready {
		return ErrNotInitialized
	}

	if operation == nil {
		return ErrInvalidOperation
	}

	s := &Session{hw: hw}
	defer func() { s.hw = nil }()

	return operation(s)
}

func validateSerialCommand(cmd Command, write bool) (err error) {
	// only single width is supported through registers
	if cmd.Width != Single || cmd.DummyCycles > 31 {
		return ErrInvalidCommand
	}

	switch cmd.AddressBits {
	case 0, 8, 16, 24, 32:
	default:
		return ErrInvalidCommand
	}

	// writes send the address as data, without dummy cycles
	if write && cmd.DummyCycles != 0 {
		return ErrInvalidCommand
	}

	return nil
}

func validateCommandAddress(cmd Command, address uint32) (err error) {
	if cmd.AddressBits == 0 {
		if address != 0 {
			return ErrCommandAddress
		}

		return nil
	}

	if cmd.AddressBits < 32 && address >= 1<<cmd.AddressBits {
		return ErrCommandAddress
	}

	return nil
}
