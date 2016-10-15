// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package apa102

import (
	"errors"
	"image"
	"image/color"

	"github.com/google/pio/conn/spi"
	"github.com/google/pio/devices"
	"github.com/maruel/temperature"
)

// maxOut is the maximum intensity of each channel on a APA102 LED.
const maxOut = 0x1EE1

// ramp converts input from [0, 0xFF] as intensity to lightness on a scale of
// [0, maxOut] or other desired range [0, max].
//
// It tries to use the same curve independent of the scale used. max can be
// changed to change the color temperature or to limit power dissipation.
//
// It's the reverse of lightness; https://en.wikipedia.org/wiki/Lightness
func ramp(l uint8, max uint16) uint16 {
	if l == 0 {
		// Make sure black is black.
		return 0
	}
	// linearCutOff defines the linear section of the curve. Inputs between
	// [0, linearCutOff] are mapped linearly to the output. It is 1% of maximum
	// output.
	linearCutOff := uint32((max + 50) / 100)
	l32 := uint32(l)
	if l32 < linearCutOff {
		return uint16(l32)
	}

	// Maps [linearCutOff, 255] to use [linearCutOff*max/255, max] using a x^3
	// ramp.
	// Realign input to [0, 255-linearCutOff]. It now maps to
	// [0, max-linearCutOff*max/255].
	//const inRange = 255
	l32 -= linearCutOff
	inRange := 255 - linearCutOff
	outRange := uint32(max) - linearCutOff
	offset := inRange >> 1
	y := (l32*l32*l32 + offset) / inRange
	return uint16((y*outRange+(offset*offset))/inRange/inRange + linearCutOff)
}

// lut is a lookup table that initializes itself on the fly.
type lut struct {
	intensity   uint8  // Set an intensity between 0 (off) and 255 (full brightness).
	temperature uint16 // In Kelvin.
	r           [256]uint16
	g           [256]uint16
	b           [256]uint16
}

func (l *lut) init(i uint8, t uint16) {
	if i != l.intensity || t != l.temperature {
		l.intensity = i
		l.temperature = t
		tr, tg, tb := temperature.ToRGB(l.temperature)
		maxR := uint16((uint32(maxOut)*uint32(l.intensity)*uint32(tr) + 127*127) / 65025)
		maxG := uint16((uint32(maxOut)*uint32(l.intensity)*uint32(tg) + 127*127) / 65025)
		maxB := uint16((uint32(maxOut)*uint32(l.intensity)*uint32(tb) + 127*127) / 65025)
		for i := range l.r {
			l.r[i] = ramp(uint8(i), maxR)
		}
		if maxG == maxR {
			copy(l.g[:], l.r[:])
		} else {
			for i := range l.g {
				l.g[i] = ramp(uint8(i), maxG)
			}
		}
		if maxB == maxR {
			copy(l.b[:], l.r[:])
		} else if maxB == maxG {
			copy(l.b[:], l.g[:])
		} else {
			for i := range l.b {
				l.b[i] = ramp(uint8(i), maxB)
			}
		}
	}
}

// raster serializes converts a buffer of RGB bytes to the APA102 SPI format.
//
// It is expected to be given the part where pixels are, not the header nor
// footer.
//
// dst is in APA102 SPI 32 bits word format. src is in RGB 24 bits word format.
// maxR, maxG and maxB are the maximum light intensity to use per channel.
func (l *lut) raster(dst []byte, src []byte) {
	// Whichever is the shortest.
	length := len(src) / 3
	if o := len(dst) / 4; o < length {
		length = o
	}
	for i := 0; i < length; i++ {
		// Converts a color into the 4 bytes needed to control an APA-102 LED.
		//
		// The response as seen by the human eye is very non-linear. The APA-102
		// provides an overall brightness PWM but it is relatively slower and
		// results in human visible flicker. On the other hand the minimal color
		// (1/255) is still too intense at full brightness, so for very dark color,
		// it is worth using the overall brightness PWM. The goal is to use
		// brightness!=31 as little as possible.
		//
		// Global brightness frequency is 580Hz and color frequency at 19.2kHz.
		// https://cpldcpu.wordpress.com/2014/08/27/apa102/
		// Both are multiplicative, so brightness@50% and color@50% means an
		// effective 25% duty cycle but it is not properly distributed, which is
		// the main problem.
		//
		// It is unclear to me if brightness is exactly in 1/31 increment as I don't
		// have an oscilloscope to confirm. Same for color in 1/255 increment.
		// TODO(maruel): I have one now!
		//
		// Each channel duty cycle ramps from 100% to 1/(31*255) == 1/7905.
		//
		// Computes brighness, blue, green, red.
		j := 3 * i
		r := l.r[src[j]]
		g := l.g[src[j+1]]
		b := l.b[src[j+2]]
		m := r | g | b
		j += i
		if m <= 1023 {
			if m <= 255 {
				dst[j], dst[j+1], dst[j+2], dst[j+3] = byte(0xE0+1), byte(b), byte(g), byte(r)
			} else if m <= 511 {
				dst[j], dst[j+1], dst[j+2], dst[j+3] = byte(0xE0+2), byte(b>>1), byte(g>>1), byte(r>>1)
			} else {
				dst[j], dst[j+1], dst[j+2], dst[j+3] = byte(0xE0+4), byte((b+2)>>2), byte((g+2)>>2), byte((r+2)>>2)
			}
		} else {
			// In this case we need to use a ramp of 255-1 even for lower colors.
			dst[j], dst[j+1], dst[j+2], dst[j+3] = byte(0xE0+31), byte((b+15)/31), byte((g+15)/31), byte((r+15)/31)
		}
	}
}

// rasterImg is the generic version of raster.
func (l *lut) rasterImg(dst []byte, r image.Rectangle, src image.Image, srcR image.Rectangle) {
	// Render directly into the buffer for maximum performance and to keep
	// untouched sections intact.
	deltaX4 := 4 * (r.Min.X - srcR.Min.X)
	if img, ok := src.(*image.NRGBA); ok {
		// Fast path for image.NRGBA.
		pix := img.Pix[srcR.Min.Y*img.Stride:]
		for sX := srcR.Min.X; sX < srcR.Max.X; sX++ {
			sX4 := 4 * sX
			r := l.r[pix[sX4]]
			g := l.g[pix[sX4+1]]
			b := l.b[pix[sX4+2]]
			m := r | g | b
			rX := sX4 + deltaX4
			if m <= 1023 {
				if m <= 255 {
					dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+1), byte(b), byte(g), byte(r)
				} else if m <= 511 {
					dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+2), byte(b>>1), byte(g>>1), byte(r>>1)
				} else {
					dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+4), byte((b+2)>>2), byte((g+2)>>2), byte((r+2)>>2)
				}
			} else {
				// In this case we need to use a ramp of 255-1 even for lower colors.
				dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+31), byte((b+15)/31), byte((g+15)/31), byte((r+15)/31)
			}
		}
	} else {
		// Generic version.
		for sX := srcR.Min.X; sX < srcR.Max.X; sX++ {
			r16, g16, b16, _ := src.At(sX, srcR.Min.Y).RGBA()
			r := l.r[byte(r16>>8)]
			g := l.g[byte(g16>>8)]
			b := l.b[byte(b16>>8)]
			m := r | g | b
			rX := sX*4 + deltaX4
			if m <= 1023 {
				if m <= 255 {
					dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+1), byte(b), byte(g), byte(r)
				} else if m <= 511 {
					dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+2), byte(b>>1), byte(g>>1), byte(r>>1)
				} else {
					dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+4), byte((b+2)>>2), byte((g+2)>>2), byte((r+2)>>2)
				}
			} else {
				// In this case we need to use a ramp of 255-1 even for lower colors.
				dst[rX], dst[rX+1], dst[rX+2], dst[rX+3] = byte(0xE0+31), byte((b+15)/31), byte((g+15)/31), byte((r+15)/31)
			}
		}
	}
}

// ToRGB converts a slice of color.NRGBA to a byte stream of RGB pixels.
//
// Ignores alpha.
func ToRGB(p []color.NRGBA) []byte {
	b := make([]byte, 0, len(p)*3)
	for _, c := range p {
		b = append(b, c.R, c.G, c.B)
	}
	return b
}

// Dev represents a strip of APA-102 LEDs as a strip connected over a SPI bus.
// It accepts a stream of raw RGB pixels and converts it to the full dynamic
// range as supported by APA102 protocol (nearly 8000:1 contrast ratio).
//
// Includes intensity and temperature correction.
type Dev struct {
	Intensity   uint8  // Set an intensity between 0 (off) and 255 (full brightness).
	Temperature uint16 // In Kelvin.
	s           spi.Conn
	l           lut // Updated at each .Write() call.
	numLights   int
	buf         []byte
}

// ColorModel implements devices.Display. There's no surprise, it is
// color.NRGBAModel.
func (d *Dev) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements devices.Display. Min is guaranteed to be {0, 0}.
func (d *Dev) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: d.numLights, Y: 1}}
}

// Draw implements devices.Display.
//
// Using something else than image.NRGBA is 10x slower. When using image.NRGBA,
// the alpha channel is ignored.
func (d *Dev) Draw(r image.Rectangle, src image.Image, sp image.Point) {
	r = r.Intersect(d.Bounds())
	srcR := src.Bounds()
	srcR.Min = srcR.Min.Add(sp)
	if dX := r.Dx(); dX < srcR.Dx() {
		srcR.Max.X = srcR.Min.X + dX
	}
	if dY := r.Dy(); dY < srcR.Dy() {
		srcR.Max.Y = srcR.Min.Y + dY
	}
	d.l.init(d.Intensity, d.Temperature)
	d.l.rasterImg(d.buf[4:4+4*d.numLights], r, src, srcR)
	_, _ = d.s.Write(d.buf)
}

// Write accepts a stream of raw RGB pixels and sends it as APA102 encoded
// stream.
func (d *Dev) Write(pixels []byte) (int, error) {
	if len(pixels)%3 != 0 {
		return 0, errLength
	}
	d.l.init(d.Intensity, d.Temperature)
	d.l.raster(d.buf[4:4+4*d.numLights], pixels)
	_, err := d.s.Write(d.buf)
	return len(pixels), err
}

// New returns a strip that communicates over SPI to APA102 LEDs.
//
// The SPI bus speed should be high, at least in the Mhz range, as
// there's 32 bits sent per LED, creating a staggered effect. See
// https://cpldcpu.wordpress.com/2014/11/30/understanding-the-apa102-superled/
//
// Temperature is in °Kelvin and a reasonable default value is 6500°K.
//
// As per APA102-C spec, the chip's max refresh rate is 400hz.
// https://en.wikipedia.org/wiki/Flicker_fusion_threshold is a recommended
// reading.
func New(s spi.Conn, numLights int, intensity uint8, temperature uint16) (*Dev, error) {
	if err := s.Configure(spi.Mode3, 8); err != nil {
		return nil, err
	}
	// End frames are needed to be able to push enough SPI clock signals due to
	// internal half-delay of data signal from each individual LED. See
	// https://cpldcpu.wordpress.com/2014/11/30/understanding-the-apa102-superled/
	buf := make([]byte, 4*(numLights+1)+numLights/2/8+1)
	tail := buf[4+4*numLights:]
	for i := range tail {
		tail[i] = 0xFF
	}
	return &Dev{
		Intensity:   intensity,
		Temperature: temperature,
		s:           s,
		numLights:   numLights,
		buf:         buf,
	}, nil
}

//

var errLength = errors.New("invalid RGB stream length")

var _ devices.Display = &Dev{}
