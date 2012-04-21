package uik

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.wde"
	"image"
	"image/draw"
)

type CompositeBlockRequest struct {
	CompositeRequest
	Block *Block
}

type BlockSizeHint struct {
	SizeHint
	Block *Block
}

// The foundation type is for channeling events to children, and passing along
// draw calls.
type Foundation struct {
	Block

	Children            map[*Block]bool
	ChildrenBounds      map[*Block]geom.Rect
	ChildrenLastBuffers map[*Block]image.Image

	CompositeBlockRequests chan CompositeBlockRequest

	ChildrenHints  map[*Block]SizeHint
	BlockSizeHints chan BlockSizeHint

	DragOriginBlocks map[wde.Button][]*Block

	// this block currently has keyboard priority
	KeyboardBlock *Block
	KeyFocusRequests KeyFocusChan
}

func (f *Foundation) Initialize() {
	f.Block.Initialize()
	f.CompositeBlockRequests = make(chan CompositeBlockRequest)
	f.BlockSizeHints = make(chan BlockSizeHint)
	f.KeyFocusRequests = make(KeyFocusChan, 1)
	f.Children = map[*Block]bool{}
	f.ChildrenBounds = map[*Block]geom.Rect{}
	f.ChildrenLastBuffers = map[*Block]image.Image{}
	f.DragOriginBlocks = map[wde.Button][]*Block{}
}

func (f *Foundation) RemoveBlock(b *Block) {
	if b.Parent != f {
		// TODO: log
		return
	}
	close(b.Compositor)
	b.Compositor = nil
	if bounds, ok := f.ChildrenBounds[b]; ok {
		RedrawEventChan(f.Redraw).Stack(RedrawEvent{
			bounds,
		})
	}
	delete(f.Children, b)
	delete(f.ChildrenBounds, b)
	delete(f.ChildrenLastBuffers, b)
	b.Parent = nil
}

func (f *Foundation) AddBlock(b *Block) {
	if b.Parent == nil {
		f.Children[b] = true
	} else if b.Parent != f {
		// TODO: communication here
		b.Parent.RemoveBlock(b)
	}

	b.Compositor = make(CompositeRequestChan, 1)
	go func(b *Block, blockCompositor chan CompositeRequest) {
		for cr := range blockCompositor {
			f.CompositeBlockRequests <- CompositeBlockRequest{
				CompositeRequest: cr,
				Block:            b,
			}
		}
	}(b, b.Compositor)

	sizeHints := make(SizeHintChan, 1)
	go func(b *Block, sizeHints chan SizeHint) {
		for sh := range sizeHints {
			f.BlockSizeHints <- BlockSizeHint{
				SizeHint: sh,
				Block:    b,
			}
		}
	}(b, sizeHints)

	b.PlacementNotifications.Stack(PlacementNotification{
		Foundation: f,
		SizeHints:  sizeHints,
	})
}

func (f *Foundation) PlaceBlock(b *Block, bounds geom.Rect) {
	f.AddBlock(b)
	f.ChildrenBounds[b] = bounds
	RedrawEventChan(f.Redraw).Stack(RedrawEvent{
		bounds,
	})
}

func (f *Foundation) BlocksForCoord(p geom.Coord) (bs []*Block) {
	// quad-tree one day?
	for bl := range f.Children {
		bbs, ok := f.ChildrenBounds[bl]
		if !ok {
			continue
		}
		if bbs.ContainsCoord(p) {
			bs = append(bs, bl)
		}
	}
	return
}

func (f *Foundation) InvokeOnBlocksUnder(p geom.Coord, foo func(*Block)) {
	// quad-tree one day?
	for bl := range f.Children {
		bbs, ok := f.ChildrenBounds[bl]
		if !ok {
			continue
		}
		if bbs.ContainsCoord(p) {
			foo(bl)
			return
		}
	}
	return

}

func (f *Foundation) CompositeBlockBuffer(b *Block, buf image.Image) (composited bool) {
	bounds, ok := f.ChildrenBounds[b]
	if !ok {
		composited = false
		return
	}
	f.PrepareBuffer()
	draw.Draw(f.Buffer, RectangleForRect(bounds), buf, image.Point{0, 0}, draw.Over)
	composited = true
	return
}

func (f *Foundation) DoCompositeBlockRequest(cbr CompositeBlockRequest) {
	b := cbr.Block
	f.ChildrenLastBuffers[b] = cbr.Buffer
	f.Rebuffer()
}

func (f *Foundation) Rebuffer() {
	bgc := f.PrepareBuffer()
	f.DoPaint(bgc)
	for child := range f.Children {
		if buf, ok := f.ChildrenLastBuffers[child]; ok {
			f.CompositeBlockBuffer(child, buf)
		}
	}
	CompositeRequestChan(f.Compositor).Stack(CompositeRequest{
		Buffer: f.Buffer,
	})
}

func (f *Foundation) DoRedraw(e RedrawEvent) {
	bgc := f.PrepareBuffer()
	f.DoPaint(bgc)
	for child := range f.Children {
		translatedDirty := e.Bounds
		bbs, ok := f.ChildrenBounds[child]
		if !ok {
			continue
		}

		translatedDirty.Min.X -= bbs.Min.X
		translatedDirty.Min.Y -= bbs.Min.Y

		RedrawEventChan(child.Redraw).Stack(RedrawEvent{translatedDirty})

		if buf, ok := f.ChildrenLastBuffers[child]; ok {
			f.CompositeBlockBuffer(child, buf)
		}
	}
	if f.Compositor != nil {
		f.Compositor <- CompositeRequest{
			Buffer: f.Buffer,
		}
	}
}

func (f *Foundation) DoMouseDownEvent(e MouseDownEvent) {
	f.InvokeOnBlocksUnder(e.Loc, func(b *Block) {
		bbs := f.ChildrenBounds[b]
		if b == nil {
			return
		}
		f.DragOriginBlocks[e.Which] = append(f.DragOriginBlocks[e.Which], b)
		e.Loc.X -= bbs.Min.X
		e.Loc.Y -= bbs.Min.Y
		b.EventsIn <- e
	})
}

func (f *Foundation) DoMouseUpEvent(e MouseUpEvent) {
	touched := map[*Block]bool{}
	f.InvokeOnBlocksUnder(e.Loc, func(b *Block) {
		touched[b] = true
		bbs := f.ChildrenBounds[b]
		if b != nil {
			be := e
			be.Loc.X -= bbs.Min.X
			be.Loc.Y -= bbs.Min.Y
			b.EventsIn <- be
		}
	})
	if origins, ok := f.DragOriginBlocks[e.Which]; ok {
		for _, origin := range origins {
			if touched[origin] {
				continue
			}
			oe := e
			obbs := f.ChildrenBounds[origin]
			oe.Loc.X -= obbs.Min.X
			oe.Loc.Y -= obbs.Min.Y
			origin.EventsIn <- oe
		}
	}
	delete(f.DragOriginBlocks, e.Which)
}

func (f *Foundation) DoResizeEvent(e ResizeEvent) {
	if e.Size == f.Size {
		return
	}
	f.Size = e.Size
	f.Rebuffer()
}

func (f *Foundation) DoCloseEvent(e CloseEvent) {
	for b := range f.Children {
		b.EventsIn <- e
	}
}

func (f *Foundation) HandleEvent(e interface{}) {
	switch e := e.(type) {
	case CloseEvent:
		f.DoCloseEvent(e)
	case MouseDownEvent:
		f.DoMouseDownEvent(e)
	case MouseUpEvent:
		f.DoMouseUpEvent(e)
	case ResizeEvent:
		f.DoResizeEvent(e)
	default:
		f.Block.HandleEvent(e)
	}
}

// dispense events to children, as appropriate
func (f *Foundation) HandleEvents() {
	for {
		select {
		case e := <-f.Events:
			f.HandleEvent(e)
		case e := <-f.Redraw:
			f.DoRedraw(e)
		case e := <-f.CompositeBlockRequests:
			f.DoCompositeBlockRequest(e)
		}
	}
}
