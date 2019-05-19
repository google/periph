// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

import (
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
)

// ServoGroup a group of servos connected to a pca9685 module
type ServoGroup struct {
	*Dev
	minPwm   gpio.Duty
	maxPwm   gpio.Duty
	minAngle physic.Angle
	maxAngle physic.Angle
}

// Servo individual servo from a group of servos connected to a pca9685 module
type Servo struct {
	group    *ServoGroup
	channel  int
	minAngle physic.Angle
	maxAngle physic.Angle
}

// NewServoGroup returns a servo group connected through the pca9685 module
// some pwm and angle limits can be set
func NewServoGroup(dev *Dev, minPwm, maxPwm gpio.Duty, minAngle, maxAngle physic.Angle) *ServoGroup {
	return &ServoGroup{
		Dev:      dev,
		minPwm:   minPwm,
		maxPwm:   maxPwm,
		minAngle: minAngle,
		maxAngle: maxAngle,
	}
}

// SetMinMaxPwm change pwm and angle limits
func (s *ServoGroup) SetMinMaxPwm(minAngle, maxAngle physic.Angle, minPwm, maxPwm gpio.Duty) {
	s.maxPwm = maxPwm
	s.minPwm = minPwm
	s.minAngle = minAngle
	s.maxAngle = maxAngle
}

// SetAngle set an angle in a given channel of the servo group
func (s *ServoGroup) SetAngle(channel int, angle physic.Angle) error {
	value := mapValue(int(angle), int(s.minAngle), int(s.maxAngle), int(s.minPwm), int(s.maxPwm))
	return s.Dev.SetPwm(channel, 0, gpio.Duty(value))
}

// GetServo returns a individual Servo to be controlled
func (s *ServoGroup) GetServo(channel int) *Servo {
	return &Servo{
		group:    s,
		channel:  channel,
		minAngle: s.minAngle,
		maxAngle: s.maxAngle,
	}
}

// SetMinMaxAngle change angle limits for the servo
func (s *Servo) SetMinMaxAngle(min, max physic.Angle) {
	s.minAngle = min
	s.maxAngle = max
}

// SetAngle set an angle on the servo
// will consider the angle limits set
func (s *Servo) SetAngle(angle physic.Angle) error {
	if angle < s.minAngle {
		angle = s.minAngle
	}
	if angle > s.maxAngle {
		angle = s.maxAngle
	}
	return s.group.SetAngle(s.channel, angle)
}

// SetPwm set an pmw value to the servo
func (s *Servo) SetPwm(pwm gpio.Duty) error {
	return s.group.SetPwm(s.channel, 0, pwm)
}

func mapValue(x, inMin, inMax, outMin, outMax int) int {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
