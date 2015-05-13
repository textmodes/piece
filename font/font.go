package font

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	Transparent = color.Alpha{0x00}
	Opaque      = color.Alpha{0xff}
)

type Font struct {
	image.Image
	Size image.Point
	Mask image.Image
}

type Collection struct {
	Font map[string]map[image.Point]*Font
}

func New(name string, size image.Point) (f *Font, err error) {
	var r *os.File
	r, err = os.Open(name)
	if err != nil {
		return
	}
	defer r.Close()

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

func (f *Font) BoundsFor(c byte) image.Rectangle {
	p0 := image.Pt(f.Size.X*int(c), 0)
	p1 := image.Pt(f.Size.X+p0.X, f.Size.Y)
	return image.Rectangle{p0, p1}
}

func NewCollection() *Collection {
	return &Collection{
		Font: make(map[string]map[image.Point]*Font, 0),
	}
}

func (c *Collection) Find(name string, size image.Point) *Font {
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
		if c.Font[name] == nil {
			c.Font[name] = make(map[image.Point]*Font, 0)
		}
		c.Font[name][size] = font
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
		err = fmt.Errorf("Invalid file name %q: no separator", path)
		return
	}
	name = strings.Join(part[:len(part)-1], "-")
	size = image.Point{}
	part = strings.SplitN(part[len(part)-1], "x", 2)
	if len(part) != 2 {
		err = fmt.Errorf("Invalid file name %q: no size", path)
		return
	}
	if size.X, err = strconv.Atoi(part[0]); err != nil {
		return
	}
	if size.Y, err = strconv.Atoi(part[1]); err != nil {
		return
	}
	return
}
