// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioreg

import (
	"fmt"
	"log"
	"sort"
	"testing"

	"periph.io/x/periph/conn/gpio"
)

func ExampleAll() {
	fmt.Print("GPIO pins available:\n")
	for _, pin := range All() {
		fmt.Printf("- %s: %s\n", pin, pin.Function())
	}
}

func ExampleByName() {
	p := ByName("GPIO6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}
	fmt.Printf("%s: %s\n", p, p.Function())
}

func ExampleByName_alias() {
	p := ByName("LCD-D2")
	if p == nil {
		log.Fatal("Failed to find LCD-D2")
	}
	if rp, ok := p.(gpio.RealPin); ok {
		fmt.Printf("%s is an alias for %s\n", p, rp.Real())
	} else {
		fmt.Printf("%s is not an alias!\n", p)
	}
}

func ExampleByName_number() {
	// The string representation of a number works too.
	p := ByName("6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}
	fmt.Printf("%s: %s\n", p, p.Function())
}

func ExampleByNumber() {
	p := ByNumber(6)
	if p == nil {
		log.Fatal("Failed to find #6")
	}
	fmt.Printf("%s: %s\n", p, p.Function())
}

func TestRegister(t *testing.T) {
	defer reset()
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "a"}, false); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 1 {
		t.Fatalf("Expected one pin, got %v", a)
	}
	if a := Aliases(); len(a) != 0 {
		t.Fatalf("Expected zero alias, got %v", a)
	}
	if ByName("a") == nil {
		t.Fail()
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "a"}, true); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 1 {
		t.Fatalf("Expected one pin, got %v", a)
	}
	if a := Aliases(); len(a) != 0 {
		t.Fatalf("Expected zero alias, got %v", a)
	}
	if ByNumber(0) == nil {
		t.Fail()
	}
	if ByNumber(1) != nil {
		t.Fail()
	}
	if ByName("a") == nil {
		t.Fail()
	}
	if ByName("0") == nil {
		t.Fail()
	}
	if ByName("1") != nil {
		t.Fail()
	}
	if ByName("b") != nil {
		t.Fail()
	}
}

func TestRegister_fail(t *testing.T) {
	defer reset()
	if err := Register(&basicPin{PinIO: gpio.INVALID}, false); err == nil {
		t.Fatal("Expected error")
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "a", num: -1}, false); err == nil {
		t.Fatal("Expected error")
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "1"}, false); err == nil {
		t.Fatal("Expected error")
	}
}

func TestRegisterAlias(t *testing.T) {
	defer reset()
	if err := RegisterAlias("alias0", 0); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAlias("alias0", 0); err == nil {
		t.Fatal(err)
	}
	if p := ByName("alias0"); p != nil {
		t.Fatalf("unexpected alias0: %v", p)
	}
	if a := All(); len(a) != 0 {
		t.Fatalf("Expected zero pin, got %v", a)
	}
	if a := Aliases(); len(a) != 0 {
		t.Fatalf("Expected zero alias, got %v", a)
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "GPIO0"}, false); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 1 {
		t.Fatalf("Expected one pin, got %v", a)
	}
	if a := Aliases(); len(a) != 1 {
		t.Fatalf("Expected one alias, got %v", a)
	}
	if p := ByName("alias0"); p == nil {
		t.Fail()
	} else if r := p.(gpio.RealPin).Real(); r.Name() != "GPIO0" {
		t.Fatalf("Expected real GPIO0, got %v", r)
	} else if s := p.String(); s != "alias0(GPIO0)" {
		t.Fatal(s)
	}

	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "GPIO1"}, false); err == nil {
		t.Fail()
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "GPIO0", num: 1}, false); err == nil {
		t.Fail()
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, N: "alias0", num: 1}, false); err == nil {
		t.Fail()
	}
	if err := Register(&pinAlias{&basicPin{PinIO: gpio.INVALID, N: "GPIO1", num: 1}, "alias1", 2}, false); err == nil {
		t.Fail()
	}
}

func TestRegisterAlias_fail(t *testing.T) {
	defer reset()
	if err := RegisterAlias("", 1); err == nil {
		t.Fatal("Expected error")
	}
	if err := RegisterAlias("alias0", -1); err == nil {
		t.Fatal("Expected error")
	}
	if err := RegisterAlias("0", 0); err == nil {
		t.Fatal("Expected error")
	}
}

func TestPinList(t *testing.T) {
	l := pinList{&basicPin{PinIO: gpio.INVALID, num: 1}, &basicPin{PinIO: gpio.INVALID}}
	sort.Sort(l)
	if l[0].(*basicPin).num != 0 || l[1].(*basicPin).num != 1 {
		t.Fatal(l)
	}
}

//

// basicPin implements Pin as a non-functional pin.
type basicPin struct {
	gpio.PinIO
	N   string
	num int
}

func (b *basicPin) String() string {
	return b.N
}

func (b *basicPin) Name() string {
	return b.N
}

func (b *basicPin) Number() int {
	return b.num
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byNumber = [2]map[int]gpio.PinIO{{}, {}}
	byName = [2]map[string]gpio.PinIO{{}, {}}
	byAlias = map[string]*pinAlias{}
}
