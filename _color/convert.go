package color

import "math"

// http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/
// http://www.brucelindbloom.com/Eqn_RGB_to_XYZ.html

func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// LinearRGB converts the color into the linear RGB space (see http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/).
func (c Color) LinearRGB() (r, g, b float64) {
	r = linearize(c.R)
	g = linearize(c.G)
	b = linearize(c.B)
	return
}

// FastLinearRGB is much faster than and almost as accurate as LinearRgb.
func (c Color) FastLinearRGB() (r, g, b float64) {
	r = math.Pow(c.R, 2.2)
	g = math.Pow(c.G, 2.2)
	b = math.Pow(c.B, 2.2)
	return
}

func delinearize(v float64) float64 {
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

// LinearRgb creates an sRGBA color out of the given linear RGB color (see http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/).
func LinearRGB(r, g, b float64) Color {
	return Color{delinearize(r), delinearize(g), delinearize(b), 0}
}

// FastLinearRgb is much faster than and almost as accurate as LinearRgb.
func FastLinearRGB(r, g, b float64) Color {
	return Color{math.Pow(r, 1.0/2.2), math.Pow(g, 1.0/2.2), math.Pow(b, 1.0/2.2), 0}
}

// XYZToLinearRGB converts from CIE XYZ-space to Linear RGB space.
func XYZToLinearRGB(x, y, z float64) (r, g, b float64) {
	r = 3.2404542*x - 1.5371385*y - 0.4985314*z
	g = -0.9692660*x + 1.8760108*y + 0.0415560*z
	b = 0.0556434*x - 0.2040259*y + 1.0572252*z
	return
}

func LinearRGBToXYZ(r, g, b float64) (x, y, z float64) {
	x = 0.4124564*r + 0.3575761*g + 0.1804375*b
	y = 0.2126729*r + 0.7151522*g + 0.0721750*b
	z = 0.0193339*r + 0.1191920*g + 0.9503041*b
	return
}

// XYZ colors, http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/

// ToXYZ converts the sRGB color to XYZ color space.
func (c Color) ToXYZ() (x, y, z float64) {
	return LinearRGBToXYZ(c.LinearRGB())
}

// XYZ converts the XYZ color to sRGB color space.
func XYZ(x, y, z float64) Color {
	return LinearRGB(XYZToLinearRGB(x, y, z))
}

// xyY colors, http://www.brucelindbloom.com/Eqn_XYZ_to_xyY.html

// ToxyY converts the sRGB color to xyY color space.
func (c Color) ToxyY() (x, y, Y float64) {
	return XYZToxyY(c.ToXYZ())
}

func XYZToxyY(X, Y, Z float64) (x, y, Yout float64) {
	return XYZToxyYWhiteRef(X, Y, Z, D65)
}

func XYZToxyYWhiteRef(X, Y, Z float64, white [3]float64) (x, y, Yout float64) {
	Yout = Y
	N := X + Y + Z
	if math.Abs(N) < 1e-14 {
		// When we have black, Bruce Lindbloom recommends to use
		// the reference white's chromacity for x and y.
		x = white[0] / (white[0] + white[1] + white[2])
		y = white[1] / (white[0] + white[1] + white[2])
	} else {
		x = X / N
		y = Y / N
	}
	return
}

func XyYToXYZ(x, y, Y float64) (X, Yout, Z float64) {
	Yout = Y
	if -1e-14 < y && y < 1e-14 {
		X = 0.0
		Y = 0.0
	} else {
		X = Y / y * x
		Z = Y / y * (1.0 - x - y)
	}
	return
}

// L*a*b* color space, http://en.wikipedia.org/wiki/Lab_color_space#CIELAB-CIEXYZ_conversions

func (c Color) Tolab() (l, a, b float64) {
	return XYZToLab(c.ToXYZ())
}

func lab_f(t float64) float64 {
	if t > 6.0/29.0*6.0/29.0*6.0/29.0 {
		return math.Cbrt(t)
	}
	return t/3.0*29.0/6.0*29.0/6.0 + 4.0/29.0
}

// XYZToLab converts a color from XYZ color space to L*a*b* color space using D65 as the white reference point.
func XYZToLab(x, y, z float64) (l, a, b float64) {
	return XYZToLabWhiteRef(x, y, z, D65)
}

// XYZToLabWhiteRef converts a color from XYZ color space to L*a*b* color space using a white reference point.
func XYZToLabWhiteRef(x, y, z float64, white [3]float64) (l, a, b float64) {
	fy := lab_f(y / white[1])
	l = 1.16*fy - 0.16
	a = 5.0 * (lab_f(x/white[0]) - fy)
	b = 2.0 * (fy - lab_f(z/white[2]))
	return
}

func lab_finv(t float64) float64 {
	if t > 6.0/29.0 {
		return t * t * t
	}
	return 3.0 * 6.0 / 29.0 * 6.0 / 29.0 * (t - 4.0/29.0)
}

// LabToXYZ converts a color from L*a*b* color space to XYZ color space using D65 as the white reference point.
func LabToXYZ(l, a, b float64) (x, y, z float64) {
	// D65 white
	return LabToXYZWhiteRef(l, a, b, D65)
}

// LabToXYZWhiteRef converts a color from L*a*b* color space to XYZ color space using a white reference point.
func LabToXYZWhiteRef(l, a, b float64, white [3]float64) (x, y, z float64) {
	l2 := (l + 0.16) / 1.16
	x = white[0] * lab_finv(l2+a/5.0)
	y = white[1] * lab_finv(l2)
	z = white[2] * lab_finv(l2-b/2.0)
	return
}
