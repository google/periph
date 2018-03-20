// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package hd44780 controls the Hitachi LCD display chipset HD-44780
//
// Datasheet
//
// https://www.sparkfun.com/datasheets/LCD/HD44780.pdf
package hd44780

// LCD interface to interact with the controller.
type LCD interface {
	Init()

	Cls()

	Print(data string)

	WriteChar(data uint8)

	SetCursor(line uint8, column uint8)
}
