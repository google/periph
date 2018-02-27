// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mfrc522

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
)

func TestBitCalc(t *testing.T) {

	ba := BlocksAccess{
		B3: KeyA_RN_WB_BITS_RAB_WN_KeyB_RN_WB, // 100
		B2: RAB_WN_IN_DAB,                     // 001
		B1: RB_WN_IN_DN,                       // 101
		B0: RAB_WB_IB_DAB,                     // 110
	}

	access := CalculateBlockAccess(&ba)

	reader := func(s string) (res byte) {
		d, err := strconv.ParseUint(s, 2, 8)
		if err != nil {
			t.Fatal(err)
		}
		res = byte(d & 0xFF)
		return
	}

	if reader("0110") != ba.getBits(1) {
		t.Fatalf("0110 is not equal to %d", ba.getBits(1))
	}

	expected := []byte{
		reader("11101001"),
		reader("01100100"),
		reader("10110001"),
		0,
	}

	expected[3] = expected[0] ^ expected[1] ^ expected[2]

	if !bytes.Equal(expected, access) {
		t.Fatal("Access is incorrect")
	}

	parsedAccess := ParseBlockAccess(access)

	if !reflect.DeepEqual(ba, *parsedAccess) {
		t.Fatal("Parsed access mismatch")
	}
}
