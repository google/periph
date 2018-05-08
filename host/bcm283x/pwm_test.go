// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import "testing"

func TestPWMMap(t *testing.T) {
	p := pwmMap{}
	p.reset()
	if _, err := setPWMClockSource(); err == nil {
		t.Fatal("pwmMemory is nil")
	}
	defer func() {
		drvDMA.pwmMemory = nil
	}()
	drvDMA.pwmMemory = &p
	if _, err := setPWMClockSource(); err == nil {
		t.Fatal("clockMemory is nil")
	}
}
