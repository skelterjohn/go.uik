package uik

import (
	"image"
	"github.com/skelterjohn/geom"
)

// type Coord struct {
// 	X, Y float64
// }

// type Bounds struct {
// 	Min, Max Coord
// }

// func (b Bounds) Contains(c Coord) bool {
// 	return c.X >= b.Min.X && c.Y >= b.Min.Y && c.X <= b.Max.X && c.Y <= b.Max.Y
// }

func RectangleForRect(b geom.Rect) (r image.Rectangle) {
	r.Min.X = int(b.Min.X)
	r.Max.X = int(b.Max.X)
	r.Min.Y = int(b.Min.Y)
	r.Max.Y = int(b.Max.Y)
	return
}
