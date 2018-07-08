// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build go1.10

package main

import "encoding/json"

func callDisallowUnknownFields(d *json.Decoder) {
	d.DisallowUnknownFields()
}
