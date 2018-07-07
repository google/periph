// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	http.HandleFunc("/api/periph/v1/gpio/aliases", s.api(s.apiGPIOAliases))
	http.HandleFunc("/api/periph/v1/gpio/list", s.api(s.apiGPIOList))
	http.HandleFunc("/api/periph/v1/gpio/read", s.api(s.apiGPIORead))
	http.HandleFunc("/api/periph/v1/gpio/out", s.api(s.apiGPIOOut))
	http.HandleFunc("/api/periph/v1/header/list", s.api(s.apiHeaderList))
	http.HandleFunc("/api/periph/v1/i2c/list", s.api(s.apiI2CList))
	http.HandleFunc("/api/periph/v1/spi/list", s.api(s.apiSPIList))
	http.HandleFunc("/api/periph/v1/server/state", s.api(s.apiServerState))
	// Not JSON:
	http.HandleFunc("/raw/periph/v1/xsrf_token", s.apiXSRFTokenHandler)
}

// api is a simple JSON api wrapper.
func (s *webServer) api(h func(b []byte) (interface{}, int)) http.HandlerFunc {
	return s.enforceXSRF(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "POST" {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		data, code := h(b)
		raw, err := json.Marshal(data)
		if err != nil {
			log.Printf("Malformed response: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", cacheControlNone)
		w.WriteHeader(code)
		w.Write(raw)
	})
}

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
type gpioAliasesOut []pinAlias

type pinAlias struct {
	Name string
	Dest string
}

func (s *webServer) apiGPIOAliases(b []byte) (interface{}, int) {
	all := gpioreg.Aliases()
	data := make(gpioAliasesOut, 0, len(all))
	for _, p := range all {
		r := p.(gpio.RealPin).Real()
		data = append(data, pinAlias{p.Name(), r.Name()})
	}
	return data, 200
}

// /api/periph/v1/gpio/list
type gpioListOut []gpioPin

func (s *webServer) apiGPIOList(b []byte) (interface{}, int) {
	all := gpioreg.All()
	data := make(gpioListOut, 0, len(all))
	for _, p := range all {
		data = append(data, toPin(p))
	}
	return data, 200
}

// /api/periph/v1/gpio/read
type gpioReadIn []string
type gpioReadOut []int

func (s *webServer) apiGPIORead(b []byte) (interface{}, int) {
	var in gpioReadIn
	if err := json.Unmarshal(b, &in); err != nil {
		log.Printf("Malformed user data: %v", err)
		return map[string]string{"error": err.Error()}, 400
	}
	out := make(gpioReadOut, 0, len(in))
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
type gpioOutIn map[string]bool
type gpioOutOut []string

func (s *webServer) apiGPIOOut(b []byte) (interface{}, int) {
	var in gpioOutIn
	if err := json.Unmarshal(b, &in); err != nil {
		log.Printf("Malformed user data: %v", err)
		return map[string]string{"error": err.Error()}, 400
	}
	out := make(gpioOutOut, 0, len(in))
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
type headerListOut map[string]header

type header struct {
	Pins [][]gpioPin
}

func (s *webServer) apiHeaderList(b []byte) (interface{}, int) {
	hdrs := pinreg.All()
	data := make(headerListOut, len(hdrs))
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

// /api/periph/v1/i2c/list
type i2cListOut []i2cRef

type i2cRef struct {
	Name    string
	Aliases []string
	Number  int
	Err     string
	SCL     string
	SDA     string
}

func (s *webServer) apiI2CList(b []byte) (interface{}, int) {
	buses := i2creg.All()
	data := make(i2cListOut, 0, len(buses))
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

// /api/spi/list
type spiListOut []spiRef

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

func (s *webServer) apiSPIList(b []byte) (interface{}, int) {
	buses := spireg.All()
	data := make(spiListOut, 0, len(buses))
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

// /api/periph/v1/server/state
type serverStateOut struct {
	Hostname    string
	State       state
	PeriphExtra bool
}

func (s *webServer) apiServerState(b []byte) (interface{}, int) {
	data := serverStateOut{
		Hostname:    s.hostname,
		State:       s.state,
		PeriphExtra: periphExtra,
	}
	return data, 200
}

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
