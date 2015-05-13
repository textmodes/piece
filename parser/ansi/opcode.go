package ansi

import (
	"errors"
	"image/color"
	"log"
	"strconv"

	"github.com/textmodes/piece/buffer"
	"github.com/textmodes/piece/calc"
)

// "24 bit ANSi" by picoe.ca (meh)
// http://picoe.ca/2014/03/07/24-bit-ansi/
func (p *ANSI) parse24B(s *Sequence) (err error) {
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

	var found bool
	var c int
	for c = range p.Palette {
		switch t := p.Palette[c].(type) {
		case *color.RGBA:
			if t.R == i[0] && t.G == i[1] && t.B == i[2] {
				found = true
			}
		}
	}

	if !found {
		rgba := color.RGBA{i[1], i[2], i[3], 0xff}
		c = len(p.Palette)
		p.Palette = append(p.Palette, rgba)
	}

	switch i[0] {
	case 0: // Background
		p.buffer.Cursor.Background = c
	case 1: // Foreground
		p.buffer.Cursor.Color = c
	default:
		return errors.New("Unexpected 24 bit color selector")
	}

	return
}

// Cursor Character Absolute
func (p *ANSI) parseCHA(s *Sequence) (err error) {
	x := 0
	if s.Len() > 0 {
		x = s.Int(0) - 1
	}
	p.buffer.Cursor.X = calc.MaxInt(0, x)
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
	p.buffer.Cursor.Left(calc.MaxInt(1, s.Int(0)))
	return
}

// Cursor Down
func (p *ANSI) parseCUD(s *Sequence) (err error) {
	p.buffer.Cursor.Down(calc.MaxInt(1, s.Int(0)))
	return
}

// Cursor Right
func (p *ANSI) parseCUF(s *Sequence) (err error) {
	p.buffer.Cursor.Right(calc.MaxInt(1, s.Int(0)))
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
	p.buffer.Cursor.Goto(calc.MaxInt(0, x-1), calc.MaxInt(0, y-1))
	return
}

// Cursor Up
func (p *ANSI) parseCUU(s *Sequence) (err error) {
	p.buffer.Cursor.Up(calc.MaxInt(1, s.Int(0)))
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

	o = calc.MaxInt(o, 0)
	e = calc.MinInt(e, p.buffer.Len())

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
	for _, n := range s.Ints() {
		switch n {
		// ECMA-48 standard codes
		case 0: // Default rendition
			p.buffer.Cursor.ResetAttrib()
		case 1: // Bold
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_BOLD
		case 2: // Faint
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_FAINT
		case 3: // Italicized
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_ITALICS
		case 4: // Underlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE_DOUBLE
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_UNDERLINE
		case 5, 6: // Blink
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_BLINK
		case 7: // Negative
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_NEGATIVE
		case 8: // Concealed characters
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_CONCEAL
		case 9: // Crossed-out
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_CROSS_OUT
		case 10, 11, 12, 13, 14, 15, 16, 17, 18, 19:
			p.buffer.Cursor.Font = n - 10
		case 20: // Fraktur (Gothic)
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_GOTHIC
		case 21: // Doubly underlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_UNDERLINE_DOUBLE
		case 22: // Neither bold nor faint
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_BOLD
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_FAINT
		case 23: // Neither italicized nor fraktur
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_ITALICS
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_GOTHIC
		case 24: // Not underlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE_DOUBLE
		case 25: // Not blinking
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_BLINK
		case 26: // Reserved
		case 27: // Positive
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_NEGATIVE
		case 28: // Revealed
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_CONCEAL
		case 29: // Not crossed out
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_CROSS_OUT
		case 30, 31, 32, 33, 34, 35, 36, 37:
			p.buffer.Cursor.Color = n - 30
		case 38: // Reserved (TODO 24 bit ANSi)
		case 39: // Default display colour
			p.buffer.Cursor.Color = buffer.TILE_DEFAULT_COLOR
		case 40, 41, 42, 43, 44, 45, 46, 47:
			p.buffer.Cursor.Background = n - 40
		case 48: // Reserved (TODO 24 bit ANSi)
		case 49: // Default background colour
			p.buffer.Cursor.Background = buffer.TILE_DEFAULT_BACKGROUND
		case 50: // Reserved (cancels 26)
		case 51: // Framed
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_FRAME
		case 52: // Encircled
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_ENCIRCLE
		case 53: // Overlined
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_OVERLINE
		case 54: // Not framed nor encircled
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_FRAME
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_ENCIRCLE
		case 55: // Not overlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_OVERLINE
		case 56, 57, 58, 59: // Reserved
		case 60: // Ideogram underline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_UNDERLINE
		case 61: // Ideogram double underline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_UNDERLINE_DOUBLE
		case 62: // Ideogram overline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_OVERLINE
		case 63: // Ideogram double overline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_OVERLINE_DOUBLE
		case 64: // Ideogram stress marking
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_STRESS_MARKING
		case 65: // Cancels 60..64
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_UNDERLINE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_UNDERLINE_DOUBLE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_OVERLINE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_OVERLINE_DOUBLE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_STRESS_MARKING

		// Non default aixterm codes, bright color variants
		case 90, 91, 92, 93, 94, 95, 96, 97:
			p.buffer.Cursor.Color = n - 84
		case 100, 101, 102, 103, 104, 105, 106, 107:
			p.buffer.Cursor.Background = n - 94

		default: // Fallthrough
			log.Printf("unsupported SGR %d\n", n)
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
