package layouts

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"math"
)

type Flow struct {
	uik.Foundation

	childSizeHints map[*uik.Block]uik.SizeHint
	childIndices   map[*uik.Block]int

	count int

	size     geom.Coord
	sizeHint uik.SizeHint

	Add    chan *uik.Block
	Remove chan *uik.Block
}

func NewFlow(size geom.Coord) (f *Flow) {
	f = new(Flow)
	f.Size = size
	f.Initialize()

	go f.HandleEvents()

	return
}

func (f *Flow) Initialize() {
	f.Foundation.Initialize()

	f.Add = make(chan *uik.Block, 10)
	f.Remove = make(chan *uik.Block, 10)
	f.childSizeHints = map[*uik.Block]uik.SizeHint{}
	f.childIndices = map[*uik.Block]int{}
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
		csh, ok := f.childSizeHints[child]
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
		f.ChildrenBounds[child] = cbounds

		child.EventsIn <- uik.ResizeEvent{
			Size: geom.Coord{cbounds.Max.X-cbounds.Min.X, cbounds.Max.Y},
		}
		//fmt.Println("flow", cbounds.Width(), cbounds.Height())
		left = cbounds.Max.X
	}
	// fmt.Println()
}

// dispense events to children, as appropriate
func (f *Flow) HandleEvents() {
	for {
		select {
		case e := <-f.Events:
			switch e := e.(type) {
			case uik.ResizeEvent:
				f.Size = e.Size
				f.reflow()
			default:
				f.Foundation.HandleEvent(e)
			}
		case e := <-f.Redraw:
			f.DoRedraw(e)
		case e := <-f.CompositeBlockRequests:
			f.DoCompositeBlockRequest(e)
		case bsh := <-f.BlockSizeHints:

			if !f.Children[bsh.Block] {
				// Do I know you?
				break
			}
			
			if osh, ok := f.childSizeHints[bsh.Block]; ok {
				f.sizeHint.MinSize.X -= osh.MinSize.X
				f.sizeHint.MinSize.Y -= osh.MinSize.Y
				f.sizeHint.PreferredSize.X -= osh.PreferredSize.X
				f.sizeHint.PreferredSize.Y -= osh.PreferredSize.Y
				f.sizeHint.MaxSize.X -= osh.MaxSize.X
				f.sizeHint.MaxSize.Y -= osh.MaxSize.Y
			}
			f.childSizeHints[bsh.Block] = bsh.SizeHint
			f.sizeHint.MinSize.X += bsh.SizeHint.MinSize.X
			f.sizeHint.MinSize.Y += bsh.SizeHint.MinSize.Y
			f.sizeHint.PreferredSize.X += bsh.SizeHint.PreferredSize.X
			f.sizeHint.PreferredSize.Y += bsh.SizeHint.PreferredSize.Y
			f.sizeHint.MaxSize.X += bsh.SizeHint.MaxSize.X
			f.sizeHint.MaxSize.Y += bsh.SizeHint.MaxSize.Y

			f.SizeHints.Stack(f.sizeHint)

			f.reflow()

		case b := <-f.Add:
			f.PlaceBlock(b)
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

			delete(f.childSizeHints, b)

			f.RemoveBlock(b)

			f.reflow()
		}
	}
}
