// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build debug

package main

import (
	"io/ioutil"
	"log"
)

const cacheControlContent = "Cache-Control:no-cache,private"

func getContent(path string) []byte {
	r, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("getContent(%q): %v", path, err)
	}
	return r
}
