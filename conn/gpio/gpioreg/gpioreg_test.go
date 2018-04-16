// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioreg

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
)

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
	if err := RegisterAlias("alias0", "GPIO0"); err != nil {
		t.Fatal(err)
	}
	if RegisterAlias("alias0", "GPIO1") == nil {
		t.Fatal("can't register an alias to a different gpio")
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
	if Register(&pinAlias{PinIO: &basicPin{PinIO: gpio.INVALID, name: "GPIO1", num: 1}, name: "alias1"}, false) == nil {
		t.Fatal("can't register a pin implementing RealPin")
	}

	if ByName("0") == nil {
		t.Fatal("getByName for low priority pin")
	}
}

func TestRegisterAlias_chain(t *testing.T) {
	defer reset()
	if err := RegisterAlias("a0", "a1"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAlias("a1", "a2"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAlias("a2", "GPIO0"); err != nil {
		t.Fatal(err)
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "GPIO0", num: 0}, false); err != nil {
		t.Fatal(err)
	}
	p := ByName("a0")
	if p == nil {
		t.Fatal("ByName(\"a0\") didn't find pin")
	}
	if s := p.String(); s != "a0(GPIO0)" {
		t.Fatalf("unexpected pin name: %q", s)
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

func TestUnRegister(t *testing.T) {
	defer reset()
	if err := RegisterAlias("Alias", "GPIO0"); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("Alias"); err != nil {
		t.Fatal(err)
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "GPIO0", num: 0}, false); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("GPIO0"); err != nil {
		t.Fatal(err)
	}
	if err := Register(&basicPin{PinIO: gpio.INVALID, name: "GPIO0", num: 0}, true); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("GPIO0"); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 0 {
		t.Fatalf("Expected no pin, got %v", a)
	}
	if err := Unregister("Unknown"); err == nil {
		t.Fatal("Can't unregister unknown pin")
	}
}

func TestInsertPinByNumber(t *testing.T) {
	out := insertPinByNumber(nil, &basicPin{name: "b", num: 1})
	out = insertPinByNumber(out, &basicPin{name: "d", num: 3})
	out = insertPinByNumber(out, &basicPin{name: "c", num: 2})
	out = insertPinByNumber(out, &basicPin{name: "a", num: 0})
	for i := 0; i < 4; i++ {
		if out[i].Number() != i {
			t.Fatal(out)
		}
	}
}

func TestInsertPinByName(t *testing.T) {
	out := insertPinByName(nil, &basicPin{name: "b"})
	out = insertPinByName(out, &basicPin{name: "d"})
	out = insertPinByName(out, &basicPin{name: "c"})
	out = insertPinByName(out, &basicPin{name: "a"})
	for i, l := range []string{"a", "b", "c", "d"} {
		if out[i].Name() != l {
			t.Fatal(out)
		}
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
	byAlias = map[string]string{}
}
