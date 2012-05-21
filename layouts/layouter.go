package layouts

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
)

type Layout map[*uik.Block]geom.Rect

type LayoutEngine interface {
	SetHint(block *uik.Block, hint uik.SizeHint)
	GetHint() uik.SizeHint
	GetLayout(size geom.Coord) Layout
}

type Layouter struct {
	uik.Foundation

	engine LayoutEngine
}

func NewLayouter(engine LayoutEngine) (l *Layouter) {
	l = new(Layouter)

	l.Initialize()

	l.engine = engine

	go l.HandleEvents()

	return
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

			if !l.Children[bsh.Block] {
				// Do I know you?
				break
			}

			l.engine.SetHint(bsh.Block, bsh.SizeHint)
			l.placeBlocks()
			l.SetSizeHint(l.engine.GetHint())
		case e := <-l.ResizeEvents:
			l.DoResizeEvent(e)
			l.placeBlocks()
		}
	}
}