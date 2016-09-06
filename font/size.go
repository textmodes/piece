package font

import (
	"fmt"
	"image"
	"strconv"
	"strings"
)

var fontSize = map[string]image.Point{
	"Amiga MicroKnight":  image.Pt(8, 16),
	"Amiga MicroKnight+": image.Pt(8, 16),
	"Amiga mOsOul":       image.Pt(8, 16),
	"Amiga P0T-NOoDLE":   image.Pt(8, 16),
	"Amiga Topaz 1":      image.Pt(8, 16),
	"Amiga Topaz 1+":     image.Pt(8, 16),
	"Amiga Topaz 2":      image.Pt(8, 16),
	"Amiga Topaz 2+":     image.Pt(8, 16),
	"Atari ATASCII":      image.Pt(8, 16),
	"IBM VGA":            image.Pt(8, 16),
	"IBM VGA50":          image.Pt(8, 8),
}

func ParseSize(size string) (pnt image.Point, err error) {
	part := strings.SplitN(size, "x", 2)
	if len(part) != 2 {
		err = fmt.Errorf("font: invalid size %q, expected \"<width>x<height>\"", size)
		return
	}

	if pnt.X, err = strconv.Atoi(part[0]); err != nil {
		return
	}

	if pnt.Y, err = strconv.Atoi(part[1]); err != nil {
		return
	}

	return pnt, nil
}
