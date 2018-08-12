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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"periph.io/x/periph"
)

const cacheControlNone = "Cache-Control:no-cache,private"

type apiHandler struct {
	path string
	fn   interface{}
}

type webServer struct {
	ln     net.Listener
	server http.Server
	apis   jsonAPI
	key    [8]byte
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
	var err error
	host, port, err := getHostAndPort(hostport)
	if err != nil {
		return nil, err
	}
	hostname := ""
	if isLocalhost(host) {
		hostname = "localhost"
		s.server.Handler = localOnly(s.server.Handler)
	} else {
		if hostname, err = os.Hostname(); err != nil {
			return nil, err
		}
	}

	// Setup handlers.
	s.apis.init(hostname, state)
	for _, h := range s.apis.getAPIs() {
		http.HandleFunc(h.path, s.api(h.fn))
	}
	s.addOtherHandlers()
	if verbose {
		s.server.Handler = loggingHandler(s.server.Handler)
	}

	// Start serving.
	if s.ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port)); err != nil {
		return nil, err
	}
	s.server.Addr = s.ln.Addr().String()
	go func() {
		_ = s.server.Serve(s.ln)
	}()
	return s, nil
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

// http.Handler/HandlerFunc decorators.

// localOnly disallow remote access.
//
// It must be the front line decorator.
func localOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil || !isLocalhost(host) {
			http.Error(w, "permission denied", http.StatusForbidden)
			r.Body.Close()
			return
		}
		h.ServeHTTP(w, r)
	})
}

// enforceXSRF is an handler wrapper that enforces the XSRF token.
//
// In practice it's only used for the APIs within the api() decorator.
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

// getOnly returns an http.Handler that refuses other verbs than GET or HEAD.
//
// Also uses noContent().
func getOnly(h http.HandlerFunc) http.HandlerFunc {
	return noContent(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	})
}

// noContent ensure no content is posted. It closes r.Body.
func noContent(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
		if n != 0 {
			http.Error(w, "Unexpected content", 400)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		h(w, r)
	})
}

// api wraps a JSON api handler.
func (s *webServer) api(h interface{}) http.HandlerFunc {
	v := reflect.ValueOf(h)
	t := v.Type()
	if t.Kind() != reflect.Func {
		panic("send API func")
	}
	var inT reflect.Type
	if nArg := t.NumIn(); nArg == 1 {
		inT = t.In(0)
	} else if nArg != 0 {
		panic("pass func that accepts zero or one arg")
	}
	if t.NumOut() != 2 {
		panic("pass func that returns two args")
	}
	return s.enforceXSRF(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "POST" {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.RawQuery != "" {
			http.Error(w, "Do not use query argment", 400)
			return
		}
		// Ignore suffix "; charset=utf-8" for now.
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Content-Type must be application/json", 400)
			return
		}
		d := json.NewDecoder(r.Body)
		// d.DisallowUnknownFields() is only available in Go 1.10+.
		callDisallowUnknownFields(d)
		var in []reflect.Value
		if inT != nil {
			inv := reflect.New(inT)
			if err := d.Decode(inv.Interface()); err != nil {
				http.Error(w, fmt.Sprintf("Malformed user data: %v", err), 400)
			}
			in = append(in, inv.Elem())
		} else {
			var m map[string]string
			if err := d.Decode(&m); err != nil {
				http.Error(w, fmt.Sprintf("Malformed user data: %v", err), 400)
			}
			if len(m) != 0 {
				http.Error(w, "Unexpected data", 400)
			}
		}
		out := v.Call(in)
		raw, err := json.Marshal(out[0].Interface())
		if err != nil {
			http.Error(w, fmt.Sprintf("Malformed response: %v", err), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", cacheControlNone)
		w.WriteHeader(int(out[1].Int()))
		_, _ = w.Write(raw)
	})
}
