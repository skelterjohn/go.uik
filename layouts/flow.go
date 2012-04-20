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


// dispense events to children, as appropriate
func (f *Flow) HandleEvents() {
	for {
		select {
		case e := <-f.Events:
			f.Foundation.HandleEvent(e)
		case e := <-f.Redraw:
			f.DoRedraw(e)
		case e := <-f.CompositeBlockRequests:
			f.DoCompositeBlockRequest(e)
		case bsh := <-f.BlockSizeHints:
			var bbs geom.Rect
			var ok bool
			if bbs, ok = f.ChildrenBounds[bsh.Block]; !ok {
				break
			}
			bbs.Max.X = bbs.Min.X + bsh.SizeHint.PreferredSize.X
			bbs.Max.Y = bbs.Min.Y + bsh.SizeHint.PreferredSize.Y
			f.ChildrenBounds[bsh.Block] = bbs
			bsh.Block.EventsIn <- uik.ResizeEvent {
				Size: bsh.SizeHint.PreferredSize,
			}
		}
	}
}