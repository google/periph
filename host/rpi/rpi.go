// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Raspberry Pi pin out.

package rpi

import (
	"errors"
	"fmt"
	"os"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/bcm283x"
	"periph.io/x/periph/host/distro"
)

// Present returns true if running on a Raspberry Pi board.
//
// https://www.raspberrypi.org/
func Present() bool {
	if isArm {
		// This is iffy at best.
		_, err := os.Stat("/sys/bus/platform/drivers/raspberrypi-firmware")
		return err == nil
	}
	return false
}

// Pin as connect on the 40 pins extension header.
//
// Schematics are useful to know what is connected to what:
// https://www.raspberrypi.org/documentation/hardware/raspberrypi/schematics/README.md
//
// The actual pin mapping depends on the board revision! The default values are
// set as the 40 pins header on Raspberry Pi 2 and Raspberry Pi 3.
//
// Some header info here: http://elinux.org/RPi_Low-level_peripherals
//
// P1 is also known as J8 on A+, B+, 2 and later.
var (
	// Raspberry Pi A and B, 26 pin header:
	P1_1  pin.Pin    = pin.V3_3       // max 30mA
	P1_2  pin.Pin    = pin.V5         // (filtered)
	P1_3  gpio.PinIO = bcm283x.GPIO2  // High, I2C1_SDA
	P1_4  pin.Pin    = pin.V5         //
	P1_5  gpio.PinIO = bcm283x.GPIO3  // High, I2C1_SCL
	P1_6  pin.Pin    = pin.GROUND     //
	P1_7  gpio.PinIO = bcm283x.GPIO4  // High, CLK0
	P1_8  gpio.PinIO = bcm283x.GPIO14 // Low,  UART0_TX, UART1_TX
	P1_9  pin.Pin    = pin.GROUND     //
	P1_10 gpio.PinIO = bcm283x.GPIO15 // Low,  UART0_RX, UART1_RX
	P1_11 gpio.PinIO = bcm283x.GPIO17 // Low,  UART0_RTS, SPI1_CS1, UART1_RTS
	P1_12 gpio.PinIO = bcm283x.GPIO18 // Low,  I2S_SCK, SPI1_CS0, PWM0
	P1_13 gpio.PinIO = bcm283x.GPIO27 // Low,
	P1_14 pin.Pin    = pin.GROUND     //
	P1_15 gpio.PinIO = bcm283x.GPIO22 // Low,
	P1_16 gpio.PinIO = bcm283x.GPIO23 // Low,
	P1_17 pin.Pin    = pin.V3_3       //
	P1_18 gpio.PinIO = bcm283x.GPIO24 // Low,
	P1_19 gpio.PinIO = bcm283x.GPIO10 // Low, SPI0_MOSI
	P1_20 pin.Pin    = pin.GROUND     //
	P1_21 gpio.PinIO = bcm283x.GPIO9  // Low, SPI0_MISO
	P1_22 gpio.PinIO = bcm283x.GPIO25 // Low,
	P1_23 gpio.PinIO = bcm283x.GPIO11 // Low, SPI0_CLK
	P1_24 gpio.PinIO = bcm283x.GPIO8  // High, SPI0_CS0
	P1_25 pin.Pin    = pin.GROUND     //
	P1_26 gpio.PinIO = bcm283x.GPIO7  // High, SPI0_CS1

	// Raspberry Pi A+, B+, 2 and later, 40 pin header (also named J8):
	P1_27 gpio.PinIO = bcm283x.GPIO0  // High, I2C0_SDA used to probe for HAT EEPROM, see https://github.com/raspberrypi/hats
	P1_28 gpio.PinIO = bcm283x.GPIO1  // High, I2C0_SCL
	P1_29 gpio.PinIO = bcm283x.GPIO5  // High, CLK1
	P1_30 pin.Pin    = pin.GROUND     //
	P1_31 gpio.PinIO = bcm283x.GPIO6  // High, CLK2
	P1_32 gpio.PinIO = bcm283x.GPIO12 // Low,  PWM0
	P1_33 gpio.PinIO = bcm283x.GPIO13 // Low,  PWM1
	P1_34 pin.Pin    = pin.GROUND     //
	P1_35 gpio.PinIO = bcm283x.GPIO19 // Low,  I2S_WS, SPI1_MISO, PWM1
	P1_36 gpio.PinIO = bcm283x.GPIO16 // Low,  UART0_CTS, SPI1_CS2, UART1_CTS
	P1_37 gpio.PinIO = bcm283x.GPIO26 //
	P1_38 gpio.PinIO = bcm283x.GPIO20 // Low,  I2S_DIN, SPI1_MOSI, CLK0
	P1_39 pin.Pin    = pin.GROUND     //
	P1_40 gpio.PinIO = bcm283x.GPIO21 // Low,  I2S_DOUT, SPI1_CLK, CLK1

	// P5 header on Raspberry Pi A and B, PCB v2:
	P5_1 pin.Pin    = pin.V5
	P5_2 pin.Pin    = pin.V3_3
	P5_3 gpio.PinIO = bcm283x.GPIO28 // Float, I2C0_SDA, I2S_SCK
	P5_4 gpio.PinIO = bcm283x.GPIO29 // Float, I2C0_SCL, I2S_WS
	P5_5 gpio.PinIO = bcm283x.GPIO30 // Low,   I2S_DIN, UART0_CTS, UART1_CTS
	P5_6 gpio.PinIO = bcm283x.GPIO31 // Low,   I2S_DOUT, UART0_RTS, UART1_RTS
	P5_7 pin.Pin    = pin.GROUND
	P5_8 pin.Pin    = pin.GROUND

	AUDIO_RIGHT         = bcm283x.GPIO40 // Low,   PWM0, SPI2_MISO, UART1_TX
	AUDIO_LEFT          = bcm283x.GPIO41 // Low,   PWM1, SPI2_MOSI, UART1_RX
	HDMI_HOTPLUG_DETECT = bcm283x.GPIO46 // High,
)

// Pin as connected on the SODIMM header.
//
// Documentation is https://www.raspberrypi.org/documentation/hardware/computemodule/datasheets/rpi_DATA_CM_1p0.pdf
//
// There are some differences for CM3-Lite and CM1.
var (
	SO_1   pin.Pin    = pin.GROUND     // GND
	SO_2   pin.Pin    = pin.INVALID    // EMMC_DISABLE_N
	SO_3   gpio.PinIO = bcm283x.GPIO0  // GPIO0
	SO_4   pin.Pin    = pin.INVALID    // NC, SDX_VDD, NC
	SO_5   gpio.PinIO = bcm283x.GPIO1  // GPIO1
	SO_6   pin.Pin    = pin.INVALID    // NC, SDX_VDD, NC
	SO_7   pin.Pin    = pin.GROUND     // GND
	SO_8   pin.Pin    = pin.GROUND     // GND
	SO_9   gpio.PinIO = bcm283x.GPIO2  // GPIO2
	SO_10  pin.Pin    = pin.INVALID    // NC, SDX_CLK, NC
	SO_11  gpio.PinIO = bcm283x.GPIO3  // GPIO3
	SO_12  pin.Pin    = pin.INVALID    // NC, SDX_CMD, NC
	SO_13  pin.Pin    = pin.GROUND     // GND
	SO_14  pin.Pin    = pin.GROUND     // GND
	SO_15  gpio.PinIO = bcm283x.GPIO4  // GPIO4
	SO_16  pin.Pin    = pin.INVALID    // NC, SDX_D0, NC
	SO_17  gpio.PinIO = bcm283x.GPIO5  // GPIO5
	SO_18  pin.Pin    = pin.INVALID    // NC, SDX_D1, NC
	SO_19  pin.Pin    = pin.GROUND     // GND
	SO_20  pin.Pin    = pin.GROUND     // GND
	SO_21  gpio.PinIO = bcm283x.GPIO6  // GPIO6
	SO_22  pin.Pin    = pin.INVALID    // NC, SDX_D2, NC
	SO_23  gpio.PinIO = bcm283x.GPIO7  // GPIO7
	SO_24  pin.Pin    = pin.INVALID    // NC, SDX_D3, NC
	SO_25  pin.Pin    = pin.GROUND     // GND
	SO_26  pin.Pin    = pin.GROUND     // GND
	SO_27  gpio.PinIO = bcm283x.GPIO8  // GPIO8
	SO_28  gpio.PinIO = bcm283x.GPIO28 // GPIO28
	SO_29  gpio.PinIO = bcm283x.GPIO9  // GPIO9
	SO_30  gpio.PinIO = bcm283x.GPIO29 // GPIO29
	SO_31  pin.Pin    = pin.GROUND     // GND
	SO_32  pin.Pin    = pin.GROUND     // GND
	SO_33  gpio.PinIO = bcm283x.GPIO10 // GPIO10
	SO_34  gpio.PinIO = bcm283x.GPIO30 // GPIO30
	SO_35  gpio.PinIO = bcm283x.GPIO11 // GPIO11
	SO_36  gpio.PinIO = bcm283x.GPIO31 // GPIO31
	SO_37  pin.Pin    = pin.GROUND     // GND
	SO_38  pin.Pin    = pin.GROUND     // GND
	SO_39  pin.Pin    = pin.DC_IN      // GPIO0-27_VDD
	SO_40  pin.Pin    = pin.DC_IN      // GPIO0-27_VDD
	SO_41  pin.Pin    = pin.DC_IN      // GPIO28-45_VDD
	SO_42  pin.Pin    = pin.DC_IN      // GPIO28-45_VDD
	SO_43  pin.Pin    = pin.GROUND     // GND
	SO_44  pin.Pin    = pin.GROUND     // GND
	SO_45  gpio.PinIO = bcm283x.GPIO12 // GPIO12
	SO_46  gpio.PinIO = bcm283x.GPIO32 // GPIO32
	SO_47  gpio.PinIO = bcm283x.GPIO13 // GPIO13
	SO_48  gpio.PinIO = bcm283x.GPIO33 // GPIO33
	SO_49  pin.Pin    = pin.GROUND     // GND
	SO_50  pin.Pin    = pin.GROUND     // GND
	SO_51  gpio.PinIO = bcm283x.GPIO14 // GPIO14
	SO_52  gpio.PinIO = bcm283x.GPIO34 // GPIO34
	SO_53  gpio.PinIO = bcm283x.GPIO15 // GPIO15
	SO_54  gpio.PinIO = bcm283x.GPIO35 // GPIO35
	SO_55  pin.Pin    = pin.GROUND     // GND
	SO_56  pin.Pin    = pin.GROUND     // GND
	SO_57  gpio.PinIO = bcm283x.GPIO16 // GPIO16
	SO_58  gpio.PinIO = bcm283x.GPIO36 // GPIO36
	SO_59  gpio.PinIO = bcm283x.GPIO17 // GPIO17
	SO_60  gpio.PinIO = bcm283x.GPIO37 // GPIO37
	SO_61  pin.Pin    = pin.GROUND     // GND
	SO_62  pin.Pin    = pin.GROUND     // GND
	SO_63  gpio.PinIO = bcm283x.GPIO18 // GPIO18
	SO_64  gpio.PinIO = bcm283x.GPIO38 // GPIO38
	SO_65  gpio.PinIO = bcm283x.GPIO19 // GPIO19
	SO_66  gpio.PinIO = bcm283x.GPIO39 // GPIO39
	SO_67  pin.Pin    = pin.GROUND     // GND
	SO_68  pin.Pin    = pin.GROUND     // GND
	SO_69  gpio.PinIO = bcm283x.GPIO20 // GPIO20
	SO_70  gpio.PinIO = bcm283x.GPIO40 // GPIO40
	SO_71  gpio.PinIO = bcm283x.GPIO21 // GPIO21
	SO_72  gpio.PinIO = bcm283x.GPIO41 // GPIO41
	SO_73  pin.Pin    = pin.GROUND     // GND
	SO_74  pin.Pin    = pin.GROUND     // GND
	SO_75  gpio.PinIO = bcm283x.GPIO22 // GPIO22
	SO_76  gpio.PinIO = bcm283x.GPIO42 // GPIO42
	SO_77  gpio.PinIO = bcm283x.GPIO23 // GPIO23
	SO_78  gpio.PinIO = bcm283x.GPIO43 // GPIO43
	SO_79  pin.Pin    = pin.GROUND     // GND
	SO_80  pin.Pin    = pin.GROUND     // GND
	SO_81  gpio.PinIO = bcm283x.GPIO24 // GPIO24
	SO_82  gpio.PinIO = bcm283x.GPIO44 // GPIO44
	SO_83  gpio.PinIO = bcm283x.GPIO25 // GPIO25
	SO_84  gpio.PinIO = bcm283x.GPIO45 // GPIO45
	SO_85  pin.Pin    = pin.GROUND     // GND
	SO_86  pin.Pin    = pin.GROUND     // GND
	SO_87  gpio.PinIO = bcm283x.GPIO26 // GPIO26
	SO_88  pin.Pin    = pin.INVALID    // HDMI_HPD_N_1V8, HDMI_HPD_N_1V8, GPIO46_1V8
	SO_89  gpio.PinIO = bcm283x.GPIO27 // GPIO27
	SO_90  pin.Pin    = pin.INVALID    // EMMC_EN_N_1V8, EMMC_EN_N_1V8, GPIO47_1V8
	SO_91  pin.Pin    = pin.GROUND     // GND
	SO_92  pin.Pin    = pin.GROUND     // GND
	SO_93  pin.Pin    = pin.INVALID    // DSI0_DN1
	SO_94  pin.Pin    = pin.INVALID    // DSI1_DP0
	SO_95  pin.Pin    = pin.INVALID    // DSI0_DP1
	SO_96  pin.Pin    = pin.INVALID    // DSI1_DN0
	SO_97  pin.Pin    = pin.GROUND     // GND
	SO_98  pin.Pin    = pin.GROUND     // GND
	SO_99  pin.Pin    = pin.INVALID    // DSI0_DN0
	SO_100 pin.Pin    = pin.INVALID    // DSI1_CP
	SO_101 pin.Pin    = pin.INVALID    // DSI0_DP0
	SO_102 pin.Pin    = pin.INVALID    // DSI1_CN
	SO_103 pin.Pin    = pin.GROUND     // GND
	SO_104 pin.Pin    = pin.GROUND     // GND
	SO_105 pin.Pin    = pin.INVALID    // DSI0_CN
	SO_106 pin.Pin    = pin.INVALID    // DSI1_DP3
	SO_107 pin.Pin    = pin.INVALID    // DSI0_CP
	SO_108 pin.Pin    = pin.INVALID    // DSI1_DN3
	SO_109 pin.Pin    = pin.GROUND     // GND
	SO_110 pin.Pin    = pin.GROUND     // GND
	SO_111 pin.Pin    = pin.INVALID    // HDMI_CLK_N
	SO_112 pin.Pin    = pin.INVALID    // DSI1_DP2
	SO_113 pin.Pin    = pin.INVALID    // HDMI_CLK_P
	SO_114 pin.Pin    = pin.INVALID    // DSI1_DN2
	SO_115 pin.Pin    = pin.GROUND     // GND
	SO_116 pin.Pin    = pin.GROUND     // GND
	SO_117 pin.Pin    = pin.INVALID    // HDMI_D0_N
	SO_118 pin.Pin    = pin.INVALID    // DSI1_DP1
	SO_119 pin.Pin    = pin.INVALID    // HDMI_D0_P
	SO_120 pin.Pin    = pin.INVALID    // DSI1_DN1
	SO_121 pin.Pin    = pin.GROUND     // GND
	SO_122 pin.Pin    = pin.GROUND     // GND
	SO_123 pin.Pin    = pin.INVALID    // HDMI_D1_N
	SO_124 pin.Pin    = pin.INVALID    // NC
	SO_125 pin.Pin    = pin.INVALID    // HDMI_D1_P
	SO_126 pin.Pin    = pin.INVALID    // NC
	SO_127 pin.Pin    = pin.GROUND     // GND
	SO_128 pin.Pin    = pin.INVALID    // NC
	SO_129 pin.Pin    = pin.INVALID    // HDMI_D2_N
	SO_130 pin.Pin    = pin.INVALID    // NC
	SO_131 pin.Pin    = pin.INVALID    // HDMI_D2_P
	SO_132 pin.Pin    = pin.INVALID    // NC
	SO_133 pin.Pin    = pin.GROUND     // GND
	SO_134 pin.Pin    = pin.GROUND     // GND
	SO_135 pin.Pin    = pin.INVALID    // CAM1_DP3
	SO_136 pin.Pin    = pin.INVALID    // CAM0_DP0
	SO_137 pin.Pin    = pin.INVALID    // CAM1_DN3
	SO_138 pin.Pin    = pin.INVALID    // CAM0_DN0
	SO_139 pin.Pin    = pin.GROUND     // GND
	SO_140 pin.Pin    = pin.GROUND     // GND
	SO_141 pin.Pin    = pin.INVALID    // CAM1_DP2
	SO_142 pin.Pin    = pin.INVALID    // CAM0_CP
	SO_143 pin.Pin    = pin.INVALID    // CAM1_DN2
	SO_144 pin.Pin    = pin.INVALID    // CAM0_CN
	SO_145 pin.Pin    = pin.GROUND     // GND
	SO_146 pin.Pin    = pin.GROUND     // GND
	SO_147 pin.Pin    = pin.INVALID    // CAM1_CP
	SO_148 pin.Pin    = pin.INVALID    // CAM0_DP1
	SO_149 pin.Pin    = pin.INVALID    // CAM1_CN
	SO_150 pin.Pin    = pin.INVALID    // CAM0_DN1
	SO_151 pin.Pin    = pin.GROUND     // GND
	SO_152 pin.Pin    = pin.GROUND     // GND
	SO_153 pin.Pin    = pin.INVALID    // CAM1_DP1
	SO_154 pin.Pin    = pin.INVALID    // NC
	SO_155 pin.Pin    = pin.INVALID    // CAM1_DN1
	SO_156 pin.Pin    = pin.INVALID    // NC
	SO_157 pin.Pin    = pin.GROUND     // GND
	SO_158 pin.Pin    = pin.INVALID    // NC
	SO_159 pin.Pin    = pin.INVALID    // CAM1_DP0
	SO_160 pin.Pin    = pin.INVALID    // NC
	SO_161 pin.Pin    = pin.INVALID    // CAM1_DN0
	SO_162 pin.Pin    = pin.INVALID    // NC
	SO_163 pin.Pin    = pin.GROUND     // GND
	SO_164 pin.Pin    = pin.GROUND     // GND
	SO_165 pin.Pin    = pin.INVALID    // USB_DP
	SO_166 pin.Pin    = pin.INVALID    // TVDAC
	SO_167 pin.Pin    = pin.INVALID    // USB_DM
	SO_168 pin.Pin    = pin.INVALID    // USB_OTGID
	SO_169 pin.Pin    = pin.GROUND     // GND
	SO_170 pin.Pin    = pin.GROUND     // GND
	SO_171 pin.Pin    = pin.INVALID    // HDMI_CEC
	SO_172 pin.Pin    = pin.INVALID    // VC_TRST_N
	SO_173 pin.Pin    = pin.INVALID    // HDMI_SDA
	SO_174 pin.Pin    = pin.INVALID    // VC_TDI
	SO_175 pin.Pin    = pin.INVALID    // HDMI_SCL
	SO_176 pin.Pin    = pin.INVALID    // VC_TMS
	SO_177 pin.Pin    = pin.INVALID    // RUN
	SO_178 pin.Pin    = pin.INVALID    // VC_TDO
	SO_179 pin.Pin    = pin.INVALID    // VDD_CORE (DO NOT CONNECT)
	SO_180 pin.Pin    = pin.INVALID    // VC_TCK
	SO_181 pin.Pin    = pin.GROUND     // GND
	SO_182 pin.Pin    = pin.GROUND     // GND
	SO_183 pin.Pin    = pin.V1_8       // 1V8
	SO_184 pin.Pin    = pin.V1_8       // 1V8
	SO_185 pin.Pin    = pin.V1_8       // 1V8
	SO_186 pin.Pin    = pin.V1_8       // 1V8
	SO_187 pin.Pin    = pin.GROUND     // GND
	SO_188 pin.Pin    = pin.GROUND     // GND
	SO_189 pin.Pin    = pin.DC_IN      // VDAC
	SO_190 pin.Pin    = pin.DC_IN      // VDAC
	SO_191 pin.Pin    = pin.V3_3       // 3V3
	SO_192 pin.Pin    = pin.V3_3       // 3V3
	SO_193 pin.Pin    = pin.V3_3       // 3V3
	SO_194 pin.Pin    = pin.V3_3       // 3V3
	SO_195 pin.Pin    = pin.GROUND     // GND
	SO_196 pin.Pin    = pin.GROUND     // GND
	SO_197 pin.Pin    = pin.DC_IN      // VBAT
	SO_198 pin.Pin    = pin.DC_IN      // VBAT
	SO_199 pin.Pin    = pin.DC_IN      // VBAT
	SO_200 pin.Pin    = pin.DC_IN      // VBAT
)

//

// revisionCode processes the CPU revision code based on documentation at
// https://www.raspberrypi.org/documentation/hardware/raspberrypi/revision-codes/README.md
//
//    Format is: uuuuuuuuFMMMCCCCPPPPTTTTTTTTRRRR
//                       2  2   1   1       0   0
//                       3  0   6   2       4   0
type revisionCode uint32

// parseRevision processes the old style revision codes to new style bitpacked
// format.
func parseRevision(v uint32) (revisionCode, error) {
	w := revisionCode(v) & warrantyVoid
	switch v &^ uint32(warrantyVoid) {
	case 0x2, 0x3:
		return w | newFormat | memory256MB | egoman | bcm2835 | board1B, nil
	case 0x4:
		return w | newFormat | memory256MB | sonyUK | bcm2835 | board1B | 2, nil // v2.0
	case 0x5:
		return w | newFormat | memory256MB | bcm2835 | board1B | 2, nil // v2.0 Qisda
	case 0x6:
		return w | newFormat | memory256MB | egoman | bcm2835 | board1B | 2, nil // v2.0
	case 0x7:
		return w | newFormat | memory256MB | egoman | bcm2835 | board1A | 2, nil // v2.0
	case 0x8:
		return w | newFormat | memory256MB | sonyUK | bcm2835 | board1A | 2, nil // v2.0
	case 0x9:
		return w | newFormat | memory256MB | bcm2835 | board1A | 2, nil // v2.0 Qisda
	case 0xd:
		return w | newFormat | memory512MB | egoman | bcm2835 | board1B | 2, nil // v2.0
	case 0xe:
		return w | newFormat | memory512MB | sonyUK | bcm2835 | board1B | 2, nil // v2.0
	case 0xf:
		return w | newFormat | memory512MB | egoman | bcm2835 | board1B | 2, nil // v2.0
	case 0x10:
		return w | newFormat | memory512MB | sonyUK | bcm2835 | board1BPlus | 2, nil // v1.2
	case 0x11:
		return w | newFormat | memory512MB | sonyUK | bcm2835 | boardCM1, nil
	case 0x12:
		return w | newFormat | memory256MB | sonyUK | bcm2835 | board1APlus | 1, nil // v1.1
	case 0x13:
		return w | newFormat | memory512MB | embest | bcm2835 | board1BPlus | 2, nil // v1.2
	case 0x14:
		return w | newFormat | memory512MB | embest | bcm2835 | boardCM1, nil
	case 0x15:
		// Can be either 256MB or 512MB.
		return w | newFormat | memory256MB | embest | bcm2835 | board1APlus | 1, nil // v1.1
	default:
		if v&uint32(newFormat) == 0 {
			return 0, fmt.Errorf("rpi: unknown hardware version: 0x%x", v)
		}
		return revisionCode(v), nil
	}
}

const (
	warrantyVoid      revisionCode = 1 << 24
	newFormat         revisionCode = 1 << 23
	memoryShift                    = 20
	memoryMask        revisionCode = 0x7 << memoryShift
	manufacturerShift              = 16
	manufacturerMask  revisionCode = 0xf << manufacturerShift
	processorShift                 = 12
	processorMask     revisionCode = 0xf << processorShift
	boardShift                     = 4
	boardMask         revisionCode = 0xff << boardShift
	revisionMask      revisionCode = 0xf << 0

	memory256MB revisionCode = 0 << memoryShift
	memory512MB revisionCode = 1 << memoryShift
	memory1GB   revisionCode = 2 << memoryShift
	memory2GB   revisionCode = 3 << memoryShift
	memory4GB   revisionCode = 4 << memoryShift

	sonyUK    revisionCode = 0 << manufacturerShift
	egoman    revisionCode = 1 << manufacturerShift
	embest    revisionCode = 2 << manufacturerShift
	sonyJapan revisionCode = 3 << manufacturerShift
	embest2   revisionCode = 4 << manufacturerShift
	stadium   revisionCode = 5 << manufacturerShift

	bcm2835 revisionCode = 0 << processorShift
	bcm2836 revisionCode = 1 << processorShift
	bcm2837 revisionCode = 2 << processorShift
	bcm2711 revisionCode = 3 << processorShift

	board1A       revisionCode = 0x0 << boardShift
	board1B       revisionCode = 0x1 << boardShift
	board1APlus   revisionCode = 0x2 << boardShift
	board1BPlus   revisionCode = 0x3 << boardShift
	board2B       revisionCode = 0x4 << boardShift
	boardAlpha    revisionCode = 0x5 << boardShift
	boardCM1      revisionCode = 0x6 << boardShift
	board3B       revisionCode = 0x8 << boardShift
	boardZero     revisionCode = 0x9 << boardShift
	boardCM3      revisionCode = 0xa << boardShift
	boardZeroW    revisionCode = 0xc << boardShift
	board3BPlus   revisionCode = 0xd << boardShift
	board3APlus   revisionCode = 0xe << boardShift
	boardReserved revisionCode = 0xf << boardShift
	boardCM3Plus  revisionCode = 0x10 << boardShift
	board4B       revisionCode = 0x11 << boardShift
)

// features represents the different features on various Raspberry Pi boards.
//
// See https://github.com/raspberrypi/firmware/blob/master/extra/dt-blob.dts
// for the official mapping.
type features struct {
	hdrP1P26    bool // P1 has 26 pins
	hdrP1P40    bool // P1 has 40 pins
	hdrP5       bool // P5 is present
	hdrAudio    bool // Audio header is present
	audioLeft41 bool // AUDIO_LEFT uses GPIO41 (RPi3 and later) instead of GPIO45 (old boards)
	hdrHDMI     bool // At least one HDMI port is present
	hdrSODIMM   bool // SODIMM port is present
}

func (f *features) init(v uint32) error {
	r, err := parseRevision(v)
	if err != nil {
		return err
	}
	// Ignore the overclock bit.
	r &^= warrantyVoid
	switch r & boardMask {
	case board1A:
		f.hdrP1P26 = true
		f.hdrAudio = true
		// Only the v2 PCB has the P5 header.
		if r&revisionMask == 2 {
			f.hdrP5 = true
			f.hdrHDMI = true
		}
	case board1B:
		f.hdrP1P26 = true
		f.hdrAudio = true
		// Only the v2 PCB has the P5 header.
		if r&revisionMask == 2 {
			f.hdrP5 = true
			f.hdrHDMI = true
		}
	case board1APlus:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.hdrHDMI = true
	case board1BPlus:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.hdrHDMI = true
	case board2B:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.hdrHDMI = true
	case boardAlpha:
	case boardCM1:
		// TODO: define CM1 SODIMM header if anyone ever needs it. Please file an
		// issue at https://github.com/google/periph/issues/new/choose
	case board3B:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.audioLeft41 = true
		f.hdrHDMI = true
	case boardZero:
		f.hdrP1P40 = true
		f.hdrHDMI = true
	case boardCM3:
		// Tell CM3 and CM3-Lite apart, if possible.
		f.hdrSODIMM = true
	case boardZeroW:
		f.hdrP1P40 = true
		f.hdrHDMI = true
	case board3BPlus:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.audioLeft41 = true
		f.hdrHDMI = true
	case board3APlus:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.audioLeft41 = true
		f.hdrHDMI = true
	case boardReserved:
	case boardCM3Plus:
		// Tell CM3 and CM3-Lite apart, if possible.
		f.hdrSODIMM = true
	case board4B:
		f.hdrP1P40 = true
		f.hdrAudio = true
		f.audioLeft41 = true
		f.hdrHDMI = true
	default:
		return fmt.Errorf("rpi: unknown hardware version: 0x%x", r)
	}
	return nil
}

// registerHeaders registers the headers for this board and fixes the GPIO
// global variables.
func (f *features) registerHeaders() error {
	if f.hdrP1P26 {
		if err := pinreg.Register("P1", [][]pin.Pin{
			{P1_1, P1_2},
			{P1_3, P1_4},
			{P1_5, P1_6},
			{P1_7, P1_8},
			{P1_9, P1_10},
			{P1_11, P1_12},
			{P1_13, P1_14},
			{P1_15, P1_16},
			{P1_17, P1_18},
			{P1_19, P1_20},
			{P1_21, P1_22},
			{P1_23, P1_24},
			{P1_25, P1_26},
		}); err != nil {
			return err
		}

		// TODO(maruel): Models from 2012 and earlier have P1_3=GPIO0, P1_5=GPIO1 and P1_13=GPIO21.
		// P2 and P3 are not useful.
		// P6 has a RUN pin for reset but it's not available after Pi version 1.
		P1_27 = gpio.INVALID
		P1_28 = gpio.INVALID
		P1_29 = gpio.INVALID
		P1_30 = pin.INVALID
		P1_31 = gpio.INVALID
		P1_32 = gpio.INVALID
		P1_33 = gpio.INVALID
		P1_34 = pin.INVALID
		P1_35 = gpio.INVALID
		P1_36 = gpio.INVALID
		P1_37 = gpio.INVALID
		P1_38 = gpio.INVALID
		P1_39 = pin.INVALID
		P1_40 = gpio.INVALID
	} else if f.hdrP1P40 {
		if err := pinreg.Register("P1", [][]pin.Pin{
			{P1_1, P1_2},
			{P1_3, P1_4},
			{P1_5, P1_6},
			{P1_7, P1_8},
			{P1_9, P1_10},
			{P1_11, P1_12},
			{P1_13, P1_14},
			{P1_15, P1_16},
			{P1_17, P1_18},
			{P1_19, P1_20},
			{P1_21, P1_22},
			{P1_23, P1_24},
			{P1_25, P1_26},
			{P1_27, P1_28},
			{P1_29, P1_30},
			{P1_31, P1_32},
			{P1_33, P1_34},
			{P1_35, P1_36},
			{P1_37, P1_38},
			{P1_39, P1_40},
		}); err != nil {
			return err
		}
	} else {
		P1_1 = pin.INVALID
		P1_2 = pin.INVALID
		P1_3 = gpio.INVALID
		P1_4 = pin.INVALID
		P1_5 = gpio.INVALID
		P1_6 = pin.INVALID
		P1_7 = gpio.INVALID
		P1_8 = gpio.INVALID
		P1_9 = pin.INVALID
		P1_10 = gpio.INVALID
		P1_11 = gpio.INVALID
		P1_12 = gpio.INVALID
		P1_13 = gpio.INVALID
		P1_14 = pin.INVALID
		P1_15 = gpio.INVALID
		P1_16 = gpio.INVALID
		P1_17 = pin.INVALID
		P1_18 = gpio.INVALID
		P1_19 = gpio.INVALID
		P1_20 = pin.INVALID
		P1_21 = gpio.INVALID
		P1_22 = gpio.INVALID
		P1_23 = gpio.INVALID
		P1_24 = gpio.INVALID
		P1_25 = pin.INVALID
		P1_26 = gpio.INVALID
		P1_27 = gpio.INVALID
		P1_28 = gpio.INVALID
		P1_29 = gpio.INVALID
		P1_30 = pin.INVALID
		P1_31 = gpio.INVALID
		P1_32 = gpio.INVALID
		P1_33 = gpio.INVALID
		P1_34 = pin.INVALID
		P1_35 = gpio.INVALID
		P1_36 = gpio.INVALID
		P1_37 = gpio.INVALID
		P1_38 = gpio.INVALID
		P1_39 = pin.INVALID
		P1_40 = gpio.INVALID
	}

	// Only the A and B v2 PCB has the P5 header.
	if f.hdrP5 {
		if err := pinreg.Register("P5", [][]pin.Pin{
			{P5_1, P5_2},
			{P5_3, P5_4},
			{P5_5, P5_6},
			{P5_7, P5_8},
		}); err != nil {
			return err
		}
	} else {
		P5_1 = pin.INVALID
		P5_2 = pin.INVALID
		P5_3 = gpio.INVALID
		P5_4 = gpio.INVALID
		P5_5 = gpio.INVALID
		P5_6 = gpio.INVALID
		P5_7 = pin.INVALID
		P5_8 = pin.INVALID
	}

	if f.hdrSODIMM {
		if err := pinreg.Register("SO", [][]pin.Pin{
			{SO_1, SO_2},
			{SO_3, SO_4},
			{SO_5, SO_6},
			{SO_7, SO_8},
			{SO_9, SO_10},
			{SO_11, SO_12},
			{SO_13, SO_14},
			{SO_15, SO_16},
			{SO_17, SO_18},
			{SO_19, SO_20},
			{SO_21, SO_22},
			{SO_23, SO_24},
			{SO_25, SO_26},
			{SO_27, SO_28},
			{SO_29, SO_30},
			{SO_31, SO_32},
			{SO_33, SO_34},
			{SO_35, SO_36},
			{SO_37, SO_38},
			{SO_39, SO_40},
			{SO_41, SO_42},
			{SO_43, SO_44},
			{SO_45, SO_46},
			{SO_47, SO_48},
			{SO_49, SO_50},
			{SO_51, SO_52},
			{SO_53, SO_54},
			{SO_55, SO_56},
			{SO_57, SO_58},
			{SO_59, SO_60},
			{SO_61, SO_62},
			{SO_63, SO_64},
			{SO_65, SO_66},
			{SO_67, SO_68},
			{SO_69, SO_70},
			{SO_71, SO_72},
			{SO_73, SO_74},
			{SO_75, SO_76},
			{SO_77, SO_78},
			{SO_79, SO_80},
			{SO_81, SO_82},
			{SO_83, SO_84},
			{SO_85, SO_86},
			{SO_87, SO_88},
			{SO_89, SO_90},
			{SO_91, SO_92},
			{SO_93, SO_94},
			{SO_95, SO_96},
			{SO_97, SO_98},
			{SO_99, SO_100},
			{SO_101, SO_102},
			{SO_103, SO_104},
			{SO_105, SO_106},
			{SO_107, SO_108},
			{SO_109, SO_110},
			{SO_111, SO_112},
			{SO_113, SO_114},
			{SO_115, SO_116},
			{SO_117, SO_118},
			{SO_119, SO_120},
			{SO_121, SO_122},
			{SO_123, SO_124},
			{SO_125, SO_126},
			{SO_127, SO_128},
			{SO_129, SO_130},
			{SO_131, SO_132},
			{SO_133, SO_134},
			{SO_135, SO_136},
			{SO_137, SO_138},
			{SO_139, SO_140},
			{SO_141, SO_142},
			{SO_143, SO_144},
			{SO_145, SO_146},
			{SO_147, SO_148},
			{SO_149, SO_150},
			{SO_151, SO_152},
			{SO_153, SO_154},
			{SO_155, SO_156},
			{SO_157, SO_158},
			{SO_159, SO_160},
			{SO_161, SO_162},
			{SO_163, SO_164},
			{SO_165, SO_166},
			{SO_167, SO_168},
			{SO_169, SO_170},
			{SO_171, SO_172},
			{SO_173, SO_174},
			{SO_175, SO_176},
			{SO_177, SO_178},
			{SO_179, SO_180},
			{SO_181, SO_182},
			{SO_183, SO_184},
			{SO_185, SO_186},
			{SO_187, SO_188},
			{SO_189, SO_190},
			{SO_191, SO_192},
			{SO_193, SO_194},
			{SO_195, SO_196},
			{SO_197, SO_198},
			{SO_199, SO_200},
		}); err != nil {
			return err
		}
	}

	if f.hdrAudio {
		// Two early versions of RPi1 had left and right reversed but we don't
		// bother handling this here.
		// https://github.com/raspberrypi/firmware/blob/master/extra/dt-blob.dts
		if !f.audioLeft41 {
			AUDIO_LEFT = bcm283x.GPIO45 // PWM1 for older boards
		}
		if err := pinreg.Register("AUDIO", [][]pin.Pin{
			{AUDIO_LEFT},
			{AUDIO_RIGHT},
		}); err != nil {
			return err
		}
	}

	if f.hdrHDMI {
		if err := pinreg.Register("HDMI", [][]pin.Pin{{HDMI_HOTPLUG_DETECT}}); err != nil {
			return err
		}
	}
	return nil
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "rpi"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return []string{"bcm283x-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("board Raspberry Pi not detected")
	}

	// Setup headers based on board revision.
	//
	// This code is not futureproof, it will error out on a Raspberry Pi 4
	// whenever it comes out.
	// Revision codes from: http://elinux.org/RPi_HardwareHistory
	f := features{}
	rev := distro.DTRevision()
	if rev == 0 {
		return true, fmt.Errorf("rpi: failed to obtain revision")
	}
	if err := f.init(rev); err != nil {
		return true, err
	}

	return true, f.registerHeaders()
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
