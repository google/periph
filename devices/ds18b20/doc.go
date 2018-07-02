// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ds18b20 interfaces to Dallas Semi / Maxim DS18B20 and MAX31820
// 1-wire temperature sensors.
//
// Note that both DS18B20 and MAX31820 use family code 0x28.
//
// Both powered sensors and parasitically powered sensors are supported
// as long as the bus driver can provide sufficient power using an active
// pull-up.
//
// The DS18B20 alarm functionality and reading/writing the 2 alarm bytes in
// the EEPROM are not supported. The DS18S20 is also not supported.
//
// More details
//
// See https://periph.io/device/ds18b20/ for more details about the device.
//
// Datasheets
//
// https://datasheets.maximintegrated.com/en/ds/DS18B20-PAR.pdf
//
// http://datasheets.maximintegrated.com/en/ds/MAX31820.pdf
package ds18b20
