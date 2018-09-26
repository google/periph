package mt7688

import "errors"

// driverGPIO implements periph.Driver.
type driverGPIO struct {
	// gpioMemory is the memory map of the CPU GPIO registers.
	gpioMemory *gpioMap
}

func (d driverGPIO) String() string {
	return "mt7688-gpio"
}

func (d driverGPIO) Prerequisites() []string {
	return nil
}

func (d driverGPIO) After() []string {
	return []string{"sysfs-gpio"}
}

func (d *driverGPIO) Init() (bool, error) {
	if !Present() {
		return false, errors.New("mt7688 board not detected")
	}

	// Initialize pins
	initPins()

	return true, nil
}

var drvGPIO driverGPIO
