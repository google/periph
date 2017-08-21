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

func TestRegister(t *testing.T) {
	defer reset()
	// Low priority pin.
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "a", num: 0}, false); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 1 {
		t.Fatalf("Expected one pin, got %v", a)
	}
	if a := Aliases(); len(a) != 0 {
		t.Fatalf("Expected zero alias, got %v", a)
	}
	if ByName("a") == nil {
		t.Fatal("failed to get pin 'a'")
	}
	// High priority pin.
	if Register(&basicPin{PinIO: gpio.INVALID, name: "a", num: 2}, true) == nil {
		t.Fatal("same name, different numbers")
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "a", num: 0}, true); err != nil {
		t.Fatal(err)
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "b", num: 0}, true); err == nil {
		t.Fatalf("#0 already registered as a")
	}
	if a := All(); len(a) != 1 {
		t.Fatalf("Expected one pin, got %v", a)
	}
	if a := Aliases(); len(a) != 0 {
		t.Fatalf("Expected zero alias, got %v", a)
	}
	if ByName("a") == nil {
		t.Fatal("failed to get pin 'a'")
	}
	if ByName("0") == nil {
		t.Fatal("failed to get pin #0")
	}
	if ByName("1") != nil {
		t.Fatal("there is no get pin #1")
	}
	if ByName("b") != nil {
		t.Fatal("there is no get pin 'b'")
	}
}

func TestRegister_fail(t *testing.T) {
	defer reset()
	if err := Register(&basicPin{PinIO: gpio.INVALID}, false); err == nil {
		t.Fatal("pin with no name")
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "a", num: -1}, false); err == nil {
		t.Fatal("Expected error")
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "1", num: 0}, false); err == nil {
		t.Fatal("pin with name is a number")
	}
}

func TestRegisterAlias(t *testing.T) {
	defer reset()
	if err := RegisterAlias("alias0", "GPIO0"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAlias("alias0", "GPIO0"); err == nil {
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
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "GPIO0", num: 0}, false); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 1 {
		t.Fatalf("Expected one pin, got %v", a)
	}
	if a := Aliases(); len(a) != 1 {
		t.Fatalf("Expected one alias, got %v", a)
	}
	if p := ByName("alias0"); p == nil {
		t.Fatal("alias0 doesn't resolve to a registered pin")
	} else if r := p.(gpio.RealPin).Real(); r.Name() != "GPIO0" {
		t.Fatalf("Expected real GPIO0, got %v", r)
	} else if s := p.String(); s != "alias0(GPIO0)" {
		t.Fatal(s)
	}

	if Register(&basicPin{PinIO: gpio.INVALID, name: "GPIO1", num: 0}, false) == nil {
		t.Fatal("pin #0 is already registered")
	}
	if Register(&basicPin{PinIO: gpio.INVALID, name: "GPIO0", num: 1}, false) == nil {
		t.Fatal("GPIO0 is already registered")
	}
	if Register(&basicPin{PinIO: gpio.INVALID, name: "alias0", num: 1}, false) == nil {
		t.Fatal("alias0 is already registered as an alias")
	}
	if Register(&pinAlias{PinIO: &basicPin{PinIO: gpio.INVALID, name: "GPIO1", num: 1}, name: "alias1", dest: "GPIO2"}, false) == nil {
		t.Fatal("can't register a pin implementing RealPin")
	}

	if ByName("0") == nil {
		t.Fatal("getByName for low priority pin")
	}
}

func TestRegisterAlias_fail(t *testing.T) {
	defer reset()
	if err := RegisterAlias("", "Dest"); err == nil {
		t.Fatal("alias with no name")
	}
	if err := RegisterAlias("alias", ""); err == nil {
		t.Fatal("dest with no name")
	}
	if err := RegisterAlias("0", "dest"); err == nil {
		t.Fatal("alias is a number")
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
	name string
	num  int
}

func (b *basicPin) String() string {
	return b.name
}

func (b *basicPin) Name() string {
	return b.name
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
