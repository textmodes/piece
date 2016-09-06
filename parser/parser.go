package parser

import (
	"errors"
	"image"
	"io"

	"git.maze.io/maze/go-piece/font"
	sauce "git.maze.io/maze/go-sauce"
)

var ErrNotSupported = errors.New(`piece: not supported`)

// Parser implements a parser for artscene pieces
type Parser interface {
	HTML(full bool) (string, error)
	String() string
	Font() *font.Font
	Image(*font.Font) (image.Image, error)
	Parse(io.Reader) error
	Width() int
	Height() int
	SAUCE() *sauce.SAUCE
}
