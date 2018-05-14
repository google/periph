// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"time"

	"periph.io/x/periph/conn"
)

// Env represents measurements from an environmental sensor.
type Env struct {
	Temperature Temperature
	Pressure    Pressure
	Humidity    RelativeHumidity
}

// SenseEnv represents an environmental sensor.
type SenseEnv interface {
	conn.Resource

	// Sense returns the value read from the sensor. Unsupported metrics are not
	// modified.
	Sense(env *Env) error
	// SenseContinuous initiates a continuous sensing at the specified interval.
	//
	// It is important to call Halt() once done with the sensing, which will turn
	// the device off and will close the channel.
	SenseContinuous(interval time.Duration) (<-chan Env, error)
}
