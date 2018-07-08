// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
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

// registerAPIs registers the JSON API handlers.
func (s *webServer) registerAPIs() {
	http.HandleFunc("/api/periph/v1/gpio/aliases", s.api(s.apiGPIOAliases))
	http.HandleFunc("/api/periph/v1/gpio/list", s.api(s.apiGPIOList))
	http.HandleFunc("/api/periph/v1/gpio/read", s.api(s.apiGPIORead))
	http.HandleFunc("/api/periph/v1/gpio/out", s.api(s.apiGPIOOut))
	http.HandleFunc("/api/periph/v1/header/list", s.api(s.apiHeaderList))
	http.HandleFunc("/api/periph/v1/i2c/list", s.api(s.apiI2CList))
	http.HandleFunc("/api/periph/v1/spi/list", s.api(s.apiSPIList))
	http.HandleFunc("/api/periph/v1/server/state", s.api(s.apiServerState))
	// Not JSON but here since it can be used by the same client:
	http.HandleFunc("/raw/periph/v1/xsrf_token", noContent(s.apiXSRFTokenHandler))
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

func (s *webServer) apiGPIOAliases() ([]pinAlias, int) {
	all := gpioreg.Aliases()
	out := make([]pinAlias, 0, len(all))
	for _, p := range all {
		r := p.(gpio.RealPin).Real()
		out = append(out, pinAlias{p.Name(), r.Name()})
	}
	return out, 200
}

// /api/periph/v1/gpio/list

func (s *webServer) apiGPIOList() ([]gpioPin, int) {
	all := gpioreg.All()
	out := make([]gpioPin, 0, len(all))
	for _, p := range all {
		out = append(out, toPin(p))
	}
	return out, 200
}

// /api/periph/v1/gpio/read

func (s *webServer) apiGPIORead(in []string) ([]int, int) {
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

func (s *webServer) apiGPIOOut(in map[string]bool) ([]string, int) {
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

func (s *webServer) apiHeaderList() (map[string]header, int) {
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

func (s *webServer) apiI2CList() ([]i2cRef, int) {
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

func (s *webServer) apiSPIList() ([]spiRef, int) {
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
	State       state
	PeriphExtra bool
}

func (s *webServer) apiServerState() (*serverStateOut, int) {
	out := &serverStateOut{
		Hostname:    s.hostname,
		State:       s.state,
		PeriphExtra: periphExtra,
	}
	return out, 200
}

// /raw/periph/v1/xsrf_token

func (s *webServer) apiXSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}
	t := s.setXSRFCookie(r.RemoteAddr, w)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", cacheControlNone)
	w.WriteHeader(200)
	w.Write([]byte(t))
}
