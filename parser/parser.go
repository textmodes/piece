package parser

import (
	"image"
	"io"

	"github.com/textmodes/piece/font"
)

// Parser implements a parser for artscene pieces
type Parser interface {
	HTML() (string, error)
	Font() *font.Font
	Image(*font.Font) (image.Image, error)
	Parse(io.Reader) error
}
