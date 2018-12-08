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

func fromBitString(t *testing.T, s string) (res byte) {
	d, err := strconv.ParseUint(s, 2, 8)
	if err != nil {
		t.Fatal(err)
	}
	res = byte(d & 0xFF)
	return
}

/*
   C1 C2 C3
3 : 1  0  0
2 : 0  0  1
1 : 1  0  1
0 : 1  1  0

1 1 1 0 0 1 0 0
1 0 1 1 1 0 0 1
0 1 1 0 0 0 0 1
*/
func TestBitCalc(t *testing.T) {

	ba := BlocksAccess{
		B3: KeyA_RN_WB_BITS_RAB_WN_KeyB_RN_WB, // 100
		B2: RAB_WN_IN_DAB,                     // 001
		B1: RB_WN_IN_DN,                       // 101
		B0: RAB_WB_IB_DAB,                     // 110
	}

	var access [4]byte

	if err := ba.serialize(access[:1]); err == nil {
		t.Fatal("destination array should be reported as insufficient")
	}

	if err := ba.serialize(access[:]); err != nil {
		t.Fatal(err)
	}

	if fromBitString(t, "1011") != ba.getBits(1) {
		t.Fatalf("1011 is not equal to %d", ba.getBits(1))
	}

	expected := []byte{
		fromBitString(t, "11100100"),
		fromBitString(t, "10111001"),
		fromBitString(t, "01100001"),
		0,
	}

	expected[3] = expected[0] ^ expected[1] ^ expected[2]

	if !bytes.Equal(expected, access[:]) {
		t.Fatalf("Access is incorrect: %v != %v", expected, access)
	}

	var parsedAccess BlocksAccess

	parsedAccess.Init(access[:])

	if !reflect.DeepEqual(ba, parsedAccess) {
		t.Fatalf("Parsed access mismatch %s != %s", ba.String(), (parsedAccess).String())
	}
}

/*
   C1 C2 C3
3 : 0  0  1
2 : 0  0  0
1 : 0  0  0
0 : 0  0  0


1 1 1 1 1 1 1 1
0 0 0 0 0 1 1 1
1 0 0 0 0 0 0 0
*/
func TestByteArrayDecipher(t *testing.T) {
	bitsData := [...]byte{
		fromBitString(t, "11111111"),
		fromBitString(t, "00000111"),
		fromBitString(t, "10000000"),
		fromBitString(t, "01111000"),
	}
	ba := BlocksAccess{
		B0: AnyKeyRWID,
		B1: AnyKeyRWID,
		B2: AnyKeyRWID,
		B3: KeyA_RN_WA_BITS_RA_WA_KeyB_RA_WA,
	}
	var access [4]byte

	if err := ba.serialize(access[:]); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(bitsData[:], access[:]) {
		t.Fatalf("Wrong access calculation: %v != %v", bitsData, access)
	}
}
