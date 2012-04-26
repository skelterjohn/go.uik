package uik

import (
	"github.com/skelterjohn/geom"
	"image"
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

type RectSet []geom.Rect

func (rs RectSet) Translate(offset geom.Coord) (nrs RectSet) {
	nrs = make(RectSet, len(rs))
	for i, r := range rs {
		nrs[i] = r
		nrs[i].Translate(offset)
	}
	return
}

func (rs RectSet) Intersection(r geom.Rect) (nrs RectSet) {
	for _, x := range rs {
		if geom.RectsIntersect(x, r) {
			nrs = append(nrs, x)
		}
	}
	return
}

func (rs RectSet) Intersects(r geom.Rect) bool {
	for _, x := range rs {
		if geom.RectsIntersect(x, r) {
			return true
		}
	}
	return false
}
