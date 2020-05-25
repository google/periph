// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate go get github.com/tdewolff/minify/cmd/minify
//go:generate go run internal/static_gen.go -o static_prod.go

package main

import "net/http"

func (s *webServer) addOtherHandlers() {
	http.HandleFunc("/raw/periph/v1/xsrf_token", noContent(s.apiXSRFTokenHandler))
	http.HandleFunc("/favicon.ico", getOnly(s.getFavicon))
	// Do not use getOnly here as it is the 'catch all, one and we want to check
	// that before the method.
	http.HandleFunc("/", noContent(s.getRoot))
}

// Static handlers

// /
func (s *webServer) getRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	s.setXSRFCookie(r.RemoteAddr, w)
	content := getContent("static/index.html")
	if content == nil {
		http.Error(w, "Content missing", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", cacheControlContent)
	_, _ = w.Write(content)
}

// /favicon.ico
func (s *webServer) getFavicon(w http.ResponseWriter, r *http.Request) {
	content := getContent("static/favicon.ico")
	if content == nil {
		http.Error(w, "Content missing", 500)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", cacheControlContent)
	_, _ = w.Write(content)
}

// Other handlers

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
	_, _ = w.Write([]byte(t))
}
