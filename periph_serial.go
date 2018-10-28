// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains the single threaded driver loading code, to be used on
// low performance cores.

// +build tinygo

package periph

import (
	"errors"
	"strconv"
)

func initImpl() (*State, error) {
	state = &State{}
	// At this point, byName is guaranteed to be immutable.
	stages, err := explodeStages()
	if err != nil {
		return state, err
	}
	loaded := make(map[string]struct{}, len(byName))
	for _, s := range stages {
		s.loadSerial(state, loaded)
	}
	return state, nil
}

// loadSerial loads all the drivers for this stage, one after the other.
func (s *stage) loadSerial(state *State, loaded map[string]struct{}) {
	for name, drv := range s.drvs {
		// Intentionally do not look at After(), only Prerequisites().
		for _, dep := range drv.Prerequisites() {
			if _, ok := loaded[dep]; !ok {
				state.Skipped = insertDriverFailure(state.Skipped, DriverFailure{drv, errors.New("dependency not loaded: " + strconv.Quote(dep))})
				goto loop
			}
		}

		// Not skipped driver, attempt loading in a goroutine.
		if ok, err := drv.Init(); ok {
			if err == nil {
				state.Loaded = insertDriver(state.Loaded, drv)
				loaded[name] = struct{}{}
			} else {
				state.Failed = insertDriverFailure(state.Failed, DriverFailure{drv, err})
			}
		} else {
			state.Skipped = insertDriverFailure(state.Skipped, DriverFailure{drv, err})
		}
	loop:
	}
}
