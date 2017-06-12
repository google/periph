// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// lepton captures a single image, prints metadata about the camera state or
// triggers a calibration.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/lepton"
	"periph.io/x/periph/host"
)

var palette = []color.NRGBA{
	{255, 255, 255, 255},
	{253, 253, 253, 255},
	{251, 251, 251, 255},
	{249, 249, 249, 255},
	{247, 247, 247, 255},
	{245, 245, 245, 255},
	{243, 243, 243, 255},
	{241, 241, 241, 255},
	{239, 239, 239, 255},
	{237, 237, 237, 255},
	{235, 235, 235, 255},
	{233, 233, 233, 255},
	{231, 231, 231, 255},
	{229, 229, 229, 255},
	{227, 227, 227, 255},
	{225, 225, 225, 255},
	{223, 223, 223, 255},
	{221, 221, 221, 255},
	{219, 219, 219, 255},
	{217, 217, 217, 255},
	{215, 215, 215, 255},
	{213, 213, 213, 255},
	{211, 211, 211, 255},
	{209, 209, 209, 255},
	{207, 207, 207, 255},
	{205, 205, 205, 255},
	{203, 203, 203, 255},
	{201, 201, 201, 255},
	{199, 199, 199, 255},
	{197, 197, 197, 255},
	{195, 195, 195, 255},
	{193, 193, 193, 255},
	{191, 191, 191, 255},
	{189, 189, 189, 255},
	{187, 187, 187, 255},
	{185, 185, 185, 255},
	{183, 183, 183, 255},
	{181, 181, 181, 255},
	{179, 179, 179, 255},
	{177, 177, 177, 255},
	{175, 175, 175, 255},
	{173, 173, 173, 255},
	{171, 171, 171, 255},
	{169, 169, 169, 255},
	{167, 167, 167, 255},
	{165, 165, 165, 255},
	{163, 163, 163, 255},
	{161, 161, 161, 255},
	{159, 159, 159, 255},
	{157, 157, 157, 255},
	{155, 155, 155, 255},
	{153, 153, 153, 255},
	{151, 151, 151, 255},
	{149, 149, 149, 255},
	{147, 147, 147, 255},
	{145, 145, 145, 255},
	{143, 143, 143, 255},
	{141, 141, 141, 255},
	{139, 139, 139, 255},
	{137, 137, 137, 255},
	{135, 135, 135, 255},
	{133, 133, 133, 255},
	{131, 131, 131, 255},
	{129, 129, 129, 255},
	{126, 126, 126, 255},
	{124, 124, 124, 255},
	{122, 122, 122, 255},
	{120, 120, 120, 255},
	{118, 118, 118, 255},
	{116, 116, 116, 255},
	{114, 114, 114, 255},
	{112, 112, 112, 255},
	{110, 110, 110, 255},
	{108, 108, 108, 255},
	{106, 106, 106, 255},
	{104, 104, 104, 255},
	{102, 102, 102, 255},
	{100, 100, 100, 255},
	{98, 98, 98, 255},
	{96, 96, 96, 255},
	{94, 94, 94, 255},
	{92, 92, 92, 255},
	{90, 90, 90, 255},
	{88, 88, 88, 255},
	{86, 86, 86, 255},
	{84, 84, 84, 255},
	{82, 82, 82, 255},
	{80, 80, 80, 255},
	{78, 78, 78, 255},
	{76, 76, 76, 255},
	{74, 74, 74, 255},
	{72, 72, 72, 255},
	{70, 70, 70, 255},
	{68, 68, 68, 255},
	{66, 66, 66, 255},
	{64, 64, 64, 255},
	{62, 62, 62, 255},
	{60, 60, 60, 255},
	{58, 58, 58, 255},
	{56, 56, 56, 255},
	{54, 54, 54, 255},
	{52, 52, 52, 255},
	{50, 50, 50, 255},
	{48, 48, 48, 255},
	{46, 46, 46, 255},
	{44, 44, 44, 255},
	{42, 42, 42, 255},
	{40, 40, 40, 255},
	{38, 38, 38, 255},
	{36, 36, 36, 255},
	{34, 34, 34, 255},
	{32, 32, 32, 255},
	{30, 30, 30, 255},
	{28, 28, 28, 255},
	{26, 26, 26, 255},
	{24, 24, 24, 255},
	{22, 22, 22, 255},
	{20, 20, 20, 255},
	{18, 18, 18, 255},
	{16, 16, 16, 255},
	{14, 14, 14, 255},
	{12, 12, 12, 255},
	{10, 10, 10, 255},
	{8, 8, 8, 255},
	{6, 6, 6, 255},
	{4, 4, 4, 255},
	{2, 2, 2, 255},
	{0, 0, 0, 255},
	{0, 0, 9, 255},
	{2, 0, 16, 255},
	{4, 0, 24, 255},
	{6, 0, 31, 255},
	{8, 0, 38, 255},
	{10, 0, 45, 255},
	{12, 0, 53, 255},
	{14, 0, 60, 255},
	{17, 0, 67, 255},
	{19, 0, 74, 255},
	{21, 0, 82, 255},
	{23, 0, 89, 255},
	{25, 0, 96, 255},
	{27, 0, 103, 255},
	{29, 0, 111, 255},
	{31, 0, 118, 255},
	{36, 0, 120, 255},
	{41, 0, 121, 255},
	{46, 0, 122, 255},
	{51, 0, 123, 255},
	{56, 0, 124, 255},
	{61, 0, 125, 255},
	{66, 0, 126, 255},
	{71, 0, 127, 255},
	{76, 1, 128, 255},
	{81, 1, 129, 255},
	{86, 1, 130, 255},
	{91, 1, 131, 255},
	{96, 1, 132, 255},
	{101, 1, 133, 255},
	{106, 1, 134, 255},
	{111, 1, 135, 255},
	{116, 1, 136, 255},
	{121, 1, 136, 255},
	{125, 2, 137, 255},
	{130, 2, 137, 255},
	{135, 3, 137, 255},
	{139, 3, 138, 255},
	{144, 3, 138, 255},
	{149, 4, 138, 255},
	{153, 4, 139, 255},
	{158, 5, 139, 255},
	{163, 5, 139, 255},
	{167, 5, 140, 255},
	{172, 6, 140, 255},
	{177, 6, 140, 255},
	{181, 7, 141, 255},
	{186, 7, 141, 255},
	{189, 10, 137, 255},
	{191, 13, 132, 255},
	{194, 16, 127, 255},
	{196, 19, 121, 255},
	{198, 22, 116, 255},
	{200, 25, 111, 255},
	{203, 28, 106, 255},
	{205, 31, 101, 255},
	{207, 34, 95, 255},
	{209, 37, 90, 255},
	{212, 40, 85, 255},
	{214, 43, 80, 255},
	{216, 46, 75, 255},
	{218, 49, 69, 255},
	{221, 52, 64, 255},
	{223, 55, 59, 255},
	{224, 57, 49, 255},
	{225, 60, 47, 255},
	{226, 64, 44, 255},
	{227, 67, 42, 255},
	{228, 71, 39, 255},
	{229, 74, 37, 255},
	{230, 78, 34, 255},
	{231, 81, 32, 255},
	{231, 85, 29, 255},
	{232, 88, 27, 255},
	{233, 92, 24, 255},
	{234, 95, 22, 255},
	{235, 99, 19, 255},
	{236, 102, 17, 255},
	{237, 106, 14, 255},
	{238, 109, 12, 255},
	{239, 112, 12, 255},
	{240, 116, 12, 255},
	{240, 119, 12, 255},
	{241, 123, 12, 255},
	{241, 127, 12, 255},
	{242, 130, 12, 255},
	{242, 134, 12, 255},
	{243, 138, 12, 255},
	{243, 141, 13, 255},
	{244, 145, 13, 255},
	{244, 149, 13, 255},
	{245, 152, 13, 255},
	{245, 156, 13, 255},
	{246, 160, 13, 255},
	{246, 163, 13, 255},
	{247, 167, 13, 255},
	{247, 171, 13, 255},
	{248, 175, 14, 255},
	{248, 178, 15, 255},
	{249, 182, 16, 255},
	{249, 185, 18, 255},
	{250, 189, 19, 255},
	{250, 192, 20, 255},
	{251, 196, 21, 255},
	{251, 199, 22, 255},
	{252, 203, 23, 255},
	{252, 206, 24, 255},
	{253, 210, 25, 255},
	{253, 213, 27, 255},
	{254, 217, 28, 255},
	{254, 220, 29, 255},
	{255, 224, 30, 255},
	{255, 227, 39, 255},
	{255, 229, 53, 255},
	{255, 231, 67, 255},
	{255, 233, 81, 255},
	{255, 234, 95, 255},
	{255, 236, 109, 255},
	{255, 238, 123, 255},
	{255, 240, 137, 255},
	{255, 242, 151, 255},
	{255, 244, 165, 255},
	{255, 246, 179, 255},
	{255, 248, 193, 255},
	{255, 249, 207, 255},
	{255, 251, 221, 255},
	{255, 253, 235, 255},
	{255, 255, 24, 255},
}

// reduce the intensity of a 14 bits images into 8 bits centered at midpoint.
//
// No AGC is done.
func reduce(src *image.Gray16) *image.NRGBA {
	midPoint := uint16(8192)
	base := midPoint - uint16(len(palette)/2)
	max := base + uint16(len(palette)) - 1
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := src.Gray16At(x, y).Y
			if i < base {
				i = base
			} else if i > max {
				i = max
			}
			dst.SetNRGBA(x, y, palette[i-base])
		}
	}
	return dst
}

// query queries the camera and prints out information about its current state.
func query(dev *lepton.Dev) error {
	status, err := dev.GetStatus()
	if err != nil {
		return err
	}
	fmt.Printf("Status.CameraStatus:             %s\n", status.CameraStatus)
	fmt.Printf("Status.CommandCount:             %d\n", status.CommandCount)
	serial, err := dev.GetSerial()
	if err != nil {
		return err
	}
	fmt.Printf("Serial:                          0x%x\n", serial)
	uptime, err := dev.GetUptime()
	if err != nil {
		return err
	}
	fmt.Printf("Uptime:                          %s\n", uptime)
	temp, err := dev.GetTemp()
	if err != nil {
		return err
	}
	fmt.Printf("Temp:                            %s\n", temp)
	temp, err = dev.GetTempHousing()
	if err != nil {
		return err
	}
	fmt.Printf("Temp housing:                    %s\n", temp)
	pos, err := dev.GetShutterPos()
	if err != nil {
		return err
	}
	fmt.Printf("ShutterPos:                      %s\n", pos)
	mode, err := dev.GetFFCModeControl()
	if err != nil {
		return err
	}
	fmt.Printf("FCCMode.FFCShutterMode:          %s\n", mode.FFCShutterMode)
	fmt.Printf("FCCMode.ShutterTempLockoutState: %s\n", mode.ShutterTempLockoutState)
	fmt.Printf("FCCMode.VideoFreezeDuringFFC:    %t\n", mode.VideoFreezeDuringFFC)
	fmt.Printf("FCCMode.FFCDesired:              %t\n", mode.FFCDesired)
	fmt.Printf("FCCMode.ElapsedTimeSinceLastFFC: %s\n", mode.ElapsedTimeSinceLastFFC)
	fmt.Printf("FCCMode.DesiredFFCPeriod:        %s\n", mode.DesiredFFCPeriod)
	fmt.Printf("FCCMode.ExplicitCommandToOpen:   %t\n", mode.ExplicitCommandToOpen)
	fmt.Printf("FCCMode.DesiredFFCTempDelta:     %s\n", mode.DesiredFFCTempDelta)
	fmt.Printf("FCCMode.ImminentDelay:           %d\n", mode.ImminentDelay)
	return nil
}

// grabFrame grabs one frame from a lepton and saves it as colored palette PNG
// file.
func grabFrame(dev *lepton.Dev, path string, meta bool) error {
	frame, err := dev.ReadImg()
	if err != nil {
		return err
	}
	if meta {
		fmt.Printf("SinceStartup: %s\n", frame.Metadata.SinceStartup)
		fmt.Printf("FrameCount:   %d\n", frame.Metadata.FrameCount)
		fmt.Printf("Temp:         %s\n", frame.Metadata.Temp)
		fmt.Printf("TempHousing:  %s\n", frame.Metadata.TempHousing)
		fmt.Printf("FFCSince:     %s\n", frame.Metadata.FFCSince)
		fmt.Printf("FFCDesired:   %t\n", frame.Metadata.FFCDesired)
		fmt.Printf("Overtemp:     %t\n", frame.Metadata.Overtemp)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, reduce(frame.Gray16))
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	spiID := flag.String("spi", "", "SPI port to use")
	csID := flag.String("cs", "", "SPI CS line to use instead of the default")
	i2cHz := flag.Int("i2chz", 0, "I²C bus speed")
	spiHz := flag.Int("spihz", 0, "SPI port speed")

	meta := flag.Bool("meta", false, "print metadata")
	output := flag.String("o", "", "PNG file to save")
	ffc := flag.Bool("ffc", false, "trigger a calibration")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unsupported arguments")
	}
	hasOutput := len(*output) != 0
	if hasOutput && *ffc {
		return errors.New("can't use both -o and -ffc")
	}
	if !hasOutput && *meta {
		return errors.New("-meta requires -o")
	}

	// Initialization.
	if _, err := host.Init(); err != nil {
		return err
	}
	spiPort, err := spireg.Open(*spiID)
	if err != nil {
		return err
	}
	defer spiPort.Close()
	if *spiHz != 0 {
		if err := spiPort.LimitSpeed(int64(*spiHz)); err != nil {
			return err
		}
	}
	var cs gpio.PinOut
	if len(*csID) != 0 {
		if p := gpioreg.ByName(*csID); p != nil {
			cs = p
		} else {
			return fmt.Errorf("%s is not a valid pin", *csID)
		}
	}

	i2cBus, err := i2creg.Open(*i2cID)
	if err != nil {
		return err
	}
	defer i2cBus.Close()
	if *i2cHz != 0 {
		if err := i2cBus.SetSpeed(int64(*i2cHz)); err != nil {
			return err
		}
	}
	dev, err := lepton.New(spiPort, i2cBus, cs)
	if err != nil {
		return fmt.Errorf("%s\nIf testing without hardware, use -fake to simulate a camera", err)
	}

	// Action.
	if *ffc {
		return dev.RunFFC()
	}
	if hasOutput {
		return grabFrame(dev, *output, *meta)
	}
	return query(dev)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "\nlepton: %s.\n", err)
		os.Exit(1)
	}
}
