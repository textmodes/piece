package binarytext

import (
	"image"
	"io"
	"io/ioutil"

	"git.maze.io/maze/go-piece/buffer"
	"git.maze.io/maze/go-piece/font"
	"git.maze.io/maze/go-piece/palette"
	"git.maze.io/maze/go-piece/parser"
	sauce "git.maze.io/maze/go-sauce"
)

type BinaryText struct {
	Palette palette.Palette
	buffer  *buffer.Buffer
	font    *font.Font
	sauce   *sauce.SAUCE
}

func New() *BinaryText {
	p := &BinaryText{
		Palette: Palette,
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
	p.sauce, err = sauce.ParseBytes(b)
	if err == nil {
		b = b[:len(b)-128] // Remove SAUCE record
		if p.sauce.FileType > 0 {
			w = int(uint16(p.sauce.FileType) << 1)
		}
	}

	h := int(len(b) / w / 2)
	p.buffer = buffer.New(w, h)
	p.buffer.FromMemory(b)
	return nil
}

// Font returns the font (always nil, no embedded font support)
func (p *BinaryText) Font() *font.Font {
	return nil
}

func (p *BinaryText) Width() int {
	w, _ := p.buffer.Size()
	return w
}

func (p *BinaryText) Height() int {
	_, h := p.buffer.Size()
	return h
}

func (p *BinaryText) Image(f *font.Font) (image.Image, error) {
	return p.buffer.Image(p.Palette, f)
}

func (p *BinaryText) HTML(full bool) (string, error) {
	return "", parser.ErrNotSupported
}

func (p *BinaryText) String() string {
	return ""
}

func (p *BinaryText) SAUCE() *sauce.SAUCE {
	return p.sauce
}
