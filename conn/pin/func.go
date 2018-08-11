// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pin

import (
	"strconv"
	"strings"
)

// Func is a pin function.
//
// The Func format must be "[A-Z]+", "[A-Z]+_[A-Z]+" or exceptionally
// "(In|Out)/(Low|High)".
type Func string

// FuncNone is returned by PinFunc.Func() for a Pin without an active
// functionality.
const FuncNone Func = ""

// Specialize converts a "BUS_LINE" function and appends the bug number and
// line number, to look like "BUS0_LINE1".
//
// Use -1 to not add a bus or line number.
func (f Func) Specialize(b, l int) Func {
	if f == FuncNone {
		return FuncNone
	}
	if b != -1 {
		parts := strings.SplitN(string(f), "_", 2)
		if len(parts) == 1 {
			return FuncNone
		}
		f = Func(parts[0] + strconv.Itoa(b) + "_" + parts[1])
	}
	if l != -1 {
		f += Func(strconv.Itoa(l))
	}
	return f
}

// Generalize is the reverse of Specialize().
func (f Func) Generalize() Func {
	parts := strings.SplitN(string(f), "_", 2)
	f = Func(strings.TrimRightFunc(parts[0], isNum))
	if len(parts) == 2 {
		f += "_"
		f += Func(strings.TrimRightFunc(parts[1], isNum))
	}
	return f
}

//

func isNum(r rune) bool {
	return r >= '0' && r <= '9'
}
