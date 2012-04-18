package layouts

import (
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/geom"
)


type Flow struct {
	uik.Foundation

	width float64
}

func NewFlow(size geom.Coord) (f *Flow) {
	f = new(Flow)
	f.Initialize()
	f.Size = size

	go f.HandleEvents()

	return
}

// places the block immediately to the right of the last block placed
func (f *Flow) PlaceBlock(b *uik.Block) {
	if b.Parent == nil {
		f.Children = append(f.Children, b)
		b.Parent = &f.Foundation
	} else if b.Parent != &f.Foundation {
		b.Parent.RemoveBlock(b)
		b.Parent = &f.Foundation
	}

	f.ChildrenBounds[b] = geom.Rect{
		Min: geom.Coord{f.width, 0},
		Max: geom.Coord{f.width+b.Size.X, b.Size.Y},
	}
	f.width += b.Size.X

	b.Compositor = make(chan uik.CompositeRequest)
	go func(b *uik.Block, blockCompositor chan uik.CompositeRequest) {
		for cr := range blockCompositor {
			f.CompositeBlockRequests <- uik.CompositeBlockRequest {
				CompositeRequest: cr,
				Block: b,
			}
		}
	}(b, b.Compositor)
	uik.RedrawEventChan(f.Redraw).Stack(uik.RedrawEvent{
		f.ChildrenBounds[b],
	})
}
