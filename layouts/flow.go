package layouts

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
)

type Flow struct {
	uik.Foundation

	childSizeHints map[*uik.Block]uik.SizeHint

	size geom.Coord
	sizeHint uik.SizeHint
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
	f.AddBlock(b)

	f.ChildrenBounds[b] = geom.Rect{
		Min: geom.Coord{f.size.X, 0},
		Max: geom.Coord{f.size.X + b.Size.X, b.Size.Y},
	}
	f.size.X += b.Size.X
}

