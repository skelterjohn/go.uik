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

	DrawOp draw.Op

	Children            map[*Block]bool
	ChildrenBounds      map[*Block]geom.Rect
	ChildrenLastBuffers map[*Block]image.Image

	CompositeBlockRequests chan CompositeBlockRequest

	ChildrenHints  map[*Block]SizeHint
	BlockSizeHints chan BlockSizeHint

	DragOriginBlocks map[wde.Button][]*Block

	// this block currently has keyboard priority
	KeyFocus *Block
}

func (f *Foundation) Initialize() {
	f.Block.Initialize()
	f.DrawOp = draw.Over
	f.CompositeBlockRequests = make(chan CompositeBlockRequest)
	f.BlockSizeHints = make(chan BlockSizeHint)
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

	b.Parent = f

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

	b.placementNotifications.Stack(placementNotification{
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
	b.EventsIn.SendOrDrop(ResizeEvent{
		Size: geom.Coord{bounds.Max.X - bounds.Min.X, bounds.Max.Y - bounds.Min.Y},
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

// drawing

func (f *Foundation) CompositeBlockBuffer(b *Block, buf image.Image) (composited bool) {
	// Report(f.ID, "composite", b.ID, "using", f.DrawOp)
	bounds, ok := f.ChildrenBounds[b]
	if !ok {
		composited = false
		return
	}
	f.PrepareBuffer()
	// dstOffset := image.Point{
	// 	int(bounds.Min.X),
	// 	int(bounds.Min.Y),
	// }
	// copyRGBA(f.Buffer.(*image.RGBA), dstOffset, buf.(*image.RGBA))
	draw.Draw(f.Buffer, RectangleForRect(bounds), buf, image.Point{0, 0}, f.DrawOp)
	composited = true
	return
}

func (f *Foundation) DoCompositeBlockRequest(cbr CompositeBlockRequest) {
	// Report(f.ID, "cbr", cbr.Block.ID)
	f.ChildrenLastBuffers[cbr.Block] = cbr.Buffer
	f.Rebuffer()
}

func (f *Foundation) Rebuffer() {
	// Report(f.ID, "rebuffer")
	bgc := f.PrepareBuffer()
	bgc.Clear()
	f.DoPaint(bgc)
	for child, buf := range f.ChildrenLastBuffers {
		f.CompositeBlockBuffer(child, buf)
	}
	f.Compositor.Stack(CompositeRequest{
		Buffer: copyImage(f.Buffer),
	})
}

func (f *Foundation) DoRedraw(e RedrawEvent) {
	// Report(f.ID, "redraw")
	bgc := f.PrepareBuffer()
	f.DoPaint(bgc)
	for child := range f.Children {
		bbs, ok := f.ChildrenBounds[child]
		if !ok {
			continue
		}

		if buf, ok := f.ChildrenLastBuffers[child]; ok {
			f.CompositeBlockBuffer(child, buf)
		} else {
			translatedDirty := e.Bounds
			translatedDirty.Min.X -= bbs.Min.X
			translatedDirty.Min.Y -= bbs.Min.Y

			child.Redraw.Stack(RedrawEvent{translatedDirty})
		}
	}
	if f.Compositor != nil {
		f.Compositor <- CompositeRequest{
			Buffer: f.Buffer,
		}
	}
}

// internal events

func (f *Foundation) DoResizeEvent(e ResizeEvent) {
	if e.Size == f.Size {
		return
	}
	f.Size = e.Size
	f.Rebuffer()
}

func (f *Foundation) KeyFocusRequest(e KeyFocusRequest) {
	if e.Block == nil {
		return
	}
	if !f.Children[e.Block] {
		return
	}
	if e.Block != f.KeyFocus && f.KeyFocus != nil {
		f.KeyFocus.EventsIn.SendOrDrop(KeyFocusEvent{
			Focus: false,
		})
	}
	f.KeyFocus = e.Block
	if f.HasKeyFocus {
		if f.KeyFocus != nil {
			f.KeyFocus.EventsIn.SendOrDrop(KeyFocusEvent{
				Focus: true,
			})
		}
	} else {
		if f.Parent != nil {
			f.Parent.EventsIn.SendOrDrop(KeyFocusRequest{
				Block: &f.Block,
			})
		}
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
	case KeyFocusEvent:
		f.DoKeyFocusEvent(e)
	case KeyFocusRequest:
		f.KeyFocusRequest(e)
	case KeyDownEvent, KeyUpEvent, KeyTypedEvent:
		f.DoKeyEvent(e)
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

// input events

func (f *Foundation) DoKeyFocusEvent(e KeyFocusEvent) {
	if e.Focus == f.HasKeyFocus {
		return
	}
	f.HasKeyFocus = e.Focus
	if f.KeyFocus != nil {
		f.KeyFocus.EventsIn.SendOrDrop(e)
	}
}

func (f *Foundation) DoKeyEvent(e interface{}) {
	if f.KeyFocus == nil {
		return
	}
	f.KeyFocus.EventsIn.SendOrDrop(e)
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
		b.EventsIn.SendOrDrop(e)
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
			b.EventsIn.SendOrDrop(be)
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
			origin.EventsIn.SendOrDrop(oe)
		}
	}
	delete(f.DragOriginBlocks, e.Which)
}

func (f *Foundation) DoCloseEvent(e CloseEvent) {
	for b := range f.Children {
		b.EventsIn.SendOrDrop(e)
	}
}
