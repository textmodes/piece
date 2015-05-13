package color

import "math"

func sq(v float64) float64 {
	return v * v
}

func cub(v float64) float64 {
	return sq(v) * v
}

// DistanceCIE76 measures the visual distance between two colors using the CIE76 formula.
func (c Color) DistanceCIE76(o Color) float64 {
	l1, a1, b1 := c.Tolab()
	l2, a2, b2 := o.Tolab()
	return math.Sqrt(sq(l1-l2) + sq(a1-a2) + sq(b1-b2))
}

// DistanceCIE94 measures the visual distance between two colors using the CIE94 formula.
func (c Color) DistanceCIE94(o Color) float64 {
	l1, a1, b1 := c.Tolab()
	l2, a2, b2 := o.Tolab()

	kl := 1.0
	k1 := 0.045
	k2 := 0.015

	dL := l1 - l2
	c1 := math.Sqrt(sq(a1) + sq(b1))
	c2 := math.Sqrt(sq(a2) + sq(b2))
	dCab := c1 - c2
	dHab := math.Sqrt(sq(a1-a2) + sq(b1-b2) + sq(dCab))
	sl := 1.0
	sc := 1.0 + k1*c1
	sh := 1.0 + k2*c1

	return math.Sqrt(sq(dL/(kl*sl)) + sq(dCab/sc) + sq(dHab/sh))
}
