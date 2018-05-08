// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"testing"
	"time"
)

func TestReadTime(t *testing.T) {
	if ReadTime() != 0 {
		t.Fatal("timerMemory is nil")
	}

	defer func() {
		drvDMA.timerMemory = nil
	}()
	drvDMA.timerMemory = &timerMap{low: 1}
	if d := ReadTime(); d != time.Microsecond {
		t.Fatal(d)
	}
}
