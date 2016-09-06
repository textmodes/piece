package buffer

import (
	"fmt"

	"git.maze.io/maze/go-piece/buffer/attribute"
)

const (
	DefaultChar       = 0x20
	DefaultColor      = 0x07
	DefaultBackground = 0x00
)

type Tile struct {
	Char              byte
	Color, Background int
	Font              int
	Attributes        uint32
}

func NewTile() *Tile {
	t := &Tile{}
	return t.Reset()
}

func (t *Tile) Equal(o *Tile) bool {
	if o == nil {
		return false
	}
	return t.Color == o.Color && t.Background == o.Background && t.Attributes == o.Attributes
}

func (t *Tile) Reset() *Tile {
	t.Char = DefaultChar
	t.ResetAttributes()
	return t
}

func (t *Tile) ResetAttributes() *Tile {
	t.Color = DefaultColor
	t.Background = DefaultBackground
	t.Font = 0
	t.Attributes = attribute.None
	return t
}

func (t *Tile) String() string {
	return fmt.Sprintf(`char=0x%02x, fg=%02d, bg=%02d, font=%d, attrib=%d`,
		t.Char, t.Color, t.Background, t.Font, t.Attributes)
}

func (t *Tile) Update(o *Tile) *Tile {
	t.Char = o.Char
	t.Color, t.Background = o.Color, o.Background
	t.Font = o.Font
	t.Attributes = o.Attributes
	return t
}
