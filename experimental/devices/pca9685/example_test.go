// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.
package pca9685_test

import (
	"log"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/pca9685"

	"periph.io/x/periph/host"
)

func Example() {
	_, err := host.Init()
	if err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}

	pca, err := pca9685.NewI2C(bus)
	if err != nil {
		log.Fatal(err)
	}

	pca.SetPwmFreq(50)
	pca.SetAllPwm(0, 0)
	servos := pca9685.NewServoGroup(pca, 50, 650, 0, 180)

	// This is an example of using with an Me Arm robot arm
	gripServo := servos.GetServo(0)
	baseServo := servos.GetServo(1)
	elbowServo := servos.GetServo(2)
	shoulderServo := servos.GetServo(3)

	gripServo.SetMinMaxAngle(15, 120)
	elbowServo.SetMinMaxAngle(50, 110)    // Set limit of the robot arm
	shoulderServo.SetMinMaxAngle(60, 140) // Set limit of the robot arm

	// Set all in the middle in a MeArm robot arm
	gripServo.SetAngle(90)
	baseServo.SetAngle(90)
	elbowServo.SetAngle(90)
	shoulderServo.SetAngle(90)
}
