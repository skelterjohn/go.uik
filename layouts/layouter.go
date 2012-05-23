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
)

type Layout map[*uik.Block]geom.Rect

// Implement this interface with your own layout code. All methods
// listed in this interface will be called from the same goroutine
// in the Layouter, so no mutual thread-safety is needed. However,
// if your implementing type has more methods (perhaps to add
// a new block, or to set configuration data), those methods will
// need to implement some sort of thread-safety when manipulating
// the data accessed by the LayoutEngine methods.
type LayoutEngine interface {
	// The layouter using this engine. The engine should use this
	// layouter's .AddBlock(), .RemoveBlock(), .Invalidate()
	// methods, etc.
	SetLayouter(layouter *Layouter)
	// Set the hint associated with this block.
	SetHint(block *uik.Block, hint uik.SizeHint)
	// Return the hint that the Layouter will report.
	GetHint() uik.SizeHint
	// Get the bounds for each of the blocks in this layout.
	GetLayout(size geom.Coord) Layout
	// Items sent on the config channel will appear as a parameter
	// to this function. It is suffixed by "Unsafe" to deter
	// sites other than the Layouter from calling it, since it
	// would not be safe for them.
	ConfigUnsafe(cfg interface{})
}

type Layouter struct {
	uik.Foundation

	engine LayoutEngine

	config chan interface{}
}

func NewLayouter(engine LayoutEngine) (l *Layouter) {
	l = new(Layouter)

	l.Initialize()

	l.engine = engine
	l.engine.SetLayouter(l)

	l.SetSizeHint(l.engine.GetHint())

	go l.HandleEvents()

	return
}

func (l *Layouter) Initialize() {
	l.Foundation.Initialize()
	l.config = make(chan interface{}, 1)
	l.Paint = nil
}

func (l *Layouter) placeBlocks() {
	layout := l.engine.GetLayout(l.Size)
	for block, bounds := range layout {
		l.PlaceBlock(block, bounds)
	}
}

// dispense events to children, as appropriate
func (l *Layouter) HandleEvents() {
	for {
		select {
		case e := <-l.UserEvents:
			l.HandleEvent(e)
		case e := <-l.BlockInvalidations:
			l.DoBlockInvalidation(e)
		case bsh := <-l.BlockSizeHints:
			l.engine.SetHint(bsh.Block, bsh.SizeHint)
			l.placeBlocks()
			l.SetSizeHint(l.engine.GetHint())
		case e := <-l.ResizeEvents:
			l.DoResizeEvent(e)
			l.placeBlocks()
		case cfg := <-l.config:
			l.engine.ConfigUnsafe(cfg)
		}
	}
}

func (l *Layouter) Config(cfg interface{}) {
	l.config <- cfg
}
