package binarytext

import (
	"image"
	"image/color"
	"io"
	"io/ioutil"

	"github.com/textmodes/piece/buffer"
	"github.com/textmodes/piece/font"
	"github.com/textmodes/sauce"
)

type BinaryText struct {
	Palette *color.Palette
	buffer  *buffer.Buffer
	font    *font.Font
}

func New() *BinaryText {
	p := &BinaryText{
		Palette: &BINPalette,
	}
	return p
}

// Parse the Binary Text buffer
func (p *BinaryText) Parse(r io.Reader) (err error) {
	var b []byte
	b, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	// SAUCE can alter the width
	var w = 80
	var s *sauce.SAUCE
	s, err = sauce.ParseBytes(b)
	if err == nil {
		b = b[:len(b)-128] // Remove SAUCE record
		if s.FileType > 0 {
			w = int(uint16(s.FileType) << 1)
		}
	}

	h := int(len(b) / w / 2)
	p.buffer = buffer.New(w, h)
	p.buffer.FromMemory(b)
	return
}

// Font returns the font (always nil, no embedded font support)
func (p *BinaryText) Font() *font.Font {
	return nil
}

func (p *BinaryText) Image(f *font.Font) (image.Image, error) {
	return p.buffer.Image(*p.Palette, f)
}
