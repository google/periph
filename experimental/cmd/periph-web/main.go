// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// periph-web runs a web server exposing periph's state.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func mainImpl() error {
	port := flag.String("http", ":7080", "IP and port to bind to")
	verbose := flag.Bool("v", false, "verbose log")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unsupported arguments")
	}
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	state, err := hostInit()
	if err != nil {
		return err
	}
	s, err := newWebServer(*port, state)
	if err != nil {
		return err
	}
	c := make(chan os.Signal)
	go func() { <-c }()
	signal.Notify(c, os.Interrupt)
	<-c
	s.Close()
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "periph-web: %s.\n", err)
		os.Exit(1)
	}
}
