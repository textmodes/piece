// Package buffer emulates a text mode screen
package buffer

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"git.maze.io/maze/go-piece/font"
	"git.maze.io/maze/go-piece/math"
	sauce "git.maze.io/maze/go-sauce"
)

var (
	errOutOfBounds = errors.New("Out of bounds")
)

// Buffer holds the information of what would be displayed on a VGA text mode screen.
type Buffer struct {
	Width, Height       int
	Cursor              *Cursor
	Tiles               []*Tile
	Flags               sauce.TFlags
	maxWidth, maxHeight int
}

// New creates a new buffer of w x h Tiles. The maximum buffer width is set to
// the supplied width w.
func New(w, h int) *Buffer {
	b := &Buffer{
		Width:  w,
		Height: h,
		Cursor: NewCursor(0, 0),
		Tiles:  make([]*Tile, w*h),
	}
	b.Resize(w, h)
	return b
}

// Clear removes all Tiles from the screen
func (b *Buffer) Clear() {
	for o := range b.Tiles {
		b.Tiles[o] = nil
	}
}

// ClearAt clears a tile at offset o
func (b *Buffer) ClearAt(o int) {
	if o < len(b.Tiles) {
		b.Tiles[o] = nil
	}
}

// ClearFrom clears all Tiles from offset o
func (b *Buffer) ClearFrom(o int) {
	for l := len(b.Tiles); o < l; o++ {
		b.Tiles[o] = nil
	}
}

// ClearTo clears all Tiles up until offset o
func (b *Buffer) ClearTo(o int) {
	for o = math.MinInt(o, len(b.Tiles)); o >= 0; o-- {
		b.Tiles[o] = nil
	}
}

// Insert inserts n Tiles at offset o.
func (b *Buffer) Insert(o, n int) {
	p := make([]*Tile, n)
	b.Tiles = append(b.Tiles[:o], append(p, b.Tiles[o:]...)...)
}

// Expand buffer to fit offset o.
func (b *Buffer) Expand(o int) *Buffer {
	l := len(b.Tiles)
	if l <= o {
		b.Tiles = append(b.Tiles, make([]*Tile, o-l+1)...)
	}
	return b
}

// FromMemory takes a VGA text memory dump and converts it to tiles
func (b *Buffer) FromMemory(m []byte) (err error) {
	l := b.Height * b.Width * 2
	if len(m) < l {
		return errors.New("Insufficient data")
	}
	/*
		if len(m) != l {
			log.Printf("buffer: got %d bytes of memory, expected %d\n", len(m), l)
		}
	*/

	b.Expand(b.Height * b.Width)
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			// Memory offset
			mo := ((y * b.Width) << 1) + (x << 1)
			// Tile offset
			to := (y * b.Width) + x

			t := b.Tile(to)
			t.Attrib = 0
			t.Char = m[mo]
			t.Color = int(m[mo+1] & 0x0f)
			t.Background = int((m[mo+1] & 0xf0) >> 4)
		}
	}

	// Update boundaries
	b.maxWidth = b.Width
	b.maxHeight = b.Height

	return
}

// Len returns the number of possible Tiles (total offset)
func (b *Buffer) Len() int {
	return b.Width * b.Height
}

// Normalize fits the cursor within the canvas.
func (b *Buffer) Normalize() *Buffer {
	w, h := b.Size()
	b.Cursor.Normalize(w, h)
	return b
}

// Resize to at least a w x h canvas. If w exceeds the maximum buffer width an
// error is thrown.
func (b *Buffer) Resize(w, h int) error {
	if w*h > len(b.Tiles) {
		return fmt.Errorf("buffer: %dx%d=%d exceeds %d tiles", w, h, w*h, len(b.Tiles))
	}
	b.Width = w
	b.Height = h
	b.Tiles = b.Tiles[:w*h]
	return nil
}

// Size mathulates the allocated buffer size.
func (b *Buffer) Size() (w int, h int) {
	var l int
	l = b.Len()
	w = b.Width
	if w > 0 {
		h = 1 + ((l - 1) / w)
	}
	return
}

// SizeMax returns the actual used buffer size.
func (b *Buffer) SizeMax() (int, int) {
	return b.maxWidth, b.maxHeight
}

// SizeMaxToSize sets the actual used buffer size to the buffer dimensions.
func (b *Buffer) SizeMaxToSize() (int, int) {
	b.maxWidth = b.Width
	b.maxHeight = b.Height
	return b.SizeMax()
}

// Tile at offset o, will allocate a new Tile if it doesn't exist at the
// requested offset.
func (b *Buffer) Tile(o int) *Tile {
	if o >= len(b.Tiles) {
		return nil
	}
	if b.Tiles[o] == nil {
		b.Tiles[o] = NewTile()
	}
	return b.Tiles[o]
}

// TileAt retrieves tile at coorindates x, y.
func (b *Buffer) TileAt(x, y int) *Tile {
	return b.Tile((y * b.Width) + x)
}

// PutChar writes a character to the buffer at the current cursor location and
// advances the cursor position.
func (b *Buffer) PutChar(c byte) error {
	b.Cursor.Char = c
	o := b.Cursor.Offset(b.Width)
	t := b.Expand(o).Tile(o)
	t.Update(&b.Cursor.Tile)
	b.Cursor.X++
	b.Cursor.NormalizeAndWrap(b.Width)
	b.maxWidth = math.MaxInt(b.maxWidth, b.Cursor.X)
	b.maxHeight = math.MaxInt(b.maxHeight, b.Cursor.Y)
	return nil
}

// Image returns the buffer as an image
func (b *Buffer) Image(p color.Palette, f *font.Font) (m image.Image, err error) {
	w, h := b.SizeMax()

	dx := f.Size.X
	dy := f.Size.Y
	if b.Flags.LetterSpacing&sauce.LetterSpacing9Pixel > 0 {
		// Adjust for 9 pixel letter spacing
		dx++
	}
	dp := image.Pt(dx+1, dy)
	i := image.NewRGBA(image.Rect(0, 0, dx*w, dy*h))

	colors := make([]*image.Uniform, len(p))
	for i, c := range p {
		colors[i] = image.NewUniform(c)
	}

	/*
		log.Printf("buffer: draw %d x %d image for %d x %d buffer with %d colors\n",
			i.Rect.Max.X, i.Rect.Max.Y, w, h, len(colors))
	*/

	// Start with a black canvas
	draw.Draw(i, i.Bounds(), colors[0], image.ZP, draw.Src)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ox := x * dx
			oy := y * dy

			t := b.TileAt(x, y)

			p := image.Pt(ox, oy)
			r := image.Rectangle{p, p.Add(dp)}

			fg := t.Color
			bg := t.Background
			if t.Attrib&ATTRIB_BOLD > 0 && fg < 8 {
				fg += 8
			}
			if b.Flags.NonBlink && t.Attrib&ATTRIB_BLINK > 0 && bg < 8 {
				bg += 8
			}
			if t.Attrib&ATTRIB_NEGATIVE == ATTRIB_NEGATIVE {
				fg, bg = bg, fg
			}

			// Background
			if bg > 0 {
				draw.Draw(i, r, colors[bg], image.ZP, draw.Src)
			}

			// Foreground
			if fg != bg && t.Char != 0x20 {
				mr := f.BoundsFor(t.Char)
				draw.DrawMask(i, mr.Sub(mr.Min).Add(p), colors[fg], image.ZP, f.Mask, mr.Min, draw.Over)
			}

			if t.Attrib&ATTRIB_UNDERLINE > 0 {
				for xx := ox; xx < ox+dx; xx++ {
					i.Set(xx, oy+f.Size.Y, colors[fg])
				}
			}
		}
	}

	return i, nil
}
