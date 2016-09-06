package ansi

import (
	"errors"
	"image/color"
	"log"
	"strconv"

	"git.maze.io/maze/go-piece/buffer"
	"git.maze.io/maze/go-piece/buffer/attribute"
	"git.maze.io/maze/go-piece/math"
	"git.maze.io/maze/go-piece/palette"
)

func addRGB(p *palette.Palette, r, g, b uint8) int {
	if palette.IsBuiltin(*p) {
		*p = p.Copy()
	}

	var found bool
	var c int
	for c = range *p {
		switch t := (*p)[c].(type) {
		case *color.RGBA:
			if t.R == r && t.G == g && t.B == b {
				found = true
			}
		}
	}

	if !found {
		rgba := color.RGBA{r, g, b, 0xff}
		c = len(*p)
		(*p) = append((*p), rgba)
	}

	return c
}

// Cursor Character Absolute
func (p *ANSI) parseCHA(s *Sequence) (err error) {
	x := 0
	if s.Len() > 0 {
		x = s.Int(0) - 1
	}
	p.buffer.Cursor.X = math.MaxInt(0, x)
	return
}

// Cursor Next Line
func (p *ANSI) parseCNL(s *Sequence) (err error) {
	y := 1
	if s.Len() > 0 {
		y = s.Int(0)
	}
	p.buffer.Cursor.X = 0
	p.buffer.Cursor.Down(y)
	return
}

// Cursor Preceding Line
func (p *ANSI) parseCPL(s *Sequence) (err error) {
	y := 1
	if s.Len() > 0 {
		y = s.Int(0)
	}
	p.buffer.Cursor.X = 0
	p.buffer.Cursor.Up(y)
	return
}

// Cursor Left
func (p *ANSI) parseCUB(s *Sequence) (err error) {
	p.buffer.Cursor.Left(math.MaxInt(1, s.Int(0)))
	return
}

// Cursor Down
func (p *ANSI) parseCUD(s *Sequence) (err error) {
	p.buffer.Cursor.Down(math.MaxInt(1, s.Int(0)))
	return
}

// Cursor Right
func (p *ANSI) parseCUF(s *Sequence) (err error) {
	p.buffer.Cursor.Right(math.MaxInt(1, s.Int(0)))
	return
}

// Cursor Position
func (p *ANSI) parseCUP(s *Sequence) (err error) {
	y := 1
	x := 1
	switch s.Len() {
	case 2:
		y = s.Int(0)
		x = s.Int(1)
	case 1:
		y = s.Int(0)
	}
	p.buffer.Cursor.Goto(math.MaxInt(0, x-1), math.MaxInt(0, y-1))
	return
}

// Cursor Up
func (p *ANSI) parseCUU(s *Sequence) (err error) {
	p.buffer.Cursor.Up(math.MaxInt(1, s.Int(0)))
	return
}

// Erase Display
func (p *ANSI) parseED(s *Sequence) (err error) {
	i := s.Int(0)

	switch i {
	case 0: // From cursor to EOF
		o := p.buffer.Cursor.Offset(p.buffer.Width)
		p.buffer.ClearFrom(o)
	case 1: // From cursor to start
		o := p.buffer.Cursor.Offset(p.buffer.Width)
		p.buffer.ClearTo(o)
	default: // Entire buffer
		p.buffer.Clear()
	}

	return
}

// Erase Line
func (p *ANSI) parseEL(s *Sequence) (err error) {
	i := s.Int(0)
	var o, e int
	switch i {
	case 0: // To EOL
		o = (p.buffer.Width * (p.buffer.Cursor.Y)) + p.buffer.Cursor.X
		e = (p.buffer.Width * (p.buffer.Cursor.Y + 1)) - 1
	case 1: // To BOL
		o = (p.buffer.Width * (p.buffer.Cursor.Y - 1)) + 1
		e = (p.buffer.Width * (p.buffer.Cursor.Y)) + p.buffer.Cursor.X
	case 2: // From BOL to EOL
		o = (p.buffer.Width * (p.buffer.Cursor.Y - 1)) + 1
		e = (p.buffer.Width * (p.buffer.Cursor.Y + 1)) - 1
	}

	o = math.MaxInt(o, 0)
	e = math.MinInt(e, p.buffer.Len())

	for i = o; i < e; i++ {
		p.buffer.ClearAt(i)
	}

	return
}

// Insert Line
func (p *ANSI) parseIL(s *Sequence) (err error) {
	i := 1
	if s.Len() > 0 {
		i = s.Int(0)
	}
	o := p.buffer.Width * p.buffer.Normalize().Cursor.Y
	for ; i > 0; i-- {
		p.buffer.Insert(o, p.buffer.Width)
	}
	return
}

// Reset Mode
func (p *ANSI) parseRM(s *Sequence) (err error) {
	b := s.Bytes()
	if len(b) == 0 {
		return
	}
	switch b[0] {
	case '?', '=': // DEC private mode
		n, _ := strconv.Atoi(string(b[1:]))
		switch n {
		case 31, 33: // http://www.keelhaul.me.uk/linux/files/noblink.txt
			p.buffer.Flags.NonBlink = false
		}
	default:
		// TODO implement mode switching
		log.Printf("ansi: unsupported mode %s\n", s)
	}
	return

}

// Set Mode
func (p *ANSI) parseSM(s *Sequence) (err error) {
	b := s.Bytes()
	if len(b) == 0 {
		return
	}
	switch b[0] {
	case '?', '=': // DEC private mode
		n, _ := strconv.Atoi(string(b[1:]))
		switch n {
		case 31, 33: // http://www.keelhaul.me.uk/linux/files/noblink.txt
			p.buffer.Flags.NonBlink = true
		}
	default:
		// TODO implement mode switching
		log.Printf("ansi: unsupported mode %s\n", s)
	}
	return
}

func (p *ANSI) parseSGR(s *Sequence) (err error) {
	for i, n := range s.Ints() {
		switch n {
		// ECMA-48 standard codes
		case 0: // Default rendition
			p.buffer.Cursor.ResetAttributes()
		case 1: // Bold
			p.buffer.Cursor.Attributes |= attribute.Bold
		case 2: // Faint
			p.buffer.Cursor.Attributes |= attribute.Faint
		case 3: // Italicized
			p.buffer.Cursor.Attributes |= attribute.Italics
		case 4: // Underlined
			p.buffer.Cursor.Attributes &^= attribute.DoubleUnderline
			p.buffer.Cursor.Attributes |= attribute.Underline
		case 5, 6: // Blink
			p.buffer.Cursor.Attributes |= attribute.Blink
		case 7: // Negative
			p.buffer.Cursor.Attributes |= attribute.Negative
		case 8: // Concealed characters
			p.buffer.Cursor.Attributes |= attribute.Conceal
		case 9: // Crossed-out
			p.buffer.Cursor.Attributes |= attribute.CrossedOut
		case 10, 11, 12, 13, 14, 15, 16, 17, 18, 19:
			p.buffer.Cursor.Font = n - 10
		case 20: // Fraktur (Gothic)
			p.buffer.Cursor.Attributes |= attribute.Gothic
		case 21: // Doubly underlined
			p.buffer.Cursor.Attributes &^= attribute.Underline
			p.buffer.Cursor.Attributes |= attribute.DoubleUnderline
		case 22: // Neither bold nor faint
			p.buffer.Cursor.Attributes &^= attribute.Bold
			p.buffer.Cursor.Attributes &^= attribute.Faint
		case 23: // Neither italicized nor fraktur
			p.buffer.Cursor.Attributes &^= attribute.Italics
			p.buffer.Cursor.Attributes &^= attribute.Gothic
		case 24: // Not underlined
			p.buffer.Cursor.Attributes &^= attribute.Underline
			p.buffer.Cursor.Attributes &^= attribute.DoubleUnderline
		case 25: // Not blinking
			p.buffer.Cursor.Attributes &^= attribute.Blink
		case 26: // Reserved
		case 27: // Positive
			p.buffer.Cursor.Attributes &^= attribute.Negative
		case 28: // Revealed
			p.buffer.Cursor.Attributes &^= attribute.Conceal
		case 29: // Not crossed out
			p.buffer.Cursor.Attributes &^= attribute.CrossedOut
		case 30, 31, 32, 33, 34, 35, 36, 37:
			p.buffer.Cursor.Color = n - 30
		case 38: // Extended set foreground color
			if i > 0 {
				s.Shift(i)
			}
			if s.Len() < 3 {
				log.Printf("broken extended set background <ESC>[%s\n", s)
				return
			}
			switch s.Int(1) {
			case 2: // RGB color
				if s.Len() >= 5 {
					r := uint8(s.Int(2))
					g := uint8(s.Int(3))
					b := uint8(s.Int(4))
					p.buffer.Cursor.Color = addRGB(&p.Palette, r, g, b)
				} else {
					log.Printf("broken RGB color in sequence <ESC>[%s\n", s)
					return
				}
				s.Shift(5)
			case 5: // VGA color index
				if palette.IsBuiltin(p.Palette) && len(p.Palette) == 16 {
					p.Palette = palette.VGA
				}
				p.buffer.Cursor.Color = s.Int(2)
				s.Shift(3)
			}
			// It is, in theory, valid to have more sequences, so parse them too
			if s.Len() > 0 {
				p.parseSGR(s)
			}
			return
		case 39: // Default display colour
			p.buffer.Cursor.Color = buffer.DefaultColor
		case 40, 41, 42, 43, 44, 45, 46, 47:
			p.buffer.Cursor.Background = n - 40
		case 48: // Extended set background color
			if i > 0 {
				s.Shift(i)
			}
			if s.Len() < 3 {
				log.Printf("broken extended set background <ESC>[%s\n", s)
				return
			}
			switch s.Int(1) {
			case 2: // RGB color
				if s.Len() >= 5 {
					r := uint8(s.Int(2))
					g := uint8(s.Int(3))
					b := uint8(s.Int(4))
					p.buffer.Cursor.Background = addRGB(&p.Palette, r, g, b)
				} else {
					log.Printf("broken RGB color in sequence <ESC>[%s\n", s)
					return
				}
				s.Shift(5)
			case 5: // VGA color index
				if palette.IsBuiltin(p.Palette) && len(p.Palette) == 16 {
					p.Palette = palette.VGA
				}
				p.buffer.Cursor.Background = s.Int(2)
				s.Shift(3)
			}
			// It is, in theory, valid to have more sequences, so parse them too
			if s.Len() > 0 {
				p.parseSGR(s)
			}
			return
		case 49: // Default background colour
			p.buffer.Cursor.Background = buffer.DefaultBackground
		case 50: // Reserved (cancels 26)
		case 51: // Framed
			p.buffer.Cursor.Attributes |= attribute.Frame
		case 52: // Encircled
			p.buffer.Cursor.Attributes |= attribute.Encircle
		case 53: // Overlined
			p.buffer.Cursor.Attributes |= attribute.Overline
		case 54: // Not framed nor encircled
			p.buffer.Cursor.Attributes &^= attribute.Frame
			p.buffer.Cursor.Attributes &^= attribute.Encircle
		case 55: // Not overlined
			p.buffer.Cursor.Attributes &^= attribute.Overline
		case 56, 57, 58, 59: // Reserved
		case 60: // Ideogram underline
			p.buffer.Cursor.Attributes |= attribute.IdeogramUnderline
		case 61: // Ideogram double underline
			p.buffer.Cursor.Attributes |= attribute.IdeogramDoubleUnderline
		case 62: // Ideogram overline
			p.buffer.Cursor.Attributes |= attribute.IdeogramOverline
		case 63: // Ideogram double overline
			p.buffer.Cursor.Attributes |= attribute.IdeogramDoubleOverline
		case 64: // Ideogram stress marking
			p.buffer.Cursor.Attributes |= attribute.IdeogramStressMarking
		case 65: // Cancels 60..64
			p.buffer.Cursor.Attributes &^= attribute.IdeogramUnderline
			p.buffer.Cursor.Attributes &^= attribute.IdeogramDoubleUnderline
			p.buffer.Cursor.Attributes &^= attribute.IdeogramOverline
			p.buffer.Cursor.Attributes &^= attribute.IdeogramDoubleOverline
			p.buffer.Cursor.Attributes &^= attribute.IdeogramStressMarking

		// Non default aixterm codes, bright color variants
		case 90, 91, 92, 93, 94, 95, 96, 97:
			p.buffer.Cursor.Color = n - 84
		case 100, 101, 102, 103, 104, 105, 106, 107:
			p.buffer.Cursor.Background = n - 94

		default: // Fallthrough
			log.Printf("unsupported SGR %d (<ESC>[%s)\n", n, s)
		}
	}

	return
}

// Save Cursor Position
func (p *ANSI) parseSCP(s *Sequence) (err error) {
	p.save = buffer.NewCursor(p.buffer.Cursor.X, p.buffer.Cursor.Y)
	return
}

// Restore Cursor Position
func (p *ANSI) parseRCP(s *Sequence) (err error) {
	if p.save != nil {
		p.buffer.Cursor.X = p.save.X
		p.buffer.Cursor.Y = p.save.Y
	}
	return
}

// "24 bit ANSi" by picoe.ca (meh)
// http://picoe.ca/2014/03/07/24-bit-ansi/
func (p *ANSI) parseXXX(s *Sequence) (err error) {
	var i = make([]uint8, 0)
	for _, n := range s.Ints() {
		if n > 0xff {
			continue
		}
		i = append(i, uint8(n))
	}
	if len(i) != 4 {
		return errors.New("Expected a 4 element sequence")
	}

	switch i[0] {
	case 0: // Background
		p.buffer.Cursor.Background = addRGB(&p.Palette, i[1], i[2], i[3])
	case 1: // Foreground
		p.buffer.Cursor.Color = addRGB(&p.Palette, i[1], i[2], i[3])
	default:
		return errors.New("Unexpected 24 bit color selector")
	}

	return
}
