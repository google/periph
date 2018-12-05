// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// mfrc522 reads RFID tags.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host"
)

func mainImpl() error {

	sector := flag.Int("sector", 1, "Sector to access")
	block := flag.Int("block", 0, "Block to access")

	rsPin := flag.String("rs-pin", "", "Reset pin")
	irqPin := flag.String("irq-pin", "", "IRQ pin")

	keyCommand := flag.Bool("wa", false, "Overwrite keys")
	blockCommand := flag.String("wb", "", "Overwrite block by provided data (comma-separated list of 16 bytes)")

	spiID := flag.String("spi", "", "SPI device")

	key := flag.String("key", "", "Comma-separated key bytes")

	flag.Parse()

	if *irqPin == "" || *rsPin == "" {
		return errors.New("please provide -rs-pin and -irq-pin arguments, or -h for help")
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	var currentAccessKey [6]byte

	if *key != "" {
		keyBytes := strings.SplitN(*key, ",", 6)
		if len(keyBytes) != 6 {
			return errors.New("key should consist of 6 decimal numbers")
		}
		for i, v := range keyBytes {
			intV, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return err
			}
			currentAccessKey[i] = byte(intV)
		}
	} else {
		copy(currentAccessKey[:], mfrc522.DefaultKey[:])
	}

	currentAccessMethod := byte(commands.PICC_AUTHENT1B)

	spiDev, err := spireg.Open(*spiID)
	if err != nil {
		return err
	}

	rsPinReg := gpioreg.ByName(*rsPin)
	if rsPinReg == nil {
		return fmt.Errorf("Reset pin %s can not be found", *rsPin)
	}

	irqPinReg := gpioreg.ByName(*irqPin)
	if rsPinReg == nil {
		return fmt.Errorf("IRQ pin %s can not be found", *irqPin)
	}

	rfid, err := mfrc522.NewSPI(spiDev, rsPinReg, irqPinReg)
	if err != nil {
		return err
	}

	data, err := rfid.ReadCard(currentAccessMethod, *sector, *block, currentAccessKey)
	if err != nil {
		return err
	}
	auth, err := rfid.ReadAuth(currentAccessMethod, *sector, currentAccessKey)
	if err != nil {
		return err
	}

	var access mfrc522.BlocksAccess

	access.Init(auth[6:10])

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
			&mfrc522.BlocksAccess{
				B0: mfrc522.RAB_WB_IB_DAB,
				B1: mfrc522.RB_WB_IN_DN,
				B2: mfrc522.AnyKeyRWID,
				B3: mfrc522.KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN,
			},
			currentAccessKey,
		)
		if err != nil {
			return err
		}
		fmt.Println("Write successful")
	} else if *blockCommand != "" {
		var defaultDataBytes [16]byte
		bytesBuffer := strings.Split(*blockCommand, ",")
		if len(bytesBuffer) != 16 {
			return errors.New("data bytes must contain exactly 16 elements")
		}
		for i := range defaultDataBytes {
			intVal, err := strconv.ParseUint(bytesBuffer[i], 10, 8)
			if err != nil {
				return err
			}
			defaultDataBytes[i] = byte(intVal)
		}
		err = rfid.WriteCard(currentAccessMethod,
			*sector,
			*block,
			defaultDataBytes,
			currentAccessKey)
		if err != nil {
			return err
		}
	}

	return rfid.Halt()
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "mfrc522: %s.\n", err)
		os.Exit(1)
	}
}
