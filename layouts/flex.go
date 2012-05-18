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
	"fmt"
	"math"
)

func r(x ...interface{}) {
	if true {
		// fmt.Print("flex ")
		fmt.Println(x...)
	}
}

type elem struct {
	index, extra               int
	minSize, prefSize, maxSize float64
	sizes                      [3]float64
}

type flex struct {
	elemsets                    [][]*elem
	minSize, prefSize, maxSize  float64
	oldWidths, oldMins, oldMaxs []float64
	oldLength                   float64
}

func (f *flex) add(e *elem) {
	e.sizes[0] = e.minSize
	e.sizes[1] = e.prefSize
	e.sizes[2] = e.maxSize
	if e.index+e.extra >= len(f.elemsets) {
		nelems := make([][]*elem, e.index+e.extra+1)
		copy(nelems, f.elemsets)
		f.elemsets = nelems
	}
	f.elemsets[e.index] = append(f.elemsets[e.index], e)
	// f.oldWidths = nil
}

func (f *flex) constrain(length float64) (widths, mins, maxs []float64) {

	r("checking length", length, f.oldLength)
	// first check if the last settings are still ok - better to not change if possible
	if length == f.oldLength {
		satisfied := true

		if f.oldWidths != nil {
		loop:
			for _, elems := range f.elemsets {
				for _, elem := range elems {
					// r(elem.index, elem.extra)
					if elem.index >= len(f.oldMins) {
						satisfied = false
						r("extra min")
						break loop
					}
					// r("index ok", elem.index)
					emin := f.oldMins[elem.index]
					if elem.index+elem.extra >= len(f.oldMaxs) {
						satisfied = false
						r("extra max")
						break loop
					}
					// r("index+extra ok", elem.index+elem.extra)
					emax := f.oldMaxs[elem.index+elem.extra]
					width := emax - emin
					if width < elem.minSize || width > elem.maxSize {
						satisfied = false
						r("too small")
						break loop
					}
				}
			}

			if satisfied {
				//r("satisfied", f.oldWidths)
				widths, mins, maxs = f.oldWidths, f.oldMins, f.oldMaxs
				return
			}
		}
	} else {
		r("wrong length", length, f.oldLength)
	}
	f.oldLength = length
	r("setting length", length, f.oldLength)

	// uik.Report("constrain", length)
	type item struct {
		extra                      int
		minSize, prefSize, maxSize float64
	}
	var curItems []item

	var total float64
	widths = make([]float64, len(f.elemsets))
	for index, elems := range f.elemsets {
		for _, elem := range elems {
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

	if diff > 0 {
		avg := diff / float64(len(f.elemsets))
		curItems = nil
		for index, elems := range f.elemsets {
			// fmt.Println(diff, "to give")
			for _, elem := range elems {
				curItems = append(curItems, item{
					extra:    elem.extra,
					minSize:  elem.minSize,
					prefSize: elem.prefSize,
					maxSize:  elem.maxSize,
				})
			}

			nw := widths[index] + avg
			for _, item := range curItems {
				// fmt.Println(item)
				if item.extra < 0 {
					continue
				}
				nw = math.Min(nw, item.maxSize)
			}
			// fmt.Println("nw", nw)
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
	}

	if diff < 0 {
		avg := diff / float64(len(f.elemsets))
		curItems = nil
		for index, elems := range f.elemsets {
			// fmt.Println(diff, "to give")
			for _, elem := range elems {
				curItems = append(curItems, item{
					extra:    elem.extra,
					minSize:  elem.minSize,
					prefSize: elem.prefSize,
					maxSize:  elem.maxSize,
				})
			}

			nw := widths[index] + avg
			for _, item := range curItems {
				// fmt.Println(item)
				if item.extra < 0 {
					continue
				}
				nw = math.Max(nw, item.minSize)
			}
			// fmt.Println("nw", nw)
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
	}

	mins = make([]float64, len(widths))
	maxs = make([]float64, len(widths))
	left := 0.0
	for i, w := range widths {
		mins[i] = left
		left += w
		maxs[i] = left
	}

	f.oldWidths = append([]float64{}, widths...)
	f.oldMins = append([]float64{}, mins...)
	f.oldMaxs = append([]float64{}, maxs...)

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
		for _, elem := range elems {
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
	return
}
