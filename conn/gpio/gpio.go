// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpio defines digital pins.
//
// The GPIO pins are described in their logical functionality, not in their
// physical position.
package gpio

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"periph.io/x/periph/conn/pins"
)

// Interfaces

// Level is the level of the pin: Low or High.
type Level bool

const (
	// Low represents 0v.
	Low Level = false
	// High represents Vin, generally 3.3v or 5v.
	High Level = true
)

func (l Level) String() string {
	if l == Low {
		return "Low"
	}
	return "High"
}

// Pull specifies the internal pull-up or pull-down for a pin set as input.
type Pull uint8

// Acceptable pull values.
const (
	Float        Pull = 0 // Let the input float
	PullDown     Pull = 1 // Apply pull-down
	PullUp       Pull = 2 // Apply pull-up
	PullNoChange Pull = 3 // Do not change the previous pull resistor setting or an unknown value
)

const pullName = "FloatPullDownPullUpPullNoChange"

var pullIndex = [...]uint8{0, 5, 13, 19, 31}

func (i Pull) String() string {
	if i >= Pull(len(pullIndex)-1) {
		return fmt.Sprintf("Pull(%d)", i)
	}
	return pullName[pullIndex[i]:pullIndex[i+1]]
}

// Edge specifies if an input pin should have edge detection enabled.
//
// Only enable it when needed, since this causes system interrupts.
type Edge int

// Acceptable edge detection values.
const (
	NoEdge      Edge = 0
	RisingEdge  Edge = 1
	FallingEdge Edge = 2
	BothEdges   Edge = 3
)

const edgeName = "NoEdgeRisingEdgeFallingEdgeBothEdges"

var edgeIndex = [...]uint8{0, 6, 16, 27, 36}

func (i Edge) String() string {
	if i >= Edge(len(edgeIndex)-1) {
		return fmt.Sprintf("Edge(%d)", i)
	}
	return edgeName[edgeIndex[i]:edgeIndex[i+1]]
}

// PinIn is an input GPIO pin.
//
// It may optionally support internal pull resistor and edge based triggering.
type PinIn interface {
	pins.Pin
	// In setups a pin as an input.
	//
	// If WaitForEdge() is planned to be called, make sure to use one of the Edge
	// value. Otherwise, use None to not generated unneeded hardware interrupts.
	//
	// Calling In() will try to empty the accumulated edges but it cannot be 100%
	// reliable due to the OS (linux) and its driver. It is possible that on a
	// gpio that is as input, doing a quick Out(), In() may return an edge that
	// occurred before the Out() call.
	In(pull Pull, edge Edge) error
	// Read return the current pin level.
	//
	// Behavior is undefined if In() wasn't used before.
	//
	// In some rare case, it is possible that Read() fails silently. This happens
	// if another process on the host messes up with the pin after In() was
	// called. In this case, call In() again.
	Read() Level
	// WaitForEdge() waits for the next edge or immediately return if an edge
	// occurred since the last call.
	//
	// Only waits for the kind of edge as specified in a previous In() call.
	// Behavior is undefined if In() with a value other than None wasn't called
	// before.
	//
	// Returns true if an edge was detected during or before this call. Return
	// false if the timeout occurred or In() was called while waiting, causing the
	// function to exit.
	//
	// Multiple edges may or may not accumulate between two calls to
	// WaitForEdge(). The behavior in this case is undefined and is OS driver
	// specific.
	//
	// It is not required to call Read() to reset the edge detection.
	//
	// Specify -1 to effectively disable timeout.
	WaitForEdge(timeout time.Duration) bool
	// Pull returns the internal pull resistor if the pin is set as input pin.
	//
	// Returns PullNoChange if the value cannot be read.
	Pull() Pull
}

const (
	// Max is the PWM fully at high. One should use Out(High) instead.
	Max = 65536
	// Half is a 50% PWM duty cycle.
	Half = Max / 2
)

// PinOut is an output GPIO pin.
type PinOut interface {
	pins.Pin
	// Out sets a pin as output if it wasn't already and sets the initial value.
	//
	// After the initial call to ensure that the pin has been set as output, it
	// is generally safe to ignore the error returned.
	//
	// Out() tries to empty the accumulated edges detected if the gpio was
	// previously set as input but this is not 100% guaranteed due to the OS.
	Out(l Level) error
	// PWM sets a pin as output with a specified duty cycle between 0 and Max.
	//
	// The pin should use the highest frequency it can use.
	//
	// Use Half for a 50% duty cycle.
	PWM(duty int) error
}

// PinIO is a GPIO pin that supports both input and output.
//
// It may fail at either input and or output, for example ground, vcc and other
// similar pins.
type PinIO interface {
	pins.Pin
	In(pull Pull, edge Edge) error
	Read() Level
	WaitForEdge(timeout time.Duration) bool
	Pull() Pull
	Out(l Level) error
	PWM(duty int) error
}

// INVALID implements PinIO and fails on all access.
var INVALID PinIO

// RealPin is implemented by aliased pin and allows the retrieval of the real
// pin underlying an alias.
//
// The purpose of the RealPin is to be able to cleanly test whether an arbitrary
// gpio.PinIO returned by ByName is really an alias for another pin.
type RealPin interface {
	Real() PinIO // Real returns the real pin behind an Alias
}

// Registry

// ByNumber returns a GPIO pin from its number.
//
// Returns nil in case the pin is not present.
func ByNumber(number int) PinIO {
	mu.Lock()
	defer mu.Unlock()
	return getByNumber(number)
}

// ByName returns a GPIO pin from its name.
//
// This can be strings like GPIO2, PB8, etc.
//
// This function also parses string representation of numbers, so that calling
// with "6" will return the pin registered as number 6.
//
// Returns nil in case the pin is not present.
func ByName(name string) PinIO {
	mu.Lock()
	defer mu.Unlock()
	if p, ok := byName[0][name]; ok {
		return p
	}
	if p, ok := byName[1][name]; ok {
		return p
	}
	if p, ok := byAlias[name]; ok {
		if p.PinIO == nil {
			if p.PinIO = getByNumber(p.number); p.PinIO == nil {
				return nil
			}
		}
		return p
	}
	if i, err := strconv.Atoi(name); err == nil {
		return getByNumber(i)
	}
	return nil
}

// All returns all the GPIO pins available on this host.
//
// The list is guaranteed to be in order of number.
//
// This list excludes aliases.
//
// This list excludes non-GPIO pins like GROUND, V3_3, etc.
func All() []PinIO {
	mu.Lock()
	defer mu.Unlock()
	out := make(pinList, 0, len(byNumber))
	seen := make(map[int]struct{}, len(byNumber[0]))
	// Memory-mapped pins have highest priority, include all of them.
	for _, p := range byNumber[0] {
		out = append(out, p)
		seen[p.Number()] = struct{}{}
	}
	// Add in OS accessible pins that cannot be accessed via memory-map.
	for _, p := range byNumber[1] {
		if _, ok := seen[p.Number()]; !ok {
			out = append(out, p)
		}
	}
	sort.Sort(out)
	return out
}

// Aliases returns all pin aliases.
//
// The list is guaranteed to be in order of aliase name.
func Aliases() []PinIO {
	mu.Lock()
	defer mu.Unlock()
	out := make(pinList, 0, len(byAlias))
	for _, p := range byAlias {
		// Skip aliases that were not resolved.
		if p.PinIO == nil {
			if p.PinIO = getByNumber(p.number); p.PinIO == nil {
				continue
			}
		}
		out = append(out, p)
	}
	sort.Sort(out)
	return out
}

// Register registers a GPIO pin.
//
// Registering the same pin number or name twice is an error.
//
// `preferred` should be true when the pin being registered is exposing as much
// functionality as possible via the underlying hardware. This is normally done
// by accessing the CPU memory mapped registers directly.
//
// `preferred` should be false when the functionality is provided by the OS and
// is limited or slower.
//
// The pin registered cannot implement the interface RealPin.
func Register(p PinIO, preferred bool) error {
	name := p.Name()
	if len(name) == 0 {
		return errors.New("gpio: can't register a pin with no name")
	}
	if _, err := strconv.Atoi(name); err == nil {
		return fmt.Errorf("gpio: can't register a pin with a name being only a number %q", name)
	}
	number := p.Number()
	if number < 0 {
		return fmt.Errorf("gpio: can't register a pin with a negative number %d", number)
	}
	i := 0
	if !preferred {
		i = 1
	}

	mu.Lock()
	defer mu.Unlock()
	if orig, ok := byNumber[i][number]; ok {
		return fmt.Errorf("gpio: can't register the same pin %d twice; had %q, registering %q", number, orig, p)
	}
	if orig, ok := byName[i][name]; ok {
		return fmt.Errorf("gpio: can't register the same pin %q twice; had %q, registering %q", name, orig, p)
	}
	if r, ok := p.(RealPin); ok {
		return fmt.Errorf("gpio: can't register %q, which is an aliased for %q, use RegisterAlias() instead", p, r)
	}
	if alias, ok := byAlias[name]; ok {
		return fmt.Errorf("gpio: can't register %q for which an alias %q already exists", p, alias)
	}
	byNumber[i][number] = p
	byName[i][name] = p
	return nil
}

// RegisterAlias registers an alias for a GPIO pin.
//
// It is possible to register an alias for a pin number that itself has not
// been registered yet.
func RegisterAlias(alias string, number int) error {
	if len(alias) == 0 {
		return errors.New("gpio: can't register an alias with no name")
	}
	if _, err := strconv.Atoi(alias); err == nil {
		return fmt.Errorf("gpio: can't register an alias being only a number %q", alias)
	}
	if number < 0 {
		return fmt.Errorf("gpio: can't register an alias to a pin with a negative number %d", number)
	}

	mu.Lock()
	defer mu.Unlock()
	if orig := byAlias[alias]; orig != nil {
		return fmt.Errorf("gpio: can't register alias %q for pin %d: it is already aliased to %q", alias, number, orig)
	}
	byAlias[alias] = &pinAlias{name: alias, number: number}
	return nil
}

//

// errInvalidPin is returned when trying to use INVALID.
var errInvalidPin = errors.New("gpio: invalid pin")

var (
	mu sync.Mutex
	// The first map is preferred pins, the second is for more limited pins,
	// usually going through OS-provided abstraction layer.
	byNumber = [2]map[int]PinIO{{}, {}}
	byName   = [2]map[string]PinIO{{}, {}}
	byAlias  = map[string]*pinAlias{}
)

func init() {
	INVALID = invalidPin{}
}

// invalidPin implements PinIO for compatibility but fails on all access.
type invalidPin struct {
}

func (invalidPin) Number() int {
	return -1
}

func (invalidPin) Name() string {
	return "INVALID"
}

func (invalidPin) String() string {
	return "INVALID"
}

func (invalidPin) Function() string {
	return ""
}

func (invalidPin) In(Pull, Edge) error {
	return errInvalidPin
}

func (invalidPin) Read() Level {
	return Low
}

func (invalidPin) WaitForEdge(timeout time.Duration) bool {
	return false
}

func (invalidPin) Pull() Pull {
	return PullNoChange
}

func (invalidPin) Out(Level) error {
	return errInvalidPin
}

func (invalidPin) PWM(duty int) error {
	return errInvalidPin
}

// pinAlias implements an alias for a PinIO.
//
// pinAlias also implements the RealPin interface, which allows querying for
// the real pin under the alias.
type pinAlias struct {
	PinIO
	name   string
	number int
}

// String returns the alias name along the real pin's Name() in parenthesis, if
// known, else the real pin's number.
func (a *pinAlias) String() string {
	if a.PinIO == nil {
		return fmt.Sprintf("%s(%d)", a.name, a.number)
	}
	return fmt.Sprintf("%s(%s)", a.name, a.PinIO.Name())
}

// Name returns the pinAlias's name.
func (a *pinAlias) Name() string {
	return a.name
}

// Real returns the real pin behind the alias
func (a *pinAlias) Real() PinIO {
	return a.PinIO
}

func getByNumber(number int) PinIO {
	if p, ok := byNumber[0][number]; ok {
		return p
	}
	if p, ok := byNumber[1][number]; ok {
		return p
	}
	return nil
}

type pinList []PinIO

func (p pinList) Len() int           { return len(p) }
func (p pinList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pinList) Less(i, j int) bool { return p[i].Number() < p[j].Number() }

var _ PinIn = INVALID
var _ PinOut = INVALID
var _ PinIO = INVALID
