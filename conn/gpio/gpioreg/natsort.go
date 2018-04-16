// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioreg

import (
	"strconv"
)

// lessNatural does a 'natural' comparison on the two strings.
//
// It is extracted from https://github.com/maruel/natural.
func lessNatural(a, b string) bool {
	for {
		if a == b {
			return false
		}
		if p := commonPrefix(a, b); p != 0 {
			a = a[p:]
			b = b[p:]
		}
		if ia := digits(a); ia > 0 {
			if ib := digits(b); ib > 0 {
				// Both sides have digits.
				an, aerr := strconv.ParseUint(a[:ia], 10, 64)
				bn, berr := strconv.ParseUint(b[:ib], 10, 64)
				if aerr == nil && berr == nil {
					if an != bn {
						return an < bn
					}
					// Semantically the same digits, e.g. "00" == "0", "01" == "1". In
					// this case, only continue processing if there's trailing data on
					// both sides, otherwise do lexical comparison.
					if ia != len(a) && ib != len(b) {
						a = a[ia:]
						b = b[ib:]
						continue
					}
				}
			}
		}
		return a < b
	}
}

// commonPrefix returns the common prefix except for digits.
func commonPrefix(a, b string) int {
	m := len(a)
	if n := len(b); n < m {
		m = n
	}
	if m == 0 {
		return 0
	}
	_ = a[m-1]
	_ = b[m-1]
	for i := 0; i < m; i++ {
		ca := a[i]
		cb := b[i]
		if (ca >= '0' && ca <= '9') || (cb >= '0' && cb <= '9') || ca != cb {
			return i
		}
	}
	return m
}

func digits(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return i
		}
	}
	return len(s)
}
