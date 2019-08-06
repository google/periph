// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package environment_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"periph.io/x/periph/conn/environment"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host"
)

type fakeSensor struct {
}

func (f *fakeSensor) String() string {
	return "fake"
}

func (f *fakeSensor) Halt() error {
	return nil
}

func (f *fakeSensor) SenseWeather(w *environment.Weather) error {
	w.Temperature = physic.ZeroCelsius + 23*physic.Celsius
	w.Pressure = 101200 * physic.Pascal
	w.Humidity = 540 * physic.MilliRH
	return nil
}

func (f *fakeSensor) SenseWeatherContinuous(ctx context.Context, interval time.Duration, c chan<- environment.WeatherSample) {
	c <- environment.WeatherSample{
		T: time.Now(),
		Weather: environment.Weather{
			Temperature: physic.ZeroCelsius + 23*physic.Celsius,
			Pressure:    101200 * physic.Pascal,
			Humidity:    540 * physic.MilliRH,
		},
	}
	c <- environment.WeatherSample{
		T: time.Now(),
		Weather: environment.Weather{
			Temperature: physic.ZeroCelsius + 24*physic.Celsius,
			Pressure:    101400 * physic.Pascal,
			Humidity:    530 * physic.MilliRH,
		},
	}
}

func (f *fakeSensor) PrecisionWeather(w *environment.Weather) {
	w.Temperature = 1 * physic.Celsius
	w.Pressure = 100 * physic.Pascal
	w.Humidity = 10 * physic.MilliRH
}

func ExampleWeather() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Get an handle to a weather sensing device.
	var s environment.SenseWeather = &fakeSensor{}

	// Print out the sensor precision.
	var w environment.Weather
	s.PrecisionWeather(&w)
	fmt.Printf("Sensor precision:\n")
	fmt.Printf("  Temperature: %gK\n", float64(w.Temperature)/float64(physic.Kelvin))
	fmt.Printf("  Pressure:    %v\n", w.Pressure)
	fmt.Printf("  Humidty:     %v\n", w.Humidity)
	fmt.Printf("\n")

	// Take one measure.
	if err := s.SenseWeather(&w); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sensor measurement:\n")
	fmt.Printf("  Temperature: %v\n", w.Temperature)
	fmt.Printf("  Pressure:    %v\n", w.Pressure)
	fmt.Printf("  Humidty:     %v\n", w.Humidity)
	// Output:
	// Sensor precision:
	//   Temperature: 1K
	//   Pressure:    100Pa
	//   Humidty:     1%rH
	//
	// Sensor measurement:
	//   Temperature: 23°C
	//   Pressure:    101.200kPa
	//   Humidty:     54%rH
}

func ExampleWeatherSample() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Get an handle to a weather sensing device.
	var s environment.SenseWeather = &fakeSensor{}

	// Measure at most for 2s.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cw := make(chan environment.WeatherSample)
	go func() {
		s.SenseWeatherContinuous(ctx, time.Second, cw)
		close(cw)
	}()

	// Terminate upon receiving an interrupt signal, like a Ctrl-C at the
	// terminal prompt.
	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)
	go func() {
		<-cs
		cancel()
	}()

	for {
		select {
		case w, ok := <-cw:
			if !ok {
				return
			}
			if w.Err != nil {
				log.Fatal(w.Err)
			}
			fmt.Printf("Sensor measurement:\n")
			fmt.Printf("  Temperature: %v\n", w.Temperature)
			fmt.Printf("  Pressure:    %v\n", w.Pressure)
			fmt.Printf("  Humidty:     %v\n", w.Humidity)
		case <-ctx.Done():
			// Wait for the goroutine to return.
			_, _ = <-cw
			return
		}
	}
	// Output:
	// Sensor measurement:
	//   Temperature: 23°C
	//   Pressure:    101.200kPa
	//   Humidty:     54%rH
	// Sensor measurement:
	//   Temperature: 24°C
	//   Pressure:    101.400kPa
	//   Humidty:     53%rH
}
