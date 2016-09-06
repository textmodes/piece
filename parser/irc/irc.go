package irc

import (
	"bufio"
	"image"
	"io"
	"strconv"

	"git.maze.io/maze/go-piece/buffer"
	"git.maze.io/maze/go-piece/buffer/attribute"
	"git.maze.io/maze/go-piece/font"
	"git.maze.io/maze/go-piece/palette"
	"git.maze.io/maze/go-piece/parser"
	sauce "git.maze.io/maze/go-sauce"
)

type IRC struct {
	buffer  *buffer.Buffer
	palette palette.Palette
}

const (
	stateExit = iota
	stateText
	stateColor
	stateBackground
)

const (
	bold       byte = 0x0b
	colored    byte = 0x0c
	italics    byte = 0x1d
	underlined byte = 0x1f
	reversed   byte = 0x16
	reset      byte = 0x0f
)

// colorMap is a mapping between mIRC colors and CGA colors
var colorMap = []int{
	0x0f, /* white */
	0x00, /* black */
	0x01, /* blue */
	0x03, /* green */
	0x09, /* red */
	0x01, /* maroon */
	0x05, /* purple */
	0x03, /* orange */
	0x0b, /* yellow */
	0x0a, /* light green */
	0x06, /* teal */
	0x0e, /* cyan */
	0x0d, /* fuchsia */
	0x08, /* dark grey */
	0x07, /* light grey */
}

func New() *IRC {
	return &IRC{
		buffer:  buffer.New(80, 1),
		palette: palette.CGA,
	}
}

func (p *IRC) Parse(r io.Reader) (err error) {
	state := stateText
	buf := bufio.NewReader(r)

	var fg, bg []byte
	for state != stateExit {
		var ch byte
		if ch, err = buf.ReadByte(); err != nil {
			if err == io.EOF {
				err = nil
			}
			state = stateExit
			continue
		}

		switch state {
		case stateText:
			switch ch {
			case '\r':
			case '\n':
				p.buffer.Cursor.X = 0
				p.buffer.Cursor.Y++
				p.buffer.Cursor.ResetAttributes()
			case bold:
				p.buffer.Cursor.Attributes |= attribute.Bold
			case italics:
				p.buffer.Cursor.Attributes |= attribute.Italics
			case underlined:
				p.buffer.Cursor.Attributes |= attribute.Underline
			case reset:
				p.buffer.Cursor.ResetAttributes()
			case reversed:
				p.buffer.Cursor.Color, p.buffer.Cursor.Background = p.buffer.Cursor.Background, p.buffer.Cursor.Color
			case colored:
				state = stateColor
				fg = []byte{}
				bg = []byte{}
			default:
				p.buffer.PutChar(ch)
			}

		case stateColor:
			switch {
			case ch == ',':
				state = stateBackground

			case ch >= '0' && ch <= '9':
				fg = append(fg, ch)
				if len(fg) == 2 { // Double digits
					c, _ := strconv.Atoi(string(fg))
					p.buffer.Cursor.Color = colorMap[c%16]
					fg = []byte{}
				}

			default:
				if len(fg) > 0 {
					c, _ := strconv.Atoi(string(fg))
					p.buffer.Cursor.Color = colorMap[c%16]
					fg = []byte{}
				}
				state = stateText
			}

		case stateBackground:
			if ch >= '0' && ch <= '9' {
				bg = append(bg, ch)
				if len(bg) == 2 { // Double digits
					c, _ := strconv.Atoi(string(fg))
					p.buffer.Cursor.Background = colorMap[c%16]
					bg = []byte{}
					state = stateText
				}
			} else {
				if len(bg) > 0 {
					c, _ := strconv.Atoi(string(fg))
					p.buffer.Cursor.Background = colorMap[c%16]
					bg = []byte{}
				}
				state = stateText
			}
		}
	}

	return
}

func (p *IRC) HTML(full bool) (string, error) {
	return "", parser.ErrNotSupported
}

func (p *IRC) Image(font *font.Font) (image.Image, error) {
	return nil, parser.ErrNotSupported
}

func (p *IRC) String() (s string) {
	w, h := p.buffer.SizeMax()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y * p.buffer.Width) + x
			t := p.buffer.Tile(o)
			if t == nil {
				s += " "
			} else {
				s += string(t.Char)
			}
		}
		s += "\n"
	}
	return
}

func (p *IRC) Font() *font.Font    { return nil }
func (p *IRC) Width() int          { return p.buffer.Width }
func (p *IRC) Height() int         { return p.buffer.Height }
func (p *IRC) SAUCE() *sauce.SAUCE { return nil }

var _ parser.Parser = (*IRC)(nil)
