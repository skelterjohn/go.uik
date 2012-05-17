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
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image/draw"
	"math"
)

type Flow struct {
	uik.Foundation

	childIndices map[*uik.Block]int

	count int

	sizeHint uik.SizeHint

	Add    chan *uik.Block
	Remove chan *uik.Block
}

func NewFlow() (f *Flow) {
	f = new(Flow)
	f.Initialize()

	go f.HandleEvents()

	return
}

func (f *Flow) Initialize() {
	f.Foundation.Initialize()
	f.DrawOp = draw.Over
	f.Add = make(chan *uik.Block, 10)
	f.Remove = make(chan *uik.Block, 10)
	f.childIndices = map[*uik.Block]int{}
}

func (f *Flow) reflow() {
	children := make([]*uik.Block, f.count)
	for child, i := range f.childIndices {
		children[i] = child
	}

	renderSize := f.Size

	renderSize.X = math.Max(f.sizeHint.MinSize.X, renderSize.X)
	renderSize.Y = math.Max(f.sizeHint.MinSize.X, renderSize.Y)

	ratioX := 1.0
	if renderSize.X < f.sizeHint.PreferredSize.X {
		ratioX = renderSize.X / f.sizeHint.PreferredSize.X
	}

	var left float64
	for i := 0; i < f.count; i++ {
		child := children[i]
		csh, ok := f.ChildrenHints[child]
		if !ok {
			//println("skip", child)
			continue
		}
		cbounds := geom.Rect{geom.Coord{left, 0}, geom.Coord{}}
		if csh.PreferredSize.Y <= renderSize.Y {
			cbounds.Max.Y = csh.PreferredSize.Y
		} else if csh.MinSize.Y <= renderSize.Y {
			cbounds.Max.Y = renderSize.Y
		} else {
			cbounds.Max.Y = csh.MinSize.Y
		}
		cbounds.Max.X = left + ratioX*csh.PreferredSize.X

		f.PlaceBlock(child, cbounds)

		//fmt.Println("flow", cbounds.Width(), cbounds.Height())
		left = cbounds.Max.X
	}
	f.Invalidate()
	// fmt.Println()
}

// dispense events to children, as appropriate
func (f *Flow) HandleEvents() {
	for {
		select {
		case e := <-f.UserEvents:
			switch e := e.(type) {
			default:
				f.Foundation.HandleEvent(e)
			}
		case e := <-f.BlockInvalidations:
			f.DoBlockInvalidation(e)

		case bsh := <-f.BlockSizeHints:

			if !f.Children[bsh.Block] {
				// Do I know you?
				break
			}

			if osh, ok := f.ChildrenHints[bsh.Block]; ok {
				f.sizeHint.MinSize.X -= osh.MinSize.X
				f.sizeHint.MinSize.Y -= osh.MinSize.Y
				f.sizeHint.PreferredSize.X -= osh.PreferredSize.X
				f.sizeHint.PreferredSize.Y -= osh.PreferredSize.Y
				f.sizeHint.MaxSize.X -= osh.MaxSize.X
				f.sizeHint.MaxSize.Y -= osh.MaxSize.Y
			}
			f.ChildrenHints[bsh.Block] = bsh.SizeHint
			f.sizeHint.MinSize.X += bsh.SizeHint.MinSize.X
			f.sizeHint.MinSize.Y += bsh.SizeHint.MinSize.Y
			f.sizeHint.PreferredSize.X += bsh.SizeHint.PreferredSize.X
			f.sizeHint.PreferredSize.Y += bsh.SizeHint.PreferredSize.Y
			f.sizeHint.MaxSize.X += bsh.SizeHint.MaxSize.X
			f.sizeHint.MaxSize.Y += bsh.SizeHint.MaxSize.Y

			f.SizeHints.Stack(f.sizeHint)

			f.reflow()

		case e := <-f.ResizeEvents:
			f.Size = e.Size
			f.reflow()

		case b := <-f.Add:
			f.childIndices[b] = f.count
			f.count++

			f.reflow()
		case b := <-f.Remove:
			i, ok := f.childIndices[b]
			if !ok {
				break
			}

			// decrement all following blocks
			for ob, j := range f.childIndices {
				if j > i {
					f.childIndices[ob] = j - 1
				}
			}
			delete(f.childIndices, b)
			f.count--

			delete(f.ChildrenHints, b)

			f.RemoveBlock(b)

			f.reflow()
		}
	}
}
