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

package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.wde"
	"image"
	"image/draw"
)

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
	ChildrenHints       map[*Block]SizeHint

	BlockSizeHints chan BlockSizeHint

	BlockInvalidations chan BlockInvalidation

	DragOriginBlocks map[wde.Button][]*Block

	// this block currently has keyboard priority
	KeyFocus *Block
}

func (f *Foundation) Initialize() {
	f.Block.Initialize()
	f.DrawOp = draw.Over
	f.BlockSizeHints = make(chan BlockSizeHint, 1)
	f.Children = map[*Block]bool{}
	f.ChildrenBounds = map[*Block]geom.Rect{}
	f.ChildrenHints = map[*Block]SizeHint{}
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
	delete(f.ChildrenHints, b)
	b.Parent = nil
}

func (f *Foundation) AddBlock(b *Block) {
	if b.Parent == f {
		return
	}

	// Report(f.ID, "adding", b.ID)
	if b.Parent != nil {
		// TODO: communication here
		b.Parent.RemoveBlock(b)
	}

	f.Children[b] = true
	b.Parent = f
	// Report("invalidation link", b.ID, "->", f.ID)
	b.Invalidations = make(InvalidationChan, 1)
	go func(b *Block, blockInvalidator chan Invalidation) {
		for inv := range blockInvalidator {
			f.BlockInvalidations <- BlockInvalidation{
				Invalidation: inv,
				Block:        b,
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

func (f *Foundation) Draw(buffer draw.Image, invalidRects RectSet) {
	gc := draw2d.NewGraphicContext(buffer)
	f.DoPaint(gc)
	for child, bounds := range f.ChildrenBounds {
		r := RectangleForRect(bounds)

		// only redraw those that have been invalidated or are
		// otherwise unable to draw themselves
		if child.buffer == nil || invalidRects.Intersects(bounds) {
			or := image.Rectangle{
				Max: image.Point{int(child.Size.X), int(child.Size.Y)},
			}
			if child.buffer == nil || child.buffer.Bounds() != or {
				// srgba := buffer.(*image.RGBA).SubImage(r).(*image.RGBA)
				// srgba.Rect.Max.X -= srgba.Rect.Min.X
				// srgba.Rect.Max.Y -= srgba.Rect.Min.Y
				// srgba.Rect.Min = image.Point{}
				// child.buffer = srgba
				child.buffer = image.NewRGBA(or)
			} else {
				ZeroRGBA(child.buffer.(*image.RGBA))
			}

			subInv := invalidRects.Intersection(bounds)
			subInv = subInv.Translate(bounds.Min.Times(-1))

			child.Drawer.Draw(child.buffer, subInv)
		}

		draw.Draw(buffer, r, child.buffer, image.Point{0, 0}, draw.Over)
	}
}

func (f *Foundation) DoBlockInvalidation(e BlockInvalidation) {
	cbounds, ok := f.ChildrenBounds[e.Block]
	if !ok {
		return
	}
	for _, invBounds := range e.Bounds {
		invBounds.Translate(cbounds.Min)
		f.Invalidate(invBounds)

	}
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
	case MouseDraggedEvent:
		f.DoMouseDraggedEvent(e)
	case MouseMovedEvent:
		f.DoMouseMovedEvent(e)
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
		// Report(f.ID, "mouse origin", b.ID)
		ce := e

		ce.Loc = e.Loc.Minus(bbs.Min)
		b.UserEventsIn.SendOrDrop(ce)
	})
}

func (f *Foundation) DoMouseMovedEvent(e MouseMovedEvent) {
	fromSet := map[*Block]bool{}
	f.InvokeOnBlocksUnder(e.From, func(b *Block) {
		fromSet[b] = true
	})
	f.InvokeOnBlocksUnder(e.Loc, func(b *Block) {
		bbs := f.ChildrenBounds[b]
		if !fromSet[b] {
			ee := MouseEnteredEvent{
				Event:        e.Event,
				MouseLocator: e.MouseLocator,
				From:         e.From,
			}
			ee.Loc = ee.Loc.Minus(bbs.Min)
			ee.From = ee.From.Minus(bbs.Min)
			b.UserEventsIn.SendOrDrop(ee)
		} else {
			delete(fromSet, b)
		}
		ce := e
		ce.Loc = e.Loc.Minus(bbs.Min)
		ce.From = e.From.Minus(bbs.Min)
		b.UserEventsIn.SendOrDrop(ce)
	})
	for fromBlock := range fromSet {
		bbs := f.ChildrenBounds[fromBlock]
		ee := MouseExitedEvent{
			Event:        e.Event,
			MouseLocator: e.MouseLocator,
			From:         e.From,
		}
		ee.Loc = ee.Loc.Minus(bbs.Min)
		ee.From = ee.From.Minus(bbs.Min)
		fromBlock.UserEventsIn.SendOrDrop(ee)
	}
}

func (f *Foundation) DoMouseUpEvent(e MouseUpEvent) {
	touched := map[*Block]bool{}
	f.InvokeOnBlocksUnder(e.Loc, func(b *Block) {
		touched[b] = true
		bbs := f.ChildrenBounds[b]
		if b != nil {
			be := e
			be.Loc = be.Loc.Minus(bbs.Min)
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
			oe.Loc = oe.Loc.Minus(obbs.Min)
			origin.UserEventsIn.SendOrDrop(oe)
		}
	}
	delete(f.DragOriginBlocks, e.Which)
}

func (f *Foundation) DoMouseDraggedEvent(e MouseDraggedEvent) {
	fromSet := map[*Block]bool{}
	f.InvokeOnBlocksUnder(e.From, func(b *Block) {
		fromSet[b] = true
	})
	// Report(f.ID, "mde")
	touched := map[*Block]bool{}
	f.InvokeOnBlocksUnder(e.Loc, func(b *Block) {
		touched[b] = true
		bbs := f.ChildrenBounds[b]
		if !fromSet[b] {
			ee := MouseEnteredEvent{
				Event:        e.Event,
				MouseLocator: e.MouseLocator,
				From:         e.From,
			}
			ee.Loc = ee.Loc.Minus(bbs.Min)
			ee.From = ee.From.Minus(bbs.Min)
			b.UserEventsIn.SendOrDrop(ee)
		} else {
			delete(fromSet, b)
		}
		if b != nil {
			be := e
			be.Loc = be.Loc.Minus(bbs.Min)
			be.From = be.From.Minus(bbs.Min)
			// Report(f.ID, "forward", b.ID)
			b.UserEventsIn.SendOrDrop(be)
		}
	})
	for fromBlock := range fromSet {
		bbs := f.ChildrenBounds[fromBlock]
		ee := MouseExitedEvent{
			Event:        e.Event,
			MouseLocator: e.MouseLocator,
			From:         e.From,
		}
		ee.Loc = ee.Loc.Minus(bbs.Min)
		ee.From = ee.From.Minus(bbs.Min)
		fromBlock.UserEventsIn.SendOrDrop(ee)
	}
	if origins, ok := f.DragOriginBlocks[e.Which]; ok {
		for _, origin := range origins {
			if touched[origin] {
				// Report(f.ID, "skip", origin.ID)
				continue
			}
			// Report(f.ID, "origin forward", origin.ID)
			oe := e
			obbs := f.ChildrenBounds[origin]
			oe.Loc = oe.Loc.Minus(obbs.Min)
			oe.From = oe.From.Minus(obbs.Min)
			origin.UserEventsIn.SendOrDrop(oe)
		}
	}
}

func (f *Foundation) DoCloseEvent(e CloseEvent) {
	for b := range f.Children {
		b.UserEventsIn.SendOrDrop(e)
	}
}
