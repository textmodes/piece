package buffer

import "git.maze.io/maze/go-piece/math"

type Cursor struct {
	X, Y int
	Tile
}

func NewCursor(x, y int) *Cursor {
	return &Cursor{
		X: x,
		Y: y,
		Tile: Tile{
			Char:       TILE_DEFAULT_CHAR,
			Color:      TILE_DEFAULT_COLOR,
			Background: TILE_DEFAULT_BACKGROUND,
		},
	}
}

func (c *Cursor) Advance(w int) *Cursor {
	c.X++
	if c.X >= w {
		c.Y++
		c.X = 0
	}
	return c
}

// Goto moves the cursor to the requested coordinates.
func (c *Cursor) Goto(x, y int) *Cursor {
	c.X = math.MaxInt(0, x)
	c.Y = math.MaxInt(0, y)
	return c
}

// Normalize ensures cursor is inside the bounding box.
func (c *Cursor) Normalize(w, h int) *Cursor {
	c.X = math.MaxInt(0, math.MinInt(c.X, w))
	c.Y = math.MaxInt(0, math.MinInt(c.Y, h))
	return c
}

// NormalizeAndWrap ensures cursor is inside the bounding box, and wraps to the next line if there is a horizontal overflow.
func (c *Cursor) NormalizeAndWrap(w int) *Cursor {
	var dy int
	dy, c.X = math.DivMod(c.X, w)
	c.Y += dy
	return c.Normalize(w, c.Y)
}

// Pos returns the x, y coorindates of the cursor
func (c *Cursor) Pos() (int, int) {
	return c.X, c.Y
}

func (c *Cursor) Up(i int) {
	c.Y = math.MaxInt(0, c.Y-i)
}

func (c *Cursor) Down(i int) {
	c.Y += i
}

func (c *Cursor) Left(i int) {
	c.X = math.MaxInt(0, c.X-i)
}

func (c *Cursor) Right(i int) {
	c.X += i
}

func (c *Cursor) Offset(w int) int {
	return (c.Y * w) + c.X
}
