// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
)

// registerAPIs registers the API handlers.
func (s *webServer) registerAPIs() {
	http.HandleFunc("/api/gpio/_aliases", api(s.apiGPIOAliases))
	http.HandleFunc("/api/gpio/_all", api(s.apiGPIOAll))
	http.HandleFunc("/api/header/_all", api(s.apiHeaderAll))
	http.HandleFunc("/api/i2c/_all", api(s.apiI2CAll))
	http.HandleFunc("/api/spi/_all", api(s.apiSPIAll))
	http.HandleFunc("/api/server/state", api(s.apiServerState))
}

// api is a simple JSON api wrapper.
func api(h func() (interface{}, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		data, code := h()
		raw, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", cacheControlNone)
		w.WriteHeader(code)
		w.Write(raw)
	}
}

// /api/gpio/_aliases
type gpioAliases []pinAlias

type pinAlias struct {
	Name string
	Dest string
}

func (s *webServer) apiGPIOAliases() (interface{}, int) {
	all := gpioreg.Aliases()
	data := make(gpioAliases, 0, len(all))
	for _, p := range all {
		r := p.(gpio.RealPin).Real()
		data = append(data, pinAlias{p.Name(), r.Name()})
	}
	return data, 200
}

// /api/gpio/_all
type gpioAll []gpioPin

type gpioPin struct {
	Name string
	Num  int
	Func string
}

func toPin(p pin.Pin) gpioPin {
	return gpioPin{p.Name(), p.Number(), p.Function()}
}

func (s *webServer) apiGPIOAll() (interface{}, int) {
	all := gpioreg.All()
	data := make(gpioAll, 0, len(all))
	for _, p := range all {
		data = append(data, toPin(p))
	}
	return data, 200
}

// /api/header/_all
type headerAll map[string]header

type header struct {
	Pins [][]gpioPin
}

func (s *webServer) apiHeaderAll() (interface{}, int) {
	hdrs := pinreg.All()
	data := make(headerAll, len(hdrs))
	for name, pins := range hdrs {
		h := header{make([][]gpioPin, len(pins))}
		for i, r := range pins {
			row := make([]gpioPin, 0, len(r))
			for _, p := range r {
				row = append(row, toPin(p))
			}
			h.Pins[i] = row
		}
		data[name] = h
	}
	return data, 200
}

// /api/i2c/_all
type i2cAll []i2cRef

type i2cRef struct {
	Name    string
	Aliases []string
	Number  int
	Err     string
	SCL     string
	SDA     string
}

func (s *webServer) apiI2CAll() (interface{}, int) {
	buses := i2creg.All()
	data := make(i2cAll, 0, len(buses))
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
		data = append(data, h)
	}
	return data, 200
}

// /api/spi/_all
type spiAll []spiRef

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

func (s *webServer) apiSPIAll() (interface{}, int) {
	buses := spireg.All()
	data := make(spiAll, 0, len(buses))
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
		data = append(data, h)
	}
	return data, 200
}

// /api/server/state
type serverState struct {
	Hostname    string
	State       state
	PeriphExtra bool
}

func (s *webServer) apiServerState() (interface{}, int) {
	data := serverState{
		Hostname:    s.hostname,
		State:       s.state,
		PeriphExtra: periphExtra,
	}
	return data, 200
}
