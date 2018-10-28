// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

// ServoGroup a group of servos connected to a pca9685 module
type ServoGroup struct {
	*Dev
	minPwm   int
	maxPwm   int
	minAngle int
	maxAngle int
}

// Servo individual servo from a group of servos connected to a pca9685 module
type Servo struct {
	group    *ServoGroup
	channel  int
	minAngle int
	maxAngle int
}

// NewServoGroup returns a servo group connected throught the pca9685 module
// some pwm and angle limits can be set
func NewServoGroup(dev *Dev, minPwm, maxPwm, minAngle, maxAngle int) *ServoGroup {
	return &ServoGroup{
		Dev:      dev,
		minPwm:   minPwm,
		maxPwm:   maxPwm,
		minAngle: minAngle,
		maxAngle: maxAngle,
	}
}

// SetMinMaxPwm change pwm and angle limits
func (s *ServoGroup) SetMinMaxPwm(minAngle, maxAngle, minPwm, maxPwm int) {
	s.maxPwm = maxPwm
	s.minPwm = minPwm
	s.minAngle = minAngle
	s.maxAngle = maxAngle
}

// SetAngle set an angle in a given channel of the servo group
func (s *ServoGroup) SetAngle(channel, angle int) {
	value := mapValue(angle, s.minAngle, s.maxAngle, s.minPwm, s.maxPwm)
	s.Dev.SetPwm(channel, 0, uint16(value))
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
func (s *Servo) SetMinMaxAngle(min, max int) {
	s.minAngle = min
	s.maxAngle = max
}

// SetAngle set an angle on the servo
// will consider the angle limits set
func (s *Servo) SetAngle(angle int) {
	if angle < s.minAngle {
		angle = s.minAngle
	}
	if angle > s.maxAngle {
		angle = s.maxAngle
	}
	s.group.SetAngle(s.channel, angle)
}

// SetPwm set an pmw value to the servo
func (s *Servo) SetPwm(pwm uint16) {
	s.group.SetPwm(s.channel, 0, pwm)
}

func mapValue(x, inMin, inMax, outMin, outMax int) int {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
