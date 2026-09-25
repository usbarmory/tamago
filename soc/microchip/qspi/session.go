// Microchip Quad SPI (QSPI) controller driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package qspi

// Session provides exclusive serial-memory access to a QSPI controller. A
// session is valid only for the duration of its [QSPI.Serialize] callback;
// later use returns ErrSessionClosed.
type Session struct {
	hw *QSPI
}

// Serialize runs operation with exclusive access to an initialized controller.
// Every serial-memory command goes through the Session passed to operation, so
// the callback can group related commands, such as write enable, page program,
// status polling, and write disable, without interleaving access from another
// goroutine. The callback must not call QSPI methods, which acquire the same
// lock. A controller transfer failure invalidates the controller; later
// commands require QSPI.Init outside the callback.
func (hw *QSPI) Serialize(operation func(*Session) error) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if !hw.ready {
		return ErrNotInitialized
	}

	if operation == nil {
		return ErrInvalidOperation
	}

	session := &Session{hw: hw}
	defer func() { session.hw = nil }()

	return operation(session)
}

func validateSerialCommand(command Command, write bool) (err error) {
	// Single is the only command-mode layout qualified by this driver.
	if command.Width != Single || command.DummyCycles > 31 {
		return ErrInvalidCommand
	}

	switch command.AddressBits {
	case 0, 8, 16, 24, 32:
	default:
		return ErrInvalidCommand
	}

	// Command-mode writes emit the address through the APB data path. Dummy
	// cycles are therefore unsupported for writes.
	if write && command.DummyCycles != 0 {
		return ErrInvalidCommand
	}

	return nil
}

func validateCommandAddress(command Command, address uint32) (err error) {
	if command.AddressBits == 0 {
		if address != 0 {
			return ErrCommandAddress
		}

		return nil
	}

	if command.AddressBits < 32 && address >= 1<<command.AddressBits {
		return ErrCommandAddress
	}

	return nil
}
