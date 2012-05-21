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
	"image/draw"
)

type BlockID int

var blockIDs = make(chan BlockID)

func init() {
	go func() {
		var counter BlockID
		for {
			counter++
			blockIDs <- counter
		}
	}()
}

type Drawer interface {
	Draw(buffer draw.Image, invalidRects RectSet)
}

// The Block type is a basic unit that can receive events and draw itself.
//
// This struct essentially defines an interface, except a synchronous interface
// based on channels rather than an asynchronous interface based on method
// calls.
type Block struct {
	ID BlockID

	Parent *Foundation

	UserEventsIn DropChan
	UserEvents   <-chan interface{}

	ResizeEvents ResizeChan

	Subscribe chan<- Subscription

	Drawer
	buffer draw.Image

	Paint func(gc draw2d.GraphicContext)

	Invalidations InvalidationChan
	SizeHints     SizeHintChan
	setSizeHint   SizeHintChan

	placementNotifications placementNotificationChan

	HasKeyFocus bool

	// size of block 
	Size geom.Coord
}

func (b *Block) Initialize() {
	b.ID = <-blockIDs

	b.Paint = ClearPaint
	b.Drawer = b

	b.UserEventsIn, b.UserEvents, b.Subscribe = SubscriptionQueue(20)

	b.ResizeEvents = make(ResizeChan, 1)
	b.placementNotifications = make(placementNotificationChan, 1)
	b.setSizeHint = make(SizeHintChan, 1)

	go b.handleSizeHints()
}

func (b *Block) Draw(buffer draw.Image, invalidRects RectSet) {
	// Report(b.ID, "Block.Draw()", buffer.Bounds())
	gc := draw2d.NewGraphicContext(buffer)
	b.DoPaint(gc)
}

func (b *Block) Invalidate(areas ...geom.Rect) {
	// Report(b.ID, "invalidation")
	if len(areas) == 0 {
		b.Invalidations.Stack(Invalidation{
			Bounds: []geom.Rect{b.Bounds()},
		})
	} else {
		b.Invalidations.Stack(Invalidation{
			Bounds: areas,
		})
	}
}

func (b *Block) HandleEvent(e interface{}) {
	switch e := e.(type) {
	case KeyFocusEvent:
		b.HasKeyFocus = e.Focus
	}
}

func (b *Block) SetSizeHint(sh SizeHint) {
	b.setSizeHint <- sh
}

func (b *Block) handleSizeHints() {
	sh := <-b.setSizeHint
	b.SizeHints.Stack(sh)
	for {
		select {
		case sh = <-b.setSizeHint:
		case pn := <-b.placementNotifications:
			b.Parent = pn.Foundation
			b.SizeHints = pn.SizeHints
		}
		b.SizeHints.Stack(sh)
	}
}

func (b *Block) Bounds() geom.Rect {
	return geom.Rect{
		geom.Coord{0, 0},
		b.Size,
	}
}

func (b *Block) DoPaint(gc draw2d.GraphicContext) {
	if b.Paint != nil {
		b.Paint(gc)
	}
}

func (b *Block) DoResizeEvent(e ResizeEvent) {
	if e.Size == b.Size {
		return
	}
	b.Size = e.Size
	b.Invalidate()
}
