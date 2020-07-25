// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mfrc522_test

import (
	"encoding/hex"
	"log"
	"reflect"
	"time"

	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Using SPI as an example. See package "periph.io/x/periph/conn/spi/spireg" for more details.
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	rfid, err := mfrc522.NewSPI(p, rpi.P1_13, rpi.P1_11)
	if err != nil {
		log.Fatal(err)
	}

	// Idling device on exit.
	defer rfid.Halt()

	// Setting the antenna signal strength.
	rfid.SetAntennaGain(5)

	// Converting access key.
	// This value corresponds to first pi "numbers": 3 14 15 92 65 35.
	hexKey, _ := hex.DecodeString("030e0f5c4123")
	var key [6]byte
	copy(key[:], hexKey)

	// Converting expected data.
	// This value corresponds to string "@~>f=Um[X{LRwA3}".
	expected, _ := hex.DecodeString("407e3e663d556d5b587b4c527741337d")

	timedOut := false
	cb := make(chan []byte)
	timer := time.NewTimer(10 * time.Second)

	// Stopping timer, flagging reader thread as timed out
	defer func() {
		timer.Stop()
		timedOut = true
		close(cb)
	}()

	go func() {
		log.Printf("Started %s", rfid.String())

		for {
			// Trying to read data from sector 1 block 0
			data, err := rfid.ReadCard(10*time.Second, byte(commands.PICC_AUTHENT1B), 1, 0, key)

			// If main thread timed out just exiting.
			if timedOut {
				return
			}

			// Some devices tend to send wrong data while RFID chip is already detected
			// but still "too far" from a receiver.
			// Especially some cheap CN clones which you can find on GearBest, AliExpress, etc.
			// This will suppress such errors.
			if err != nil {
				continue
			}

			cb <- data
		}
	}()

	for {
		select {
		case <-timer.C:
			log.Fatal("Didn't receive device data")
			return
		case data := <-cb:
			if !reflect.DeepEqual(data, expected) {
				log.Fatal("Received data is incorrect")
			} else {
				log.Println("Received data is correct")
			}

			return
		}
	}
}

func ExampleDev_ReadUID() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Using SPI as an example. See package "periph.io/x/periph/conn/spi/spireg" for more details.
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	rfid, err := mfrc522.NewSPI(p, rpi.P1_22, rpi.P1_18)
	if err != nil {
		log.Fatal(err)
	}

	// Idling device on exit.
	defer rfid.Halt()

	// Setting the antenna signal strength.
	rfid.SetAntennaGain(5)

	timedOut := false
	cb := make(chan []byte)
	timer := time.NewTimer(10 * time.Second)

	// Stopping timer, flagging reader thread as timed out
	defer func() {
		timer.Stop()
		timedOut = true
		close(cb)
	}()

	go func() {
		log.Printf("Started %s", rfid.String())

		for {
			// Trying to read card UID.
			uid, err := rfid.ReadUID(10 * time.Second)

			// If main thread timed out just exiting.
			if timedOut {
				return
			}

			// Some devices tend to send wrong data while RFID chip is already detected
			// but still "too far" from a receiver.
			// Especially some cheap CN clones which you can find on GearBest, AliExpress, etc.
			// This will suppress such errors.
			if err != nil {
				continue
			}

			cb <- uid
			return
		}
	}()

	for {
		select {
		case <-timer.C:
			log.Fatal("Didn't receive device data")
			return
		case data := <-cb:
			log.Println("UID:", hex.EncodeToString(data))
			return
		}
	}
}
