// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Inspired by:
// https://github.com/maruel/serve-dir/blob/master/loghttp/loghttp.go

package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received := time.Now()
		rwh := responseWriteHijacker{responseWriter: responseWriter{ResponseWriter: w}}
		w = &rwh
		// Not all ResponseWriter implement Hijack, so query its support upfront.
		if _, ok := w.(http.Hijacker); !ok {
			w = &rwh.responseWriter
		}
		defer func() {
			defaultOnDone(r, received, rwh.status, rwh.length, rwh.hijacked)
		}()
		h.ServeHTTP(w, r)
	})
}

// Private.

func defaultOnDone(r *http.Request, received time.Time, status, length int, hijacked bool) {
	m := r.Method
	if hijacked {
		m = "HIJACKED"
	}
	log.Printf("%s - %3d %6db %-4s %6s %s", r.RemoteAddr, status, length, m, roundDuration(time.Since(received), 3), r.RequestURI)
}

// roundDuration returns time rounded to 'precision' digits.
func roundDuration(t time.Duration, precision int) time.Duration {
	// Find the highest digit in base 10.
	// TODO(maruel): Optimize.
	l := len(strconv.FormatInt(int64(t), 10))
	if l < precision {
		return t
	}
	// TODO(maruel): Optimize.
	r := time.Duration(1)
	for i := precision; i < l; i++ {
		r *= 10
	}
	return (t + r/2) / r * r
}

type responseWriter struct {
	http.ResponseWriter
	length int
	status int
}

func (r *responseWriter) Write(data []byte) (size int, err error) {
	if r.status == 0 {
		r.status = 200
	}
	size, err = r.ResponseWriter.Write(data)
	r.length += size
	return
}

func (r *responseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.status = status
}

type responseWriteHijacker struct {
	responseWriter
	hijacked bool
}

// Hijack is needed for websocket.
//
// TODO(maruel): For now the write length is lost, it would require querying
// the ReadWriter and overriding Conn.Write().
func (r *responseWriteHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.hijacked = true
	return r.ResponseWriter.(http.Hijacker).Hijack()
}
