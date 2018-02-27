// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"periph.io/x/periph/conn/spi/spireg"
	rf522 "periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host"
	"strconv"
	"strings"
)

var (
	sector = flag.Int("sector", 1, "Sector to access")
	block  = flag.Int("block", 0, "Block to access")

	rsPin  = flag.String("rs-pin", "13", "Reset pin")
	irqPin = flag.String("irq-pin", "12", "IRQ pin")

	keyCommand   = flag.Bool("wa", false, "Overwrite keys")
	blockCommand = flag.Bool("wb", false, "Overwrite block by [0-15]")

	spiDevice = flag.String("spidev", "/dev/spidev0.0", "SPI device")

	key = flag.String("key", "", "Comma-separated key bytes")
)

func main() {

	flag.Parse()

	fmt.Println("Using pin settings: rsPin", *rsPin, "irqPin", *irqPin)

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	currentAccessKey := rf522.DefaultKey

	if *key != "" {
		keyBytes := strings.SplitN(*key, ",", 6)
		if len(keyBytes) != 6 {
			log.Fatal("Key should consist of 6 decimal numbers")
		}
		currentAccessKey = make([]byte, 6)

		for i, v := range keyBytes {
			intV, err := strconv.Atoi(v)
			if err != nil {
				log.Fatal(err)
			}
			currentAccessKey[i] = byte(intV)
		}
	}

	currentAccessMethod := byte(commands.PICC_AUTHENT1B)

	// use BCM numbering here
	log.SetOutput(os.Stdout)

	spiDev, err := spireg.Open(*spiDevice)
	if err != nil {
		log.Fatal(err)
	}

	rfid, err := rf522.NewSPI(spiDev, 1000000, *rsPin, *irqPin)
	if err != nil {
		log.Fatal(err)
	}

	data, err := rfid.ReadCard(currentAccessMethod, *sector, *block, currentAccessKey[:])
	if err != nil {
		log.Fatal(err)
	}
	auth, err := rfid.ReadAuth(currentAccessMethod, *sector, currentAccessKey[:])
	if err != nil {
		log.Fatal(err)
	}

	access := rf522.ParseBlockAccess(auth[6:10])

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RFID sector %d, block %d : %v, auth: %v\n", *sector, *block, data, auth)
	fmt.Printf("Permissions: B0: %s, B1: %s, B2: %s, B3/A: %s\n",
		strconv.FormatUint(uint64(access.B0), 2),
		strconv.FormatUint(uint64(access.B1), 2),
		strconv.FormatUint(uint64(access.B2), 2),
		strconv.FormatUint(uint64(access.B3), 2),
	)

	if *keyCommand {
		err = rfid.WriteSectorTrail(commands.PICC_AUTHENT1A,
			*sector,
			[6]byte{1, 2, 3, 4, 5, 6},
			[6]byte{6, 5, 4, 3, 2, 1},
			&rf522.BlocksAccess{
				B0: rf522.RAB_WB_IB_DAB,
				B1: rf522.RB_WB_IN_DN,
				B2: rf522.AnyKeyRWID,
				B3: rf522.KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN,
			},
			currentAccessKey[:],
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Write successful")
	} else if *blockCommand {
		err = rfid.WriteBlock(currentAccessMethod,
			*sector,
			*block,
			[16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			currentAccessKey[:])
		if err != nil {
			log.Fatal(err)
		}
	}

}
