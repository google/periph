// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	key      [8]byte
}

func getHostAndPort(hostport string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, fmt.Errorf("could not split http address: %v", err)
	}
	if host == "" {
		host = "localhost"
	}
	var port int
	if portStr != "" {
		if port, err = strconv.Atoi(portStr); err != nil {
			return "", 0, fmt.Errorf("invalid port number: %v", err)
		}
	}
	return host, port, nil
}

func isLocalhost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "[::1]", "::1":
		return true
	}
	return false
}

func newWebServer(hostport string, state *periph.State, verbose bool) (*webServer, error) {
	s := &webServer{server: http.Server{Handler: http.DefaultServeMux}}
	if _, err := rand.Read(s.key[:]); err != nil {
		return nil, err
	}
	s.state.init(state)
	var err error
	host, port, err := getHostAndPort(hostport)
	if err != nil {
		return nil, err
	}

	s.registerAPIs()
	http.HandleFunc("/favicon.ico", getOnly(s.getFavicon))
	http.HandleFunc("/", getOnly(s.getRoot))
	// We love middlewares!
	if isLocalhost(host) {
		s.hostname = "localhost"
		s.server.Handler = localOnly(s.server.Handler)
	} else {
		if s.hostname, err = os.Hostname(); err != nil {
			return nil, err
		}
	}
	if verbose {
		s.server.Handler = loggingHandler(s.server.Handler)
	}

	if s.ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port)); err != nil {
		return nil, err
	}
	s.server.Addr = s.ln.Addr().String()
	go s.server.Serve(s.ln)
	return s, nil
}

func (s *webServer) Close() error {
	return s.ln.Close()
}

// Inspired by https://github.com/golang/net/blob/master/xsrftoken/xsrf.go

func (s *webServer) generateToken(userID string, now time.Time) string {
	milliTime := (now.UnixNano() + 1e6 - 1) / 1e6
	h := hmac.New(sha1.New, s.key[:])
	fmt.Fprintf(h, "%d:%s", milliTime, userID)
	return fmt.Sprintf("%x:%s", milliTime, strings.TrimRight(base64.URLEncoding.EncodeToString(h.Sum(nil)), "="))
}

func (s *webServer) validateToken(token string, userID string) bool {
	now := time.Now()
	sep := strings.Index(token, ":")
	if sep < 0 {
		return false
	}
	millis, err := strconv.ParseInt(token[:sep], 16, 64)
	if err != nil {
		return false
	}
	issueTime := time.Unix(0, millis*1e6)
	if now.Sub(issueTime) >= 24*time.Hour || issueTime.After(now.Add(1*time.Minute)) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.generateToken(userID, issueTime))) == 1
}

func (s *webServer) setXSRFCookie(addr string, w http.ResponseWriter) string {
	t := s.generateToken(strings.SplitN(addr, ":", 2)[0], time.Now())
	c := http.Cookie{
		Name:   "XSRF-TOKEN",
		Value:  t,
		MaxAge: 23 * 60 * 60,
	}
	http.SetCookie(w, &c)
	return t
}

// enforceXSRF is an handler wrapper that enforces the XSRF token.
func (s *webServer) enforceXSRF(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Cookie("XSRF-TOKEN")
		if c == nil {
			log.Printf("Missing XSRF-TOKEN cookie")
			http.Error(w, "Missing XSRF-TOKEN cookie", 400)
			r.Body.Close()
			return
		}
		if !s.validateToken(c.Value, strings.SplitN(r.RemoteAddr, ":", 2)[0]) {
			log.Printf("Invalid XSRF-TOKEN cookie %q", c.Value)
			http.Error(w, "Invalid XSRF-TOKEN cookie", 400)
			r.Body.Close()
			return
		}
		h(w, r)
	}
}

// Static handlers.

func (s *webServer) getRoot(w http.ResponseWriter, r *http.Request) {
	r.Body.Close()
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	s.setXSRFCookie(r.RemoteAddr, w)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", cacheControl5m)
	w.Write(rootPage)
}

func (s *webServer) getFavicon(w http.ResponseWriter, r *http.Request) {
	r.Body.Close()
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
			r.Body.Close()
			return
		}
		h(w, r)
	}
}

func localOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil || !isLocalhost(host) {
			http.Error(w, "permission denied", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
