// Package ansi is a parser for ANSi text, compliant with the ECMA-48 and/or ANSI.SYS specification
package ansi

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/textmodes/piece/buffer"
	"github.com/textmodes/piece/calc"
	"github.com/textmodes/piece/font"
	"github.com/textmodes/piece/parser"
	"github.com/textmodes/sauce"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	tabStop = 8
)

const (
	stateExit = iota
	stateText
	stateANSIWaitBrace
	stateANSIWaitLiteral
)

// ECMA-48 specified Final Bytes of control sequences without intermediate bytes
const (
	AnsiICH       = iota + 0x40 // '@', Insert Character
	AnsiCUU                     // 'A', Cursor Up
	AnsiCUD                     // 'B', Cursor Down
	AnsiCUF                     // 'C', Cursor Right
	AnsiCUB                     // 'D', Cursor Left
	AnsiCNL                     // 'E', Cursor Next Line
	AnsiCPL                     // 'F', Cursor Preceiding Line
	AnsiCHA                     // 'G', Cursor Character Absolute
	AnsiCUP                     // 'H', Cursor Position
	AnsiCHT                     // 'I', Cursor Forward Tabulation
	AnsiED                      // 'J', Erase in Page
	AnsiEL                      // 'K', Erase in Line
	AnsiIL                      // 'L', Insert Line
	AnsiDL                      // 'M', Delete Line
	AnsiEF                      // 'N', Erase in Field
	AnsiEA                      // 'O', Erase in Area
	AnsiDCH                     // 'P', Delete Character
	AnsiSSE                     // 'Q, ???
	AnsiCPR                     // 'R', Active Position Report
	AnsiSU                      // 'S', Scroll Up
	AnsiSD                      // 'T', Scroll Down
	AnsiNP                      // 'U', Next Page
	AnsiPP                      // 'V', Preceding Page
	AnsiCTC                     // 'W', Cursor Tabulation Control
	AnsiECH                     // 'X', Erase Character
	AnsiCVT                     // 'Y', Cursor Line Tabulation
	AnsiCBT                     // 'Z', Cursor Backward Tabulation
	AnsiSRS                     // '[', Start Reversed String
	AnsiPTX                     // '\', Paralell Texts
	AnsiSDS                     // ']', Start Directed String
	AnsiSIMD                    // '^', Select Implicit Movement Direction
	AnsiUNDEFINED               // ' ', Unspecified
	AnsiHPA                     // '`', Character Position Absolute
	AnsiHPR                     // 'a', Character Position Forward
	AnsiREP                     // 'b', Repeat
	AnsiDA                      // 'c', Device Attributes
	AnsiVPA                     // 'd', Line Position Absolute
	AnsiVPR                     // 'e', Line Position Forward
	AnsiHVP                     // 'f', Character and Line Position
	AnsiTBC                     // 'g', Tabulation Clear
	AnsiSM                      // 'h', Set Mode
	AnsiMC                      // 'i', Media Copy
	AnsiHPB                     // 'j', Character Position Absolute
	AnsiVPB                     // 'k', Line Position Backward
	AnsiRM                      // 'l', Reset Mode
	AnsiSGR                     // 'm', Select Graphic Rendition
	AnsiDSR                     // 'n', Device Status Report
	AnsiDAQ                     // 'o', Define Area Qualification
	_                           // Unused
	_                           // Unused
	AnsiSKS                     // 'p', Set Keyboard String (ANSI.SYS)
	AnsiSCP                     // 's', Save Cursor Position (ANSI.SYS)
	Ansi24B                     // 't', "24 bit ANSi" (picoe.ca)
	AnsiRCP                     // 'u', Restore Cursor Position (ANSI.SYS)
)

type ansiOp func(seq *Sequence) error

// ANSI or ASCII parser
type ANSI struct {
	Palette   color.Palette
	buffer    *buffer.Buffer
	opcode    map[byte]ansiOp
	transform transform.Transformer
	save      *buffer.Cursor
}

// New initializes a new ANSi parser with an initial given width and height
func New(w, h int) *ANSI {
	p := &ANSI{
		Palette:   VGAPalette,
		buffer:    buffer.New(w, h),
		transform: charmap.CodePage437.NewDecoder(),
	}
	p.opcode = map[byte]ansiOp{
		Ansi24B: p.parse24B,
		AnsiCHA: p.parseCHA,
		AnsiCNL: p.parseCNL,
		AnsiCPL: p.parseCHA,
		AnsiCUB: p.parseCUB,
		AnsiCUD: p.parseCUD,
		AnsiCUF: p.parseCUF,
		AnsiCUP: p.parseCUP,
		AnsiCUU: p.parseCUU,
		AnsiED:  p.parseED,
		AnsiEL:  p.parseEL,
		AnsiIL:  p.parseIL,
		AnsiHVP: p.parseCUP, // alias
		AnsiRM:  p.parseRM,
		AnsiSGR: p.parseSGR,
		AnsiSM:  p.parseSM,
		AnsiSCP: p.parseSCP,
		AnsiRCP: p.parseRCP,
	}
	return p
}

// Parse the ANSi sequences from a reader
func (p *ANSI) Parse(r io.Reader) (err error) {
	state := stateText
	buf := bufio.NewReader(r)

	var seq = NewSequence()
	var n, t int
	for state != stateExit {
		var ch byte
		if ch, err = buf.ReadByte(); err != nil {
			log.Printf("exit parser: err=%v, t=%d\n", err, t)
			state = stateExit
			continue
		}

		t += n

		switch state {
		case stateText:
			switch ch {
			case SUB: // End Of File
				state = stateExit
			case ESC:
				state = stateANSIWaitBrace
			case NL:
				p.buffer.Cursor.Y++
			case CR:
				p.buffer.Cursor.X = 0
			case TAB:
				c := (p.buffer.Cursor.X + 1) % tabStop
				if c > 0 {
					c = tabStop - c
					for i := 0; i < c; i++ {
						p.buffer.PutChar(' ')
					}
				}
			default:
				p.buffer.PutChar(ch)
			}

		case stateANSIWaitBrace:
			if ch == '[' {
				state = stateANSIWaitLiteral
			} else {
				p.buffer.PutChar(ESC)
				p.buffer.PutChar(ch)
			}

		case stateANSIWaitLiteral:
			if ch == ';' {
				seq.Flush()
				break
			}

			if isAlpha(ch) {
				seq.Flush()
				//log.Printf("ANSI sequence <ESC>[%s%c (0x%02x)\n", seq, ch, ch)

				fn := p.opcode[ch]
				if fn == nil {
					log.Printf("Unsupported ANSI sequence <ESC>[%s%c (0x%02x)\n", seq, ch, ch)
				} else {
					// Log "special" sequences for great profit
					if ch != AnsiCUF && ch != AnsiSGR && ch != Ansi24B {
						log.Printf("ANSI sequence <ESC>[%s%c (0x%02x)\n", seq, ch, ch)
					}
					if err = fn(seq); err != nil {
						log.Printf("Parser error: %v\n", err)
					}
				}

				seq.Reset()
				state = stateText
				break
			} // if isAlpha(ch)
			seq.Buffer(ch)

		default:
			break
		}
	}

	return nil
}

// SetFlags imports SAUCE flags.
func (p *ANSI) SetFlags(f sauce.TFlags) {
	p.buffer.Flags = f
}

// HTML returns the internal buffer as HTML.
func (p *ANSI) HTML() (s string, err error) {
	s += "<!doctype html>\n"
	s += "<link rel=\"stylesheet\" href=\"cp437.css\">\n"
	s += "<style type=\"text/css\">\n"
	for i := 0; i < len(p.Palette); i++ {
		r, g, b, _ := p.Palette[i].RGBA()
		c := fmt.Sprintf("#%02x%02x%02x", r, g, b)
		s += fmt.Sprintf(".f%02x{color:%s} ", i, c)
		s += fmt.Sprintf(".b%02x{background-color:%s} ", i, c)
		s += fmt.Sprintf(".u%02x{border-bottom:1px solid %s}", i, c)
		s += "\n"
	}
	s += `.i{font-variant:italics} .u{border-bottom:1px} .ud{border-bottom:3px dashed #000}`
	s += "</style>"

	s += fmt.Sprintf(`<pre><span class="b%02x f%02x">`,
		buffer.TILE_DEFAULT_BACKGROUND,
		buffer.TILE_DEFAULT_COLOR)

	w, h := p.buffer.SizeMax()
	var l *buffer.Tile

	for o, t := range p.buffer.Tiles {
		y, x := calc.DivMod(o, p.buffer.Width)
		if x >= w {
			continue
		}
		if y >= h {
			break
		}
		if x == 0 && y > 0 {
			s += "\n"
		}
		if t == nil {
			s += " "
		} else if t.Equal(l) {
			//s += string(t.Char)
			if isPrint(t.Char) {
				s += string(t.Char)
			} else {
				s += fmt.Sprintf(`&#x%02x;`, t.Char)
			}
		} else {
			f := t.Color
			b := t.Background
			c := []string{}

			if t.Attrib&buffer.ATTRIB_BOLD == buffer.ATTRIB_BOLD {
				f += 8
			}
			if t.Attrib&buffer.ATTRIB_BLINK == buffer.ATTRIB_BLINK {
				b += 8
			}
			if t.Attrib&buffer.ATTRIB_NEGATIVE == buffer.ATTRIB_NEGATIVE {
				f, b = b, f
			}
			c = append(c, fmt.Sprintf("b%02x", b))
			c = append(c, fmt.Sprintf("f%02x", f))
			if t.Attrib&buffer.ATTRIB_ITALICS > 0 {
				c = append(c, "i")
			}
			if t.Attrib&buffer.ATTRIB_UNDERLINE > 0 {
				c = append(c, fmt.Sprintf("u%02x", f))
			}
			if t.Attrib&buffer.ATTRIB_UNDERLINE_DOUBLE > 0 {
				c = append(c, "ud")
			}

			s += `</span>`
			s += fmt.Sprintf(`<span class="%s">`, strings.Join(c, " "))
			if isPrint(t.Char) {
				s += string(t.Char)
			} else {
				s += fmt.Sprintf(`&#x%02x;`, t.Char)
			}
		}

		l = t
	}

	s += `</span>`
	s += `</pre>`
	return
}

// Font returns nil, as an ANSi file has no font data.
func (p *ANSI) Font() *font.Font {
	return nil
}

// Image returns the internal buffer as an image.
func (p *ANSI) Image(f *font.Font) (image.Image, error) {
	return p.buffer.Image(p.Palette, f)
}

func (p *ANSI) String() (s string) {
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

// Sequence holds an ANSi escape sequence.
type Sequence struct {
	s []string
	b []byte
}

// NewSequence initializes a new Sequence structure.
func NewSequence() *Sequence {
	return &Sequence{
		s: make([]string, 0),
		b: make([]byte, 0),
	}
}

// Buffer a byte to the internal buffer
func (s *Sequence) Buffer(b byte) {
	s.b = append(s.b, b)
}

// Bytes returns the sequence as bytes
func (s *Sequence) Bytes() (out []byte) {
	return []byte(s.String())
}

// Flush the internal buffer to the sequences
func (s *Sequence) Flush() {
	s.s = append(s.s, string(s.b))
	s.b = make([]byte, 0)
}

// Int returns the integer value of sequence n
func (s *Sequence) Int(n int) (i int) {
	if n < s.Len() {
		i, _ = strconv.Atoi(s.s[n])
	}
	return
}

// Ints returns all sequenced items as int
func (s *Sequence) Ints() (i []int) {
	i = make([]int, 0)
	for _, j := range s.s {
		if n, err := strconv.Atoi(j); err == nil {
			i = append(i, n)
		}
	}
	return
}

// Len returns the number of items in the sequence
func (s *Sequence) Len() int {
	return len(s.s)
}

// Reset reinitializes the internal buffers
func (s *Sequence) Reset() {
	s.s = make([]string, 0)
	s.b = make([]byte, 0)
}

func (s *Sequence) String() string {
	return strings.Join(s.s, ";")
}

func (s *Sequence) StringAt(n int) (out string) {
	if n < s.Len() {
		out = string(s.s[n])
	}
	return
}

var _ parser.Parser = (*ANSI)(nil)
