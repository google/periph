// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package environment

import (
	"context"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/physic"
)

// Weather represents measurements from an environmental sensor.
type Weather struct {
	Temperature physic.Temperature
	Pressure    physic.Pressure
	Humidity    physic.RelativeHumidity
}

// WeatherSample is a sample that occurred at the specified moment.
//
// It is used by SenseWeatherContinuous.
type WeatherSample struct {
	Weather
	// T is the moment at which the sensing was initiated.
	T time.Time
	// Err is set if sensing failed. In this case it can be assumed that
	// SenseWeatherContinuous() is aborting.
	Err error
}

// SenseWeather represents an environmental weather sensor.
type SenseWeather interface {
	conn.Resource

	// SenseWeather returns the value read from the sensor.
	//
	// Metrics in Weather unsupported by the sensor are not modified.
	SenseWeather(w *Weather) error
	// SenseWeatherContinuous sense continuously at the specified interval until
	// the context is canceled.
	//
	// If the context passed in is already canceled, no measurement is done and
	// nothing is sent to the channel.
	//
	// One measurement is done immediately upon call. The channel must be valid.
	// It is up to the client to decide if the channel is buffered or not.
	//
	// In case of operation failure, sends an error with Err set and exits.
	SenseWeatherContinuous(ctx context.Context, interval time.Duration, c chan<- WeatherSample)
	// PrecisionWeather returns this sensor's precision.
	//
	// The w values are set to the number of bits that are significant for each
	// items that this sensor can measure.
	//
	// Precision is not accuracy. The sensor may have absolute and relative
	// errors in its measurement, that are likely well above the reported
	// precision. Accuracy may be improved on some sensor by using oversampling,
	// or doing oversampling in software. Refer to the sensor datasheet if
	// available.
	PrecisionWeather(w *Weather)
}
