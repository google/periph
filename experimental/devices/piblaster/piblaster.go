// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package piblaster implements interfacing code is piblaster.
//
// See https://github.com/sarfata/pi-blaster for more details. This package
// relies on pi-blaster being installed and enabled.
//
// TODO(maruel): "dtoverlay=pwm" or "dtoverlay=pwm-2chan" works without having
// to install anything, albeit with less pins supported.
//
// Warning
//
// piblaster doesn't report what pins is controls so it is easy to misuse this
// library.
package piblaster

import (
	"fmt"
	"io"
	"os"

	"github.com/google/pio/conn/gpio"
)

// SetPWM enables and sets the PWM duty on a GPIO output pin via piblaster.
//
// duty must be [0, 1].
func SetPWM(p gpio.PinIO, duty float32) error {
	if duty < 0 || duty > 1 {
		return fmt.Errorf("duty %f is invalid for blaster", duty)
	}
	err := openPiblaster()
	if err == nil {
		_, err = io.WriteString(piblasterHandle, fmt.Sprintf("%d=%f\n", p.Number(), duty))
	}
	return err
}

// ReleasePWM releases a GPIO output and leave it floating.
//
// This function must be called on process exit for each activated pin
// otherwise the pin will stay in the state.
func ReleasePWM(p gpio.PinIO) error {
	err := openPiblaster()
	if err == nil {
		_, err = io.WriteString(piblasterHandle, fmt.Sprintf("release %d\n", p.Number()))
	}
	return err
}

//

var piblasterHandle io.WriteCloser

func openPiblaster() error {
	if piblasterHandle == nil {
		f, err := os.OpenFile("/dev/pi-blaster", os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		piblasterHandle = f
	}
	return nil
}

func closePiblaster() error {
	if piblasterHandle != nil {
		w := piblasterHandle
		piblasterHandle = nil
		return w.Close()
	}
	return nil
}
