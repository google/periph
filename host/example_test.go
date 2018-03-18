// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/host"
)

func ExampleInit() {
	// Make sure periph is initialized.
	state, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	// Prints the loaded driver.
	fmt.Printf("Using drivers:\n")
	for _, driver := range state.Loaded {
		fmt.Printf("- %s\n", driver)
	}

	// Prints the driver that were skipped as irrelevant on the platform.
	fmt.Printf("Drivers skipped:\n")
	for _, failure := range state.Skipped {
		fmt.Printf("- %s: %s\n", failure.D, failure.Err)
	}

	// Having drivers failing to load may not require process termination. It
	// is possible to continue to run in partial failure mode.
	fmt.Printf("Drivers failed to load:\n")
	for _, failure := range state.Failed {
		fmt.Printf("- %s: %v\n", failure.D, failure.Err)
	}

	// Use pins, buses, devices, etc.
}
