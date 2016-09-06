//go:generate go-bindata -pkg font -prefix data -o data.go data
package font

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	Transparent = color.Alpha{0x00}
	Opaque      = color.Alpha{0xff}
)

var builtin *Collection

type Font struct {
	image.Image
	Size image.Point
	Mask image.Image
}

type Collection struct {
	Font map[string]map[image.Point]*Font
}

// New loads a new font from disk.
func New(name string, size image.Point) (f *Font, err error) {
	var r *os.File
	r, err = os.Open(name)
	if err != nil {
		return
	}
	defer r.Close()
	return NewReader(r, size)
}

// NewReader loads a new font from an io.Reader.
func NewReader(r io.Reader, size image.Point) (f *Font, err error) {
	f = &Font{
		Size: size,
	}
	f.Image, _, err = image.Decode(r)
	if err != nil {
		return nil, err
	}

	return f, f.setMask()
}

func (f *Font) setMask() (err error) {
	switch src := f.Image.(type) {
	case *image.Gray, *image.Paletted:
		b := f.Image.Bounds()
		m := image.NewAlpha(f.Image.Bounds())
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				cr, _, _, _ := src.At(x, y).RGBA()
				if cr > 0x7f {
					m.SetAlpha(x, y, Opaque)
				}
			}
		}
		f.Mask = m
	case *BitMask:
		f.Mask = f.Image
	default:
		err = fmt.Errorf("Image type %T not supported", f.Image)
	}

	return
}

// BoundsFor returns the bounds for a glyph.
func (f *Font) BoundsFor(c byte) image.Rectangle {
	p0 := image.Pt(f.Size.X*int(c), 0)
	p1 := image.Pt(f.Size.X+p0.X, f.Size.Y)
	return image.Rectangle{p0, p1}
}

// NewCollection initialises a new, empty collection.
func NewCollection() *Collection {
	return &Collection{
		Font: make(map[string]map[image.Point]*Font, 0),
	}
}

// Add a font to the collection.
func (c *Collection) Add(name string, size image.Point, font *Font) {
	if _, ok := c.Font[name]; !ok {
		c.Font[name] = make(map[image.Point]*Font)
	}
	c.Font[name][size] = font
}

// Get a font from the collection.
func (c *Collection) Get(name string, size image.Point) *Font {
	if c.Font[name] == nil {
		if c.Font[fontAlias[name]] != nil {
			name = fontAlias[name]
		}
	}
	if c.Font[name] != nil {
		for s, font := range c.Font[name] {
			if s.X == size.X && s.Y == size.Y {
				return font
			}
		}
	}
	return nil
}

// GetSAUCE gets a font from the collection by SAUCE font name.
func (c *Collection) GetSAUCE(name string) *Font {
	if fontAlias[name] == "" {
		return nil
	}
	name = fontAlias[name]
	size := fontSize[name]
	if size.X == 0 {
		return nil
	}
	return c.Get(name, size)
}

// Len returns the number of fonts in the collection.
func (c *Collection) Len() int {
	var l int
	for _, fonts := range c.Font {
		l += len(fonts)
	}
	return l
}

// LoadAll loads all font masks from the given directory.
//
// All fonts must have the pattern: <name>-<width>x<height>.png
func (c *Collection) LoadAll(path string) error {
	matches, err := filepath.Glob(filepath.Join(path, "*.png"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		size, name, err := fontInfo(match)
		if err != nil {
			return err
		}
		font, err := New(match, size)
		if err != nil {
			return err
		}
		c.Add(name, size, font)
		log.Printf("font: %s [%dx%d]\n", name, size.X, size.Y)
		if fontAlias[name] != "" {
			log.Printf("font: %s -> %s\n", name, fontAlias[name])
		}
	}

	return nil
}

func fontInfo(path string) (size image.Point, name string, err error) {
	ext := filepath.Ext(path)
	base := filepath.Base(path[:len(path)-len(ext)])
	part := strings.Split(base, "-")
	if len(part) < 2 {
		err = fmt.Errorf("font: invalid file name %q: no separator", path)
		return
	}
	name = strings.Join(part[:len(part)-1], "-")
	size = image.Point{}
	if size, err = ParseSize(part[len(part)-1]); err != nil {
		err = fmt.Errorf("font: invalid file name %q: %v", path, err)
		return
	}
	return
}

// loadBuiltin loads the built in collection of fonts.
func loadBuiltin() (*Collection, error) {
	c := NewCollection()
	for _, match := range AssetNames() {
		size, name, err := fontInfo(match)
		if err != nil {
			return nil, err
		}
		data, err := Asset(match)
		if err != nil {
			return nil, err
		}
		font, err := NewReader(bytes.NewBuffer(data), size)
		if err != nil {
			return nil, err
		}
		c.Add(name, size, font)
	}
	return c, nil
}

// Add a font to the builtin collection
func Add(name string, size image.Point, font *Font) {
	builtin.Add(name, size, font)
}

// Get a builtin font
func Get(name string, size image.Point) *Font {
	return builtin.Get(name, size)
}

// GetSAUCE loads a builtin font from a SAUCE font name
func GetSAUCE(name string) *Font {
	return builtin.GetSAUCE(name)
}

func init() {
	var err error
	if builtin, err = loadBuiltin(); err != nil {
		panic(err)
	}
}
