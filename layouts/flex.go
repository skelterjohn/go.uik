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

package layouts

import (
	"github.com/skelterjohn/go.uik"
	"math"
)

func r(x ...interface{}) {
	if true {
		// fmt.Print("flex ")
		uik.Report(x...)
	}
}

type elem struct {
	index, extra               int
	minSize, prefSize, maxSize float64
	sizes                      [3]float64
}

func (e *elem) fix() {
	e.sizes[0] = e.minSize
	e.sizes[1] = e.prefSize
	e.sizes[2] = e.maxSize
}

type flex struct {
	elemsets                   []map[*elem]bool
	minSize, prefSize, maxSize float64
	widths, mins, maxs         []float64
	length                     float64
}

func (f *flex) has(e *elem) (ok bool) {
	elems := f.elemsets[e.index]
	ok = elems[e]
	return
}

func (f *flex) add(e *elem) {
	e.sizes[0] = e.minSize
	e.sizes[1] = e.prefSize
	e.sizes[2] = e.maxSize
	if e.index+e.extra >= len(f.elemsets) {
		nelems := make([]map[*elem]bool, e.index+e.extra+1)
		copy(nelems, f.elemsets)
		f.elemsets = nelems
	}
	if f.elemsets[e.index] == nil {
		f.elemsets[e.index] = map[*elem]bool{}
	}
	f.elemsets[e.index][e] = true
}

func (f *flex) rem(e *elem) {
	elems := f.elemsets[e.index]
	if elems[e] {
		delete(elems, e)
	}
}

func (f *flex) constrain(length float64) (widths, mins, maxs []float64) {

	// r("checking length", length, f.oldLength)
	// first check if the last settings are still ok - better to not change if possible

	if false && length == f.length {
		var satisfied = true
	loop:
		for _, elems := range f.elemsets {
			for elem := range elems {
				if elem.index >= len(f.mins) {
					satisfied = false
					break loop
				}
				min := f.mins[elem.index]
				if elem.index+elem.extra >= len(f.maxs) {
					satisfied = false
					break loop
				}
				max := f.maxs[elem.index+elem.extra]
				w := max - min
				if w < elem.minSize || w > elem.maxSize {
					satisfied = false
					break loop
				}
			}
		}
		if false && satisfied {

			widths = f.widths
			mins = f.mins
			maxs = f.maxs
			return
		}
	}

	// r("constrain", length)
	type item struct {
		extra                      int
		minSize, prefSize, maxSize float64
	}
	var curItems []item

	var total float64
	widths = make([]float64, len(f.elemsets))
	for index, elems := range f.elemsets {
		for elem := range elems {
			// fmt.Println(elem)
			curItems = append(curItems, item{
				extra:    elem.extra,
				minSize:  elem.minSize,
				prefSize: elem.prefSize,
				maxSize:  elem.maxSize,
			})
		}

		var width float64
		for _, item := range curItems {
			if item.extra == 0 {
				width = math.Max(width, item.prefSize)
			}
		}
		widths[index] = width
		total += width

		for i := range curItems {
			curItems[i].extra--
			curItems[i].minSize -= width
			curItems[i].prefSize -= width
			curItems[i].maxSize -= width
		}
	}

	diff := length - total

	avg := diff / float64(len(f.elemsets))
	curItems = nil

	for index, elems := range f.elemsets {
		for elem := range elems {
			curItems = append(curItems, item{
				extra:    elem.extra,
				minSize:  elem.minSize,
				prefSize: elem.prefSize,
				maxSize:  elem.maxSize,
			})
		}

		nw := widths[index] + avg
		for _, item := range curItems {
			if item.extra < 0 {
				continue
			}
			if diff > 0 {
				nw = math.Min(nw, item.maxSize)
			}
			if diff < 0 {
				nw = math.Max(nw, item.minSize)
			}
		}
		diff -= nw - widths[index]
		widths[index] = nw

		if diff == 0 {
			break
		}

		for i := range curItems {
			curItems[i].extra--
			curItems[i].minSize -= widths[index]
			curItems[i].prefSize -= widths[index]
			curItems[i].maxSize -= widths[index]
		}
	}

	mins = make([]float64, len(widths))
	maxs = make([]float64, len(widths))
	left := 0.0
	for i, w := range widths {
		mins[i] = left
		left += w
		maxs[i] = left
	}

	f.length = length
	f.widths = append([]float64{}, widths...)
	f.mins = append([]float64{}, mins...)
	f.maxs = append([]float64{}, maxs...)

	// r("giving", widths)

	return
}

func (f *flex) makePref(which int) (size float64) {
	type item struct {
		extra int
		size  float64
	}
	var curItems []item

	for _, elems := range f.elemsets {
		for elem := range elems {
			curItems = append(curItems, item{elem.extra, elem.sizes[which]})
		}

		var width float64
		for _, item := range curItems {
			if item.extra == 0 {
				width = math.Max(width, item.size)
			}
		}
		size += width

		for i := range curItems {
			curItems[i].extra--
			curItems[i].size -= width
		}
	}

	return
}

func (f *flex) makePrefs() (min, pref, max float64) {
	min = f.makePref(0)
	pref = f.makePref(1)
	max = f.makePref(2)

	// r("pref", min, pref, max)

	return
}
