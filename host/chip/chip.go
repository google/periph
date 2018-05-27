// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/allwinner"
	"periph.io/x/periph/host/distro"
	"periph.io/x/periph/host/fs"
)

// C.H.I.P. hardware pins.
var (
	TEMP_SENSOR = &pin.BasicPin{N: "TEMP_SENSOR"}
	PWR_SWITCH  = &pin.BasicPin{N: "PWR_SWITCH"}
	// XIO "gpio" pins attached to the pcf8574 I²C port extender.
	XIO0, XIO1, XIO2, XIO3, XIO4, XIO5, XIO6, XIO7 gpio.PinIO
)

// The U13 header is opposite the power LED.
//
// The alternate pin functionality is described at pages 322-323 of
// https://github.com/NextThingCo/CHIP-Hardware/raw/master/CHIP%5Bv1_0%5D/CHIPv1_0-BOM-Datasheets/Allwinner%20R8%20User%20Manual%20V1.1.pdf
var (
	U13_1  = pin.GROUND     //
	U13_2  = pin.DC_IN      //
	U13_3  = pin.V5         // (filtered)
	U13_4  = pin.GROUND     //
	U13_5  = pin.V3_3       //
	U13_6  = TEMP_SENSOR    // Analog temp sensor input
	U13_7  = pin.V1_8       //
	U13_8  = pin.BAT_PLUS   // External LiPo battery
	U13_9  = allwinner.PB16 // I2C1_SDA
	U13_10 = PWR_SWITCH     // Power button
	U13_11 = allwinner.PB15 // I2C1_SCL
	U13_12 = pin.GROUND     //
	U13_13 = allwinner.X1   // Touch screen X1
	U13_14 = allwinner.X2   // Touch screen X2
	U13_15 = allwinner.Y1   // Touch screen Y1
	U13_16 = allwinner.Y2   // Touch screen Y2
	U13_17 = allwinner.PD2  // LCD-D2; UART2_TX firmware probe for 1-wire to detect DIP at boot; http://docs.getchip.com/dip.html#dip-identification
	U13_18 = allwinner.PB2  // PWM0; EINT16
	U13_19 = allwinner.PD4  // LCD-D4; UART2_CTS
	U13_20 = allwinner.PD3  // LCD-D3; UART2_RX
	U13_21 = allwinner.PD6  // LCD-D6
	U13_22 = allwinner.PD5  // LCD-D5
	U13_23 = allwinner.PD10 // LCD-D10
	U13_24 = allwinner.PD7  // LCD-D7
	U13_25 = allwinner.PD12 // LCD-D12
	U13_26 = allwinner.PD11 // LCD-D11
	U13_27 = allwinner.PD14 // LCD-D14
	U13_28 = allwinner.PD13 // LCD-D13
	U13_29 = allwinner.PD18 // LCD-D18
	U13_30 = allwinner.PD15 // LCD-D15
	U13_31 = allwinner.PD20 // LCD-D20
	U13_32 = allwinner.PD19 // LCD-D19
	U13_33 = allwinner.PD22 // LCD-D22
	U13_34 = allwinner.PD21 // LCD-D21
	U13_35 = allwinner.PD24 // LCD-CLK
	U13_36 = allwinner.PD23 // LCD-D23
	U13_37 = allwinner.PD26 // LCD-VSYNC
	U13_38 = allwinner.PD27 // LCD-HSYNC
	U13_39 = pin.GROUND     //
	U13_40 = allwinner.PD25 // LCD-DE: RGB666 data
)

// The U14 header is right next to the power LED.
var (
	U14_1  = pin.GROUND         //
	U14_2  = pin.V5             // (filtered)
	U14_3  = allwinner.PG3      // UART1_TX; EINT3
	U14_4  = allwinner.HP_LEFT  // Headphone left output
	U14_5  = allwinner.PG4      // UART1_RX; EINT4
	U14_6  = allwinner.HP_COM   // Headphone amp out
	U14_7  = allwinner.FEL      // Boot mode selection
	U14_8  = allwinner.HP_RIGHT // Headphone right output
	U14_9  = pin.V3_3           //
	U14_10 = allwinner.MIC_GND  // Microphone ground
	U14_11 = allwinner.KEY_ADC  // LRADC Low res analog to digital
	U14_12 = allwinner.MIC_IN   // Microphone input
	U14_13 = XIO0               // gpio via I²C controller
	U14_14 = XIO1               // gpio via I²C controller
	U14_15 = XIO2               // gpio via I²C controller
	U14_16 = XIO3               // gpio via I²C controller
	U14_17 = XIO4               // gpio via I²C controller
	U14_18 = XIO5               // gpio via I²C controller
	U14_19 = XIO6               // gpio via I²C controller
	U14_20 = XIO7               // gpio via I²C controller
	U14_21 = pin.GROUND         //
	U14_22 = pin.GROUND         //
	U14_23 = allwinner.PG1      // GPS_CLK; AP-EINT1
	U14_24 = allwinner.PB3      // IR_TX; AP-EINT3 (EINT17)
	U14_25 = allwinner.PB18     // I2C2_SDA
	U14_26 = allwinner.PB17     // I2C2_SCL
	U14_27 = allwinner.PE0      // CSIPCK: CMOS serial interface; SPI2_CS0; EINT14
	U14_28 = allwinner.PE1      // CSICK: CMOS serial interface; SPI2_CLK; EINT15
	U14_29 = allwinner.PE2      // CSIHSYNC; SPI2_MOSI
	U14_30 = allwinner.PE3      // CSIVSYNC; SPI2_MISO
	U14_31 = allwinner.PE4      // CSID0
	U14_32 = allwinner.PE5      // CSID1
	U14_33 = allwinner.PE6      // CSID2
	U14_34 = allwinner.PE7      // CSID3
	U14_35 = allwinner.PE8      // CSID4
	U14_36 = allwinner.PE9      // CSID5
	U14_37 = allwinner.PE10     // CSID6; UART1_RX
	U14_38 = allwinner.PE11     // CSID7; UART1_TX
	U14_39 = pin.GROUND         //
	U14_40 = pin.GROUND         //
)

// Present returns true if running on a NextThing Co's C.H.I.P. board.
//
// It looks for "C.H.I.P" in the device tree. The following information is
// expected in the device dtree:
//   root@chip2:/proc/device-tree# od -c compatible
//   0000000   n   e   x   t   t   h   i   n   g   ,   c   h   i   p  \0   a
//   0000020   l   l   w   i   n   n   e   r   ,   s   u   n   5   i   -   r
//   0000040   8  \0
//   root@chip2:/proc/device-tree# od -c model
//   0000000   N   e   x   t   T   h   i   n   g       C   .   H   .   I   .
//   0000020   P   .  \0
func Present() bool {
	return strings.Contains(distro.DTModel(), "C.H.I.P")
}

//

// aliases is a list of aliases for the various gpio pins, this allows users to
// refer to pins using the documented and labeled names instead of some GPIOnnn
// name. The map key is the alias and the value is the real pin name.
var aliases = map[string]string{
	"AP-EINT1":  "PG1",
	"AP-EINT3":  "PB3",
	"CSIPCK":    "PE0",
	"CSIHSYNC":  "PE2",
	"CSID0":     "PE4",
	"CSID2":     "PE6",
	"CSID4":     "PE8",
	"CSID6":     "PE10",
	"CSICK":     "PE1",
	"CSIVSYNC":  "PE3",
	"CSID1":     "PE5",
	"CSID3":     "PE7",
	"CSID5":     "PE9",
	"CSID7":     "PE11",
	"LCD-CLK":   "PD24",
	"LCD-D10":   "PD10",
	"LCD-D11":   "PD11",
	"LCD-D12":   "PD12",
	"LCD-D13":   "PD13",
	"LCD-D14":   "PD14",
	"LCD-D15":   "PD15",
	"LCD-D18":   "PD18",
	"LCD-D19":   "PD19",
	"LCD-D2":    "PD2",
	"LCD-D20":   "PD20",
	"LCD-D21":   "PD21",
	"LCD-D22":   "PD22",
	"LCD-D23":   "PD23",
	"LCD-D3":    "PD3",
	"LCD-D4":    "PD4",
	"LCD-D5":    "PD5",
	"LCD-D6":    "PD6",
	"LCD-D7":    "PD7",
	"LCD-DE":    "PD25",
	"LCD-HSYNC": "PD27",
	"LCD-VSYNC": "PD26",
	"TWI1-SCK":  "PB15",
	"TWI1-SDA":  "PB16",
	"TWI2-SCK":  "PB17",
	"TWI2-SDA":  "PB18",
	"UART1-RX":  "PG4",
	"UART1-TX":  "PG3",
}

func init() {
	// These are initialized later by the driver.
	XIO0 = gpio.INVALID
	XIO1 = gpio.INVALID
	XIO2 = gpio.INVALID
	XIO3 = gpio.INVALID
	XIO4 = gpio.INVALID
	XIO5 = gpio.INVALID
	XIO6 = gpio.INVALID
	XIO7 = gpio.INVALID
	// These must be reinitialized.
	U14_13 = XIO0
	U14_14 = XIO1
	U14_15 = XIO2
	U14_16 = XIO3
	U14_17 = XIO4
	U14_18 = XIO5
	U14_19 = XIO6
	U14_20 = XIO7
}

// findXIOBase calculates the base of the XIO-P? gpio pins as explained in
// http://docs.getchip.com/chip.html#kernel-4-3-vs-4-4-gpio-how-to-tell-the-difference
//
// The XIO-P? sysfs mapped pin number changed in kernel 4.3, 4.4.11 and again
// in 4.4.13 so it is better to query sysfs.
func findXIOBase() int {
	chips, err := filepath.Glob("/sys/class/gpio/gpiochip*/label")
	if err != nil {
		return -1
	}
	for _, item := range chips {
		f, err := fs.Open(item, os.O_RDONLY)
		if err != nil {
			continue
		}
		b, err := ioutil.ReadAll(f)
		if err1 := f.Close(); err == nil {
			err = err1
		}
		if err != nil {
			continue
		}
		if string(b) == "pcf8574a\n" {
			id, err := strconv.Atoi(filepath.Base(filepath.Dir(item))[8:])
			if err != nil {
				return -1
			}
			return id
		}
	}
	return -1
}

// driver implements drivers.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "chip"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	// has allwinner cpu, needs sysfs for XIO0-XIO7 "gpio" pins
	return []string{"allwinner-gpio", "sysfs-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("NextThing Co. CHIP board not detected")
	}

	base := findXIOBase()
	if base == -1 {
		return true, errors.New("couldn't find XIO pins base number")
	}
	for i := 0; i < 8; i++ {
		aliases[fmt.Sprintf("XIO-P%d", i)] = fmt.Sprintf("GPIO%d", base+i)
	}

	// At this point the sysfs driver has initialized and discovered its pins,
	// we can now hook-up the appropriate CHIP pins to sysfs gpio pins.
	for alias, real := range aliases {
		if err := gpioreg.RegisterAlias(alias, real); err != nil {
			return true, err
		}
	}
	// These must be explicitly initialized.
	XIO0 = gpioreg.ByName("XIO-P0")
	XIO1 = gpioreg.ByName("XIO-P1")
	XIO2 = gpioreg.ByName("XIO-P2")
	XIO3 = gpioreg.ByName("XIO-P3")
	XIO4 = gpioreg.ByName("XIO-P4")
	XIO5 = gpioreg.ByName("XIO-P5")
	XIO6 = gpioreg.ByName("XIO-P6")
	XIO7 = gpioreg.ByName("XIO-P7")
	U14_13 = XIO0
	U14_14 = XIO1
	U14_15 = XIO2
	U14_16 = XIO3
	U14_17 = XIO4
	U14_18 = XIO5
	U14_19 = XIO6
	U14_20 = XIO7

	// U13 is one of the 20x2 connectors.
	U13 := [][]pin.Pin{
		{U13_1, U13_2},
		{U13_3, U13_4},
		{U13_5, U13_6},
		{U13_7, U13_8},
		{gpioreg.ByName("TWI1-SDA"), U13_10},
		{gpioreg.ByName("TWI1-SCK"), U13_12},
		{U13_13, U13_14},
		{U13_15, U13_16},
		{gpioreg.ByName("LCD-D2"), gpioreg.ByName("PWM0")},
		{gpioreg.ByName("LCD-D4"), gpioreg.ByName("LCD-D3")},
		{gpioreg.ByName("LCD-D6"), gpioreg.ByName("LCD-D5")},
		{gpioreg.ByName("LCD-D10"), gpioreg.ByName("LCD-D7")},
		{gpioreg.ByName("LCD-D12"), gpioreg.ByName("LCD-D11")},
		{gpioreg.ByName("LCD-D14"), gpioreg.ByName("LCD-D13")},
		{gpioreg.ByName("LCD-D18"), gpioreg.ByName("LCD-D15")},
		{gpioreg.ByName("LCD-D20"), gpioreg.ByName("LCD-D19")},
		{gpioreg.ByName("LCD-D22"), gpioreg.ByName("LCD-D21")},
		{gpioreg.ByName("LCD-CLK"), gpioreg.ByName("LCD-D23")},
		{gpioreg.ByName("LCD-VSYNC"), gpioreg.ByName("LCD-HSYNC")},
		{U13_39, gpioreg.ByName("LCD-DE")},
	}
	if err := pinreg.Register("U13", U13); err != nil {
		return true, err
	}

	// U14 is one of the 20x2 connectors.
	U14 := [][]pin.Pin{
		{U14_1, U14_2},
		{gpioreg.ByName("UART1-TX"), U14_4},
		{gpioreg.ByName("UART1-RX"), U14_6},
		{U14_7, U14_8},
		{U14_9, U14_10},
		{U14_11, U14_12}, // TODO(maruel): switch to LRADC once analog support is added
		{U14_13, U14_14},
		{U14_15, U14_16},
		{U14_17, U14_18},
		{U14_19, U14_20},
		{U14_21, U14_22},
		{gpioreg.ByName("AP-EINT1"), gpioreg.ByName("AP-EINT3")},
		{gpioreg.ByName("TWI2-SDA"), gpioreg.ByName("TWI2-SCK")},
		{gpioreg.ByName("CSIPCK"), gpioreg.ByName("CSICK")},
		{gpioreg.ByName("CSIHSYNC"), gpioreg.ByName("CSIVSYNC")},
		{gpioreg.ByName("CSID0"), gpioreg.ByName("CSID1")},
		{gpioreg.ByName("CSID2"), gpioreg.ByName("CSID3")},
		{gpioreg.ByName("CSID4"), gpioreg.ByName("CSID5")},
		{gpioreg.ByName("CSID6"), gpioreg.ByName("CSID7")},
		{U14_39, U14_40},
	}
	return true, pinreg.Register("U14", U14)
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
