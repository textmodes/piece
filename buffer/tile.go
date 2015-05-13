package buffer

import "fmt"

const (
	TILE_DEFAULT_CHAR       = 0x20
	TILE_DEFAULT_COLOR      = 0x07
	TILE_DEFAULT_BACKGROUND = 0x00
)

const (
	ATTRIB_BOLD                      = 1 << iota // bold or increased intensity
	ATTRIB_FAINT                                 // faint, decreased intensity or second colour
	ATTRIB_ITALICS                               // italicized
	ATTRIB_UNDERLINE                             // underlined
	ATTRIB_BLINK                                 // blinking
	ATTRIB_NEGATIVE                              // negative image
	ATTRIB_CONCEAL                               // concealed characters
	ATTRIB_CROSS_OUT                             // crossed-out (characters still legible but marked as to be deleted)
	ATTRIB_GOTHIC                                // Fraktur (gothic)
	ATTRIB_UNDERLINE_DOUBLE                      // doubly underlined
	ATTRIB_FRAME                                 // framed
	ATTRIB_ENCIRCLE                              // encircled
	ATTRIB_OVERLINE                              // overlined
	ATTRIB_IDEOGRAM_UNDERLINE                    // ideogram underline or right side line
	ATTRIB_IDEOGRAM_UNDERLINE_DOUBLE             // ideogram double underline or double line on the right side
	ATTRIB_IDEOGRAM_OVERLINE                     // ideogram overline or left side line
	ATTRIB_IDEOGRAM_OVERLINE_DOUBLE              // ideogram double overline or double line on the left side
	ATTRIB_IDEOGRAM_STRESS_MARKING               // ideogram stress marking
)

type Tile struct {
	Char              byte
	Color, Background int
	Font              int
	Attrib            uint32
}

func NewTile() *Tile {
	t := &Tile{}
	return t.Reset()
}

func (t *Tile) Equal(o *Tile) bool {
	if o == nil {
		return false
	}
	return t.Color == o.Color && t.Background == o.Background && t.Attrib == o.Attrib
}

func (t *Tile) Reset() *Tile {
	t.Char = TILE_DEFAULT_CHAR
	t.ResetAttrib()
	return t
}

func (t *Tile) ResetAttrib() *Tile {
	t.Color = TILE_DEFAULT_COLOR
	t.Background = TILE_DEFAULT_BACKGROUND
	t.Font = 0
	t.Attrib = 0
	return t
}

func (t *Tile) String() string {
	return fmt.Sprintf(`char=0x%02x, fg=%02d, bg=%02d, font=%d, attrib=%d`,
		t.Char, t.Color, t.Background, t.Font, t.Attrib)
}

func (t *Tile) Update(o *Tile) *Tile {
	t.Char = o.Char
	t.Color, t.Background = o.Color, o.Background
	t.Font = o.Font
	t.Attrib = o.Attrib
	return t
}
