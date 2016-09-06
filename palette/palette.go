package palette

import (
	"fmt"
	"image/color"
	"math/rand"
)

type Palette color.Palette

func (p Palette) Copy() Palette {
	var n = make(Palette, len(p))
	for i, c := range p {
		n[i] = c
	}
	return n
}

func (p *Palette) HTMLStyle() (string, string) {
	if IsBuiltin(*p) {
		return "", p.HTMLClass()
	}
	var style, class string
	class = fmt.Sprintf("p%04d", rand.Int()%10000)
	style += `<style type="text/css">`
	for i, c := range *p {
		r, g, b, _ := c.RGBA()
		rgbhex := fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
		style += fmt.Sprintf(`.%s .f%02x{color:%s} `, class, i, rgbhex)
		style += fmt.Sprintf(`.%s .b%02x{background-color:%s} `, class, i, rgbhex)
	}
	style += `</style>`
	return style, class
}

func (p *Palette) HTMLClass() string {
	switch p {
	case &CGA:
		return "cga"
	case &VGA:
		return "vga"
	}
	return ""
}

var CGA = Palette{
	color.RGBA{0x00, 0x00, 0x00, 0xff}, // Black
	color.RGBA{0xaa, 0x00, 0x00, 0xff}, // Red
	color.RGBA{0x00, 0xaa, 0x00, 0xff}, // Green
	color.RGBA{0xaa, 0x55, 0x00, 0xff}, // Brown
	color.RGBA{0x00, 0x00, 0xaa, 0xff}, // Blue
	color.RGBA{0xaa, 0x00, 0xaa, 0xff}, // Magenta
	color.RGBA{0x00, 0xaa, 0xaa, 0xff}, // Cyan
	color.RGBA{0xaa, 0xaa, 0xaa, 0xff}, // White (or gray)
	color.RGBA{0x55, 0x55, 0x55, 0xff}, // Bright black (or dark gray)
	color.RGBA{0xff, 0x55, 0x55, 0xff}, // Bright red
	color.RGBA{0x55, 0xff, 0x55, 0xff}, // Bright green
	color.RGBA{0xff, 0xff, 0x55, 0xff}, // Bright yellow
	color.RGBA{0x55, 0x55, 0xff, 0xff}, // Bright blue
	color.RGBA{0xff, 0x55, 0xff, 0xff}, // Bright magenta
	color.RGBA{0x55, 0xff, 0xff, 0xff}, // Bright cyan
	color.RGBA{0xff, 0xff, 0xff, 0xff}, // Bright white
}

var VGA = Palette{}

func init() {
	// Initialize the VGA palette
	for i := 0; i < 16; i++ {
		VGA = append(VGA, CGA[i])
	}

	// Next add a 6x6x6 color cube
	for r := uint8(0); r < 6; r++ {
		for g := uint8(0); g < 6; g++ {
			for b := uint8(0); b < 6; b++ {
				VGA = append(VGA, color.RGBA{
					0x37 + r*0x28,
					0x37 + g*0x28,
					0x37 + b*0x28,
					0xff,
				})
			}
		}
	}

	// And finally the gray scale ramp
	for i := uint8(0); i < 24; i++ {
		g := 10*i + 8
		VGA = append(VGA, color.RGBA{g, g, g, 0xff})
	}
}

func IsBuiltin(p Palette) bool {
	return &p == &CGA || &p == &VGA
}

func init() {
	// Initialize the VGA palette
	for i := 0; i < 16; i++ {
		VGA = append(VGA, CGA[i])
	}

	// Next add a 6x6x6 color cube
	for r := uint8(0); r < 6; r++ {
		for g := uint8(0); g < 6; g++ {
			for b := uint8(0); b < 6; b++ {
				VGA = append(VGA, color.RGBA{
					0x37 + r*0x28,
					0x37 + g*0x28,
					0x37 + b*0x28,
					0xff,
				})
			}
		}
	}

	// And finally the gray scale ramp
	for i := uint8(0); i < 24; i++ {
		g := 10*i + 8
		VGA = append(VGA, color.RGBA{g, g, g, 0xff})
	}
}
