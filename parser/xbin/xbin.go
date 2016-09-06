// Package xbin contains a parser for the eXtended Binary text format.
package xbin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"

	"git.maze.io/maze/go-piece/buffer"
	"git.maze.io/maze/go-piece/font"
	"git.maze.io/maze/go-piece/parser"
	"git.maze.io/maze/go-piece/parser/binarytext"
	sauce "git.maze.io/maze/go-sauce"
)

var (
	// XBINID is the XBIN header identifier
	XBINID = []byte("XBIN")

	errCompress  = errors.New("Invalid compression byte in XBIN")
	errNotXBIN   = errors.New("Not an XBIN")
	errShortRead = errors.New("Short read")

	errStrDecompress = "error parsing compression [%s]: %v"
	errStrHeader     = "error parsing header: %v"
	errStrImage      = "error parsing image: %v"
	errStrFont       = "error parsing font: %v"
	errStrPalette    = "error parsing palette: %v"
)

const (
	// FlagPalette is set if a palette is present in the XBIN
	FlagPalette = 1 << iota
	// FlagFont is set if a font is present in the XBIN
	FlagFont
	// FlagCompress is set if compression is used
	FlagCompress
	// FlagNonBlink is set if image should use high intensity background colors
	FlagNonBlink
	// Flag512Chars is set if the font has 512 characters in stead of 256
	Flag512Chars
)

const (
	compressNone uint8 = 0x00
	compressChar uint8 = 0x40
	compressAttr uint8 = 0x80
	compressBoth uint8 = 0xc0
)

var compression = map[uint8]string{
	compressNone: "none",
	compressChar: "char",
	compressAttr: "attr",
	compressBoth: "both",
}

// XBIN implements the eXtended Binary text format
type XBIN struct {
	Palette *color.Palette
	buffer  *buffer.Buffer
	data    []byte
	font    *font.Font
	header  Header
	sauce   *sauce.SAUCE
}

// Header implements the eXtended Binary header format
type Header struct {
	ID            [4]byte
	EOFChar       byte
	Width, Height uint16
	Fontsize      uint8
	Flags         uint8
}

// New initializes a new eXtended Binary parser
func New() *XBIN {
	p := &XBIN{
		Palette: &binarytext.BINPalette,
		buffer:  buffer.New(80, 1),
	}
	return p
}

// Parse the eXtended Binary buffer
func (p *XBIN) Parse(r io.Reader) (err error) {
	r.Read(p.header.ID[:])
	if !bytes.Equal(p.header.ID[:], XBINID) {
		return errNotXBIN
	}

	p.header.EOFChar, err = readByte(r)
	p.header.Width, err = readShort(r)
	p.header.Height, err = readShort(r)
	p.header.Fontsize, err = readByte(r)
	p.header.Flags, err = readByte(r)
	if err != nil {
		return fmt.Errorf(errStrHeader, err)
	}

	// Parse palette, if set
	if p.header.Flags&FlagPalette > 0 {
		//log.Println("xbin: custom palette")
		pal := color.Palette{}

		for i := 0; i < 16; i++ {
			var rgb = make([]byte, 3)
			var n int

			if n, err = r.Read(rgb); err != nil {
				return fmt.Errorf(errStrPalette, err)
			}
			if n != 3 {
				return fmt.Errorf(errStrPalette, errShortRead)
			}

			pal = append(pal, color.RGBA{
				rgb[0] << 2,
				rgb[1] << 2,
				rgb[2] << 2,
				0xff,
			})
		}
		p.Palette = &pal
	}

	// Parse font, if set
	if p.header.Flags&FlagFont > 0 {
		size := int(p.header.Fontsize) * 256
		char := make([]byte, size)
		if p.header.Flags&Flag512Chars == Flag512Chars {
			char = append(char, make([]byte, size)...)
		}
		n, err := r.Read(char)
		if err != nil {
			return fmt.Errorf(errStrFont, err)
		}
		if n != len(char) {
			return fmt.Errorf(errStrFont, errShortRead)
		}
		if p.font, err = font.NewBinary(char, int(p.header.Fontsize)); err != nil {
			return fmt.Errorf(errStrFont, err)
		}
	}

	// Parse compression, if set
	l := int(p.header.Width) * int(p.header.Height) * 2
	if p.header.Flags&FlagCompress > 0 {
		if p.data, err = decompress(r, l); err != nil {
			return fmt.Errorf(errStrImage, err)
		}
	} else {
		p.data = make([]byte, l)
		if _, err = r.Read(p.data); err != nil {
			return fmt.Errorf(errStrImage, err)
		}
	}

	// Initialize buffer
	w := int(p.header.Width)
	h := int(p.header.Height)
	//log.Printf("xbin: creating a %d x %d piece\n", w, h)
	p.buffer = buffer.New(w, h)

	// Parse remaining data, scanning for a SAUCE header
	var d []byte
	if d, err = ioutil.ReadAll(r); err == nil {
		var s *sauce.SAUCE
		if s, err = sauce.ParseBytes(d); err == nil && s.DataType == sauce.DataTypeXBIN {
			p.buffer.Flags = s.TFlags
		}
		err = nil // Don't bleed unimportant SAUCE warnings
	}

	//log.Printf("xbin: loading %d bytes of memory\n", len(p.data))
	p.buffer.FromMemory(p.data)
	return
}

// HTML returns the internal buffer as HTML.
func (p *XBIN) HTML(full bool) (string, error) {
	return "", parser.ErrNotSupported
}

// String returns the internal buffer as string.
func (p *XBIN) String() string {
	var b []byte

	w, h := p.buffer.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b = append(b, p.buffer.TileAt(x, y).Char)
		}
		b = append(b, []byte("\r\n")...)
	}

	return string(b)
}

// Font returns the font for this XBIN.
func (p *XBIN) Font() *font.Font {
	return p.font
}

// Width returns the number of columns.
func (p *XBIN) Width() int {
	w, _ := p.buffer.Size()
	return w
}

// Height returns the number of rows.
func (p *XBIN) Height() int {
	_, h := p.buffer.Size()
	return h
}

// Image returns the internal buffer as an image.
func (p *XBIN) Image(f *font.Font) (m image.Image, err error) {
	return p.buffer.Image(*p.Palette, f)
}

// SetFlags imports SAUCE flags
func (p *XBIN) SetFlags(f sauce.TFlags) {
	p.buffer.Flags = f
}

func (p *XBIN) SAUCE() *sauce.SAUCE {
	return p.sauce
}

func decompress(src io.Reader, size int) (dst []byte, err error) {
	dst = make([]byte, 0)

	var kind, counter uint8
	var b []byte
	var char, attr byte
	for size > 0 {
		if counter, err = readByte(src); err != nil {
			err = fmt.Errorf(errStrDecompress, "counter", err)
			return
		}
		kind = counter & 0xc0
		counter = (counter & 0x3f) + 1

		switch kind {
		case compressNone:
			c := int(counter) * 2
			if b, err = readBytes(src, c); err != nil {
				err = fmt.Errorf(errStrDecompress, compression[kind], err)
				return
			}
			dst = append(dst, b...)
			size -= int(c)
		case compressChar:
			if char, err = readByte(src); err != nil {
				err = fmt.Errorf(errStrDecompress, compression[kind], err)
				return
			}
			for i := uint8(0); i < counter; i++ {
				if attr, err = readByte(src); err != nil {
					err = fmt.Errorf(errStrDecompress, compression[kind], err)
					return
				}
				dst = append(dst, char)
				dst = append(dst, attr)
				size -= 2
			}
		case compressAttr:
			if attr, err = readByte(src); err != nil {
				err = fmt.Errorf(errStrDecompress, err)
				return
			}
			for i := uint8(0); i < counter; i++ {
				if char, err = readByte(src); err != nil {
					err = fmt.Errorf(errStrDecompress, compression[kind], err)
					return
				}
				dst = append(dst, char)
				dst = append(dst, attr)
				size -= 2
			}
		case compressBoth:
			var both []byte
			if both, err = readBytes(src, 2); err != nil {
				err = fmt.Errorf(errStrDecompress, compression[kind], err)
				return
			}
			for i := uint8(0); i < counter; i++ {
				dst = append(dst, both...)
				size -= 2
			}
		default:
			err = fmt.Errorf(errStrDecompress, kind, errCompress)
			return
		}
	}

	return
}

func readBytes(r io.Reader, i int) (out []byte, err error) {
	out = make([]byte, i)
	var n int
	n, err = r.Read(out)
	if err != nil {
		return
	}
	if n != i {
		return out, errShortRead
	}
	return
}

func readByte(r io.Reader) (byte, error) {
	var out = make([]byte, 1)
	if _, err := r.Read(out); err != nil {
		return 0x00, err
	}
	return out[0], nil
}

func readShort(r io.Reader) (uint16, error) {
	var out = make([]byte, 2)
	if _, err := r.Read(out); err != nil {
		return 0x0000, err
	}
	return binary.LittleEndian.Uint16(out), nil
}

var _ parser.Parser = (*XBIN)(nil)
