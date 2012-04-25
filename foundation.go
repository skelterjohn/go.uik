package uik

import (
	"code.google.com/p/draw2d/draw2d"
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

type BlockInvalidation struct {
	Invalidation
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

	BlockInvalidations chan BlockInvalidation

	DragOriginBlocks map[wde.Button][]*Block

	// this block currently has keyboard priority
	KeyFocus *Block
}

func (f *Foundation) Initialize() {
	f.Block.Initialize()
	f.DrawOp = draw.Over
	f.CompositeBlockRequests = make(chan CompositeBlockRequest, 1)
	f.BlockSizeHints = make(chan BlockSizeHint, 1)
	f.Children = map[*Block]bool{}
	f.ChildrenBounds = map[*Block]geom.Rect{}
	f.ChildrenLastBuffers = map[*Block]image.Image{}
	f.BlockInvalidations = make(chan BlockInvalidation, 1)
	f.DragOriginBlocks = map[wde.Button][]*Block{}
	f.Drawer = f
}

func (f *Foundation) RemoveBlock(b *Block) {
	if b.Parent != f {
		// TODO: log
		return
	}
	delete(f.Children, b)
	delete(f.ChildrenBounds, b)
	delete(f.ChildrenLastBuffers, b)
	b.Parent = nil
}

func (f *Foundation) AddBlock(b *Block) {
	// Report(f.ID, "adding", b.ID)
	if b.Parent == nil {
		f.Children[b] = true
	} else if b.Parent != f {
		// TODO: communication here
		b.Parent.RemoveBlock(b)
	}

	b.Parent = f
	// Report("invalidation link", b.ID, "->", f.ID)
	b.Invalidations = make(InvalidationChan, 1)
	go func(b *Block, blockInvalidator chan Invalidation) {
		for inv := range blockInvalidator {
			f.BlockInvalidations <- BlockInvalidation{
				Invalidation: 	inv,
				Block:            b,
			}
		}
	}(b, b.Invalidations)

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
	// Report(f.ID, "placing", b.ID)
	f.AddBlock(b)
	f.ChildrenBounds[b] = bounds
	b.UserEventsIn.SendOrDrop(ResizeEvent{
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

func (f *Foundation) Draw(buffer draw.Image) {
	// Report(f.ID, "Foundation.Draw()", buffer.Bounds())
	gc := draw2d.NewGraphicContext(buffer)
	f.DoPaint(gc)
	for child, bounds := range f.ChildrenBounds {
		r := RectangleForRect(bounds)
		var subbuffer draw.Image
		// subbuffer = buffer.(*image.RGBA).SubImage(r).(draw.Image)
		or := image.Rectangle{
			Max: image.Point{int(child.Size.X), int(child.Size.Y)},
		}
		subbuffer = image.NewRGBA(or)
		child.Drawer.Draw(subbuffer)
		draw.Draw(buffer, r, subbuffer, image.Point{0, 0}, draw.Over)
	}
}

func (f *Foundation) DoBlockInvalidation(e BlockInvalidation) {
	// Report(f.ID, "invalidation from", e.Block.ID)
	f.Invalidate()
}

// internal events

func (f *Foundation) DoResizeEvent(e ResizeEvent) {
	if e.Size == f.Size {
		return
	}
	f.Size = e.Size
	f.Invalidate()
}

func (f *Foundation) KeyFocusRequest(e KeyFocusRequest) {
	if e.Block == nil {
		return
	}
	if !f.Children[e.Block] {
		return
	}
	if e.Block != f.KeyFocus && f.KeyFocus != nil {
		f.KeyFocus.UserEventsIn.SendOrDrop(KeyFocusEvent{
			Focus: false,
		})
	}
	f.KeyFocus = e.Block
	if f.HasKeyFocus {
		if f.KeyFocus != nil {
			f.KeyFocus.UserEventsIn.SendOrDrop(KeyFocusEvent{
				Focus: true,
			})
		}
	} else {
		if f.Parent != nil {
			f.Parent.UserEventsIn.SendOrDrop(KeyFocusRequest{
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
		case e := <-f.UserEvents:
			f.HandleEvent(e)
		case e := <-f.BlockInvalidations:
			f.DoBlockInvalidation(e)
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
		f.KeyFocus.UserEventsIn.SendOrDrop(e)
	}
}

func (f *Foundation) DoKeyEvent(e interface{}) {
	if f.KeyFocus == nil {
		return
	}
	f.KeyFocus.UserEventsIn.SendOrDrop(e)
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
		b.UserEventsIn.SendOrDrop(e)
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
			b.UserEventsIn.SendOrDrop(be)
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
			origin.UserEventsIn.SendOrDrop(oe)
		}
	}
	delete(f.DragOriginBlocks, e.Which)
}

func (f *Foundation) DoCloseEvent(e CloseEvent) {
	for b := range f.Children {
		b.UserEventsIn.SendOrDrop(e)
	}
}
