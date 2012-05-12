/*
   Copyright 2012 the go.uik authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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

func (rs RectSet) IntersectsStrict(r geom.Rect) bool {
	for _, x := range rs {
		if geom.RectsIntersectStrict(x, r) {
			return true
		}
	}
	return false
}
