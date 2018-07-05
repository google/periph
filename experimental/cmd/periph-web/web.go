// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"periph.io/x/periph"
)

const cacheControl30d = "Cache-Control:public,max-age=259200" // 30d
const cacheControl5m = "Cache-Control:public,max-age=300"     // 5m
const cacheControlNone = "Cache-Control:no-cache,private"

type driverFailure struct {
	D   string
	Err string
}

// Similar to periph.State but is JSON marshalable as-is.
type state struct {
	Loaded  []string
	Skipped []driverFailure
	Failed  []driverFailure
}

func (s *state) init(st *periph.State) {
	s.Loaded = make([]string, len(st.Loaded))
	for i, v := range st.Loaded {
		s.Loaded[i] = v.String()
	}
	s.Skipped = make([]driverFailure, len(st.Skipped))
	for i, v := range st.Skipped {
		s.Skipped[i].D = v.D.String()
		s.Skipped[i].Err = v.Err.Error()
	}
	s.Failed = make([]driverFailure, len(st.Failed))
	for i, v := range st.Failed {
		s.Failed[i].D = v.D.String()
		s.Failed[i].Err = v.Err.Error()
	}
}

type webServer struct {
	ln       net.Listener
	server   http.Server
	state    state
	hostname string
}

func newWebServer(port string, state *periph.State) (*webServer, error) {
	s := &webServer{}
	s.state.init(state)
	log.Printf("%#v", s.state)
	var err error
	if s.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}
	s.registerAPIs()
	http.HandleFunc("/favicon.ico", getOnly(s.getFavicon))
	http.HandleFunc("/", getOnly(s.getRoot))
	if s.ln, err = net.Listen("tcp", port); err != nil {
		return nil, err
	}
	s.server = http.Server{
		Addr:    s.ln.Addr().String(),
		Handler: http.DefaultServeMux,
	}
	go s.server.Serve(s.ln)
	return s, nil
}

func (s *webServer) Close() error {
	return s.ln.Close()
}

// Static handlers.

func (s *webServer) getRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", cacheControl5m)
	w.Write(rootPage)
}

func (s *webServer) getFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", cacheControl30d)
	w.Write(favicon)
}

//

// getOnly returns an http.Handler that refuses other verbs than GET or HEAD.
func getOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}
