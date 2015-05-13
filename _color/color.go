package color

import "fmt"

const (
	RGB_GAMMA = 2.2
)

// Color contains a 64 bit precision representation of a color in sRGBA (standard RGBA) values in the range 0.0-1.0
type Color struct {
	R, G, B, A float64
}

// Opaque sets the color to opaque.
func (c Color) Opaque() Color {
	c.A = 0.0
	return c
}

// Transparent sets the color to transparent.
func (c Color) Transparent() Color {
	c.A = 1.0
	return c
}

// ARGB returns the 32-bit ARGB color.
func (c Color) ARGB() uint32 {
	r, g, b, a := c.Color8()
	return (uint32(a) << 24) | (uint32(r) << 16) | (uint32(g) << 8) | uint32(b)
}

// RGBA returns the 32-bit RGBA color.
func (c Color) RGBA() uint32 {
	r, g, b, a := c.Color8()
	return (uint32(r) << 24) | (uint32(g) << 16) | (uint32(b) << 8) | uint32(a)
}

// Color implements the Go color.Color interface.
func (c Color) Color() (r, g, b, a uint32) {
	r = uint32(c.R * 65535.0)
	g = uint32(c.G * 65535.0)
	b = uint32(c.B * 65535.0)
	a = uint32(c.A * 65535.0)
	return
}

// Color8 returns the color in 8 bit precision values.
func (c Color) Color8() (r, g, b, a uint8) {
	r = uint8(c.R*255.0 + .5)
	g = uint8(c.G*255.0 + .5)
	b = uint8(c.B*255.0 + .5)
	a = uint8(c.A*255.0 + .5)
	return
}

// Hex calculations

// Hex converts from a "web" hex encoded color
func Hex(h string) (Color, error) {
	format := "#%02x%02x%02x"
	factor := 1.0 / 255.0
	if len(h) == 4 { // Abbreviated hex
		format = "#%1x%1x%1x"
		factor = 1.0 / 15.0
	}

	var r, g, b uint8
	n, err := fmt.Sscanf(h, format, &r, &g, &b)
	if err != nil {
		return Color{}, err
	}
	if n != 3 {
		return Color{}, fmt.Errorf("Color %q is not a hex color", h)
	}

	return Color{
		R: float64(r) * factor,
		G: float64(g) * factor,
		B: float64(b) * factor,
		A: 1.0,
	}, nil
}

// Hex returns the hexadecimal "web" color representation
func (c Color) Hex() string {
	rgb := c.ARGB() & 0xffffff
	return fmt.Sprintf("#%06x", rgb)
}

// WebRGB returns the  RGB color, suitable for use in CSS
func (c Color) WebRGB() string {
	r, g, b, _ := c.Color8()
	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

// WebRGBA returns the RGBA color, suitable for use in CSS
func (c Color) WebRGBA() string {
	r, g, b, a := c.Color8()
	return fmt.Sprintf("rgba(%d, %d, %d, %d)", r, g, b, a)
}
