// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
)

// jsonAPI contains the global state/caches for the JSON API.
type jsonAPI struct {
	hostname string
	state    drvState
}

// getAPIs returns the JSON API handlers.
func (j *jsonAPI) getAPIs() []apiHandler {
	return []apiHandler{
		{"/api/periph/v1/gpio/aliases", j.apiGPIOAliases},
		{"/api/periph/v1/gpio/in", j.apiGPIOIn},
		{"/api/periph/v1/gpio/list", j.apiGPIOList},
		{"/api/periph/v1/gpio/read", j.apiGPIORead},
		{"/api/periph/v1/gpio/out", j.apiGPIOOut},
		{"/api/periph/v1/header/list", j.apiHeaderList},
		{"/api/periph/v1/i2c/list", j.apiI2CList},
		{"/api/periph/v1/spi/list", j.apiSPIList},
		{"/api/periph/v1/server/state", j.apiServerState},
	}
}

//

type gpioPin struct {
	// Immutable.
	Name string
	Num  int
	// Mutable, the GPIO can change function over time.
	Func string
}

func toPin(p pin.Pin) gpioPin {
	return gpioPin{p.Name(), p.Number(), p.Function()}
}

// /api/periph/v1/gpio/aliases

type pinAlias struct {
	Name string
	Dest string
}

func (j *jsonAPI) apiGPIOAliases() ([]pinAlias, int) {
	all := gpioreg.Aliases()
	out := make([]pinAlias, 0, len(all))
	for _, p := range all {
		r := p.(gpio.RealPin).Real()
		out = append(out, pinAlias{p.Name(), r.Name()})
	}
	return out, 200
}

// /api/periph/v1/gpio/in

type pinIn struct {
	Name string
	Pull string
	Edge string
}

func (j *jsonAPI) apiGPIOIn(in []pinIn) ([]string, int) {
	out := make([]string, 0, len(in))
	for _, l := range in {
		if p := gpioreg.ByName(l.Name); p != nil {
			pull := gpio.PullNoChange
			switch l.Pull {
			case "down":
				pull = gpio.PullDown
			case "float":
				pull = gpio.Float
			case "up":
				pull = gpio.PullUp
			}
			edge := gpio.NoEdge
			switch l.Edge {
			case "both":
				edge = gpio.BothEdges
			case "falling":
				edge = gpio.FallingEdge
			case "rising":
				edge = gpio.RisingEdge
			}
			if err := p.In(pull, edge); err != nil {
				out = append(out, err.Error())
			} else {
				out = append(out, "")
			}
		} else {
			out = append(out, "Pin not found")
		}
	}
	return out, 200
}

// /api/periph/v1/gpio/list

func (j *jsonAPI) apiGPIOList() ([]gpioPin, int) {
	all := gpioreg.All()
	out := make([]gpioPin, 0, len(all))
	for _, p := range all {
		out = append(out, toPin(p))
	}
	return out, 200
}

// /api/periph/v1/gpio/read

func (j *jsonAPI) apiGPIORead(in []string) ([]int, int) {
	out := make([]int, 0, len(in))
	for _, name := range in {
		v := -1
		if p := gpioreg.ByName(name); p != nil {
			if v = 0; p.Read() {
				v = 1
			}
		}
		out = append(out, v)
	}
	return out, 200
}

// /api/periph/v1/gpio/out

func (j *jsonAPI) apiGPIOOut(in map[string]bool) ([]string, int) {
	out := make([]string, 0, len(in))
	for name, l := range in {
		if p := gpioreg.ByName(name); p != nil {
			if err := p.Out(gpio.Level(l)); err != nil {
				out = append(out, err.Error())
			} else {
				out = append(out, "")
			}
		} else {
			out = append(out, "Pin not found")
		}
	}
	return out, 200
}

// /api/periph/v1/header/list

type header struct {
	Pins [][]gpioPin
}

func (j *jsonAPI) apiHeaderList() (map[string]header, int) {
	hdrs := pinreg.All()
	out := make(map[string]header, len(hdrs))
	for name, pins := range hdrs {
		h := header{make([][]gpioPin, len(pins))}
		for i, r := range pins {
			row := make([]gpioPin, 0, len(r))
			for _, p := range r {
				row = append(row, toPin(p))
			}
			h.Pins[i] = row
		}
		out[name] = h
	}
	return out, 200
}

// /api/periph/v1/i2c/list

type i2cRef struct {
	Name    string
	Aliases []string
	Number  int
	Err     string
	SCL     string
	SDA     string
}

func (j *jsonAPI) apiI2CList() ([]i2cRef, int) {
	buses := i2creg.All()
	out := make([]i2cRef, 0, len(buses))
	for _, ref := range buses {
		h := i2cRef{Name: ref.Name, Aliases: ref.Aliases, Number: ref.Number}
		if bus, err := ref.Open(); bus != nil {
			if p, ok := bus.(i2c.Pins); ok {
				h.SCL = p.SCL().Name()
				h.SDA = p.SDA().Name()
			}
			bus.Close()
		} else {
			h.Err = err.Error()
		}
		out = append(out, h)
	}
	return out, 200
}

// /api/spi/list

type spiRef struct {
	Name    string
	Aliases []string
	Number  int
	Err     string
	CLK     string
	MOSI    string
	MISO    string
	CS      string
}

func (j *jsonAPI) apiSPIList() ([]spiRef, int) {
	buses := spireg.All()
	out := make([]spiRef, 0, len(buses))
	for _, ref := range buses {
		h := spiRef{Name: ref.Name, Aliases: ref.Aliases, Number: ref.Number}
		if bus, err := ref.Open(); bus != nil {
			if p, ok := bus.(spi.Pins); ok {
				h.CLK = p.CLK().Name()
				h.MOSI = p.MOSI().Name()
				h.MISO = p.MISO().Name()
				h.CS = p.CS().Name()
			}
			bus.Close()
		} else {
			h.Err = err.Error()
		}
		out = append(out, h)
	}
	return out, 200
}

// /api/periph/v1/server/state

type serverStateOut struct {
	Hostname    string
	State       drvState
	PeriphExtra bool
}

type driverFailure struct {
	D   string
	Err string
}

// Similar to periph.State but is JSON marshalable as-is.
type drvState struct {
	Loaded  []string
	Skipped []driverFailure
	Failed  []driverFailure
}

func (j *jsonAPI) apiServerState() (*serverStateOut, int) {
	out := &serverStateOut{
		Hostname:    j.hostname,
		State:       j.state,
		PeriphExtra: periphExtra,
	}
	return out, 200
}
