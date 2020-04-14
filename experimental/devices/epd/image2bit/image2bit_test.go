package image2bit

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func Test_getOffset(t *testing.T) {
	tb := BitPlane{
		Rect: image.Rect(0, 0, 16, 16),
	}

	for _, test := range []struct {
		name string
		x, y int

		byteIndex, bitIndex int
		mask                byte
	}{
		{
			name: "bit order, first, edge",

			x: 0, y: 0,

			byteIndex: 0, bitIndex: 7, mask: 0b01111111,
		},
		{
			name: "bit order 2",

			x: 1, y: 0,

			byteIndex: 0, bitIndex: 6, mask: 0b10111111,
		},
		{
			name: "bit order, last, edge",

			x: 7, y: 0,

			byteIndex: 0, bitIndex: 0, mask: 0b11111110,
		},
		{
			name: "byte index",

			x: 1 + 8, y: 0,

			byteIndex: 1, bitIndex: 6, mask: 0b10111111,
		},
		{
			name: "byte index + row",
			x:    1 + 8,
			y:    1,

			byteIndex: 16/8 + 1, bitIndex: 6, mask: 0b10111111,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			byteIndex, bitIndex, mask := tb.getOffset(test.x, test.y)
			if byteIndex != test.byteIndex || bitIndex != test.bitIndex || mask != test.mask {
				t.Errorf("getOffset(%d,%d) failed: Got (%v, %v, %02x), expected (%v, %v, %02x)",
					test.x, test.y,
					byteIndex, bitIndex, mask,
					test.byteIndex, test.bitIndex, test.mask)
			}
		})
	}
}

func TestAllWhite(t *testing.T) {
	// Default color is black. Test that setting everything to White
	// touches all bits
	tb := NewBitPlane(image.Rect(0, 0, 16, 2))
	for y := 0; y < tb.Rect.Dy(); y++ {
		for x := 0; x < tb.Rect.Dx(); x++ {
			tb.Set(x, y, color.White)
		}
	}
	if !bytes.Equal([]byte{0xFF, 0xFF, 0xFF, 0xFF}, tb.PixLSB) || !bytes.Equal([]byte{0xFF, 0xFF, 0xFF, 0xFF}, tb.PixMSB) {
		t.Errorf("Expected 4x 0xFF in both planes, got %v, %v", tb.PixLSB, tb.PixMSB)
	}
}

func TestPlaneOrder(t *testing.T) {
	tb := NewBitPlane(image.Rect(0, 0, 16, 2))
	for y := 0; y < tb.Rect.Dy(); y++ {
		for x := 0; x < tb.Rect.Dx(); x++ {
			tb.Set(x, y, DarkGray)
		}
	}

	// The most significant plane should be black for *dark* gray
	if !bytes.Equal([]byte{0x00, 0x00, 0x00, 0x00}, tb.PixMSB) || !bytes.Equal([]byte{0xFF, 0xFF, 0xFF, 0xFF}, tb.PixLSB) {
		t.Errorf("Expected 4x 00 in MSB plane, 4x FF in LSB plane, got %v, %v", tb.PixMSB, tb.PixLSB)
	}
}

func TestBitPlaneEncoding(t *testing.T) {
	tb := NewBitPlane(image.Rect(0, 0, 8, 1))

	// "golden image" test for a black image with two pixels set
	tb.Set(0, 0, White)
	tb.Set(2, 0, LigthGray)

	expectedMSB := []byte{0b10100000}
	expectedLSB := []byte{0b10000000}

	if !bytes.Equal(tb.PixMSB, expectedMSB) || !bytes.Equal(tb.PixLSB, expectedLSB) {
		t.Errorf("Golden image test failed, got %02x %02x, expected %02x %02x", tb.PixMSB, tb.PixLSB, expectedMSB, expectedLSB)
	}
}

func TestOutOfBoundsRead(t *testing.T) {
	tb := NewBitPlane(image.Rect(0, 0, 32, 32))

	if tb.At(10000, 10000) != Black {
		t.Error("Expected out of bounds read to return black")
	}
}

func TestOutOfBoundsWrite(t *testing.T) {
	tb := NewBitPlane(image.Rect(0, 0, 32, 32))

	// will panic if bounds checking is not implemented :)
	tb.Set(10000, 10000, White)
}

func TestGrayAt(t *testing.T) {
	tb := NewBitPlane(image.Rect(0, 0, 16, 2))
	var grays []Gray
	for y := 0; y < tb.Rect.Dy(); y++ {
		for x := 0; x < tb.Rect.Dx(); x++ {
			g := Gray((x ^ y) & 0b11)
			tb.Set(x, y, g)
			grays = append(grays, g)
		}
	}

	for y := 0; y < tb.Rect.Dy(); y++ {
		for x := 0; x < tb.Rect.Dx(); x++ {
			expected := grays[16*y+x]
			got := tb.GrayAt(x, y)
			if expected != got {
				t.Errorf("Expected %02x at (%d,%d), got %02x", expected, x, y, got)
			}
		}
	}
}

func TestConvertGrayToSelf(t *testing.T) {
	for _, c := range []Gray{White, LigthGray, DarkGray, Black} {
		r, g, b, a := c.RGBA()
		gray := convert(color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)})
		if gray != c {
			t.Errorf("Converting '%v' to uint16(%v,%v,%v,%v) and back to gray yields different gray '%v'", c, r, g, b, a, gray)
		}
	}
}
