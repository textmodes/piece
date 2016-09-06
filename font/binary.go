package font

import (
	"errors"
	"image"
	"image/color"
	"log"

	"git.maze.io/maze/go-piece/math"
)

type BitMask struct {
	Bitmap     []byte
	CharHeight int
}

func (p *BitMask) At(x, y int) color.Color {
	if y < 0 || y > p.CharHeight {
		log.Printf("bitmask: y=%d out of bounds (%d)\n", y, p.CharHeight)
		return color.Alpha{}
	}
	if x < 0 || x > 8*len(p.Bitmap) {
		log.Printf("bitmask: x=%d out of bounds (%d)\n", x, 8*len(p.Bitmap))
	}

	dx, mx := math.DivMod(x, 8)
	o := p.CharHeight*dx + y
	px := uint8(1 << uint8(7-mx))

	/*
		log.Printf("bitmask: at %d, %d\n", x, y)

		if x >= 776 && x < 784 {
			log.Printf("bitmask: x=%d, y=%d, dx=%d, mx=%d, o=%d, px=%d, bitmap=%02x\n", x, y, dx, mx, o, px, p.Bitmap[o])
		}
	*/

	if p.Bitmap[o]&px == px {
		return color.Alpha{0xff}
	}
	return color.Alpha{0x00}
}

func (p *BitMask) Bounds() image.Rectangle {
	c := len(p.Bitmap) / p.CharHeight
	return image.Rectangle{image.ZP, image.Pt(c*8, p.CharHeight)}
}

func (p *BitMask) ColorModel() color.Model {
	return color.AlphaModel
}

func NewBinary(d []byte, h int) (*Font, error) {
	c := len(d) / h
	if c%256 != 0 {
		return nil, errors.New("Number of glyphs must be a multiple of 256")
	}
	log.Printf("font: new binary font with %d glyphs\n", c)
	f := &Font{
		Image: &BitMask{
			Bitmap:     d,
			CharHeight: h,
		},
		Size: image.Pt(8, h),
	}
	return f, f.setMask()
}

var _ image.Image = (*BitMask)(nil)
