package parser

import (
	"image"
	"io"

	"git.maze.io/maze/go-piece/font"
	sauce "git.maze.io/maze/go-sauce"
)

// Parser implements a parser for artscene pieces
type Parser interface {
	HTML(full bool) string
	String() string
	Font() *font.Font
	Image(*font.Font) (image.Image, error)
	Parse(io.Reader) error
	Width() int
	Height() int
	SAUCE() *sauce.SAUCE
}
