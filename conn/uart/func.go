// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package uart

import "periph.io/x/periph/conn/pin"

const (
	RX  pin.Func = "UART_RX"  // Receive
	TX  pin.Func = "UART_TX"  // Transmit
	RTS pin.Func = "UART_RTS" // Request to send
	CTS pin.Func = "UART_CTS" // Clear to send

	// These are rarely used.
	DTR pin.Func = "UART_DTR" // Data terminal ready
	DSR pin.Func = "UART_DSR" // Data set ready
	DCD pin.Func = "UART_DCD" // Data carrier detect
	RI  pin.Func = "UART_RI"  // Ring indicator
)
