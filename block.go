package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"image"
	"image/draw"
)

// The Block type is a basic unit that can receive events and draw itself.
//
// This struct essentially defines an interface, except a synchronous interface
// based on channels rather than an asynchronous interface based on method
// calls.
type Block struct {
	Parent *Foundation

	EventsIn chan<- interface{}
	Events   <-chan interface{}

	Subscribe chan<- Subscription

	Redraw RedrawEventChan

	Paint  func(gc draw2d.GraphicContext)
	Buffer draw.Image

	Compositor  CompositeRequestChan
	SizeHints   SizeHintChan
	setSizeHint SizeHintChan

	placementNotifications placementNotificationChan

	HasKeyFocus bool

	// size of block
	Size geom.Coord
}

func (b *Block) Initialize() {
	b.Paint = ClearPaint

	b.EventsIn, b.Events, b.Subscribe = SubscriptionQueue(0)

	b.Redraw = make(RedrawEventChan, 1)

	b.placementNotifications = make(placementNotificationChan, 1)
	b.setSizeHint = make(SizeHintChan, 1)

	go b.handleSizeHints()
}

func (b *Block) HandleEvent(e interface{}) {
	switch e := e.(type) {
	case ResizeEvent:
		b.Size = e.Size
		b.PaintAndComposite()
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

func (b *Block) PrepareBuffer() (gc draw2d.GraphicContext) {
	min := image.Point{0, 0}
	max := image.Point{int(b.Size.X), int(b.Size.Y)}
	if b.Buffer == nil || b.Buffer.Bounds().Min != min || b.Buffer.Bounds().Max != max {
		b.Buffer = image.NewRGBA(image.Rectangle{
			Min: min,
			Max: max,
		})
	}
	gc = draw2d.NewGraphicContext(b.Buffer)
	return
}

func (b *Block) DoPaint(gc draw2d.GraphicContext) {
	if b.Paint != nil {
		b.Paint(gc)
	}
}

func copyImage(src image.Image) (dst image.Image) {
	di := image.NewRGBA(src.Bounds())
	dst = di

	draw.Draw(di, dst.Bounds(), src, image.Point{0, 0}, draw.Over)

	return
}

func (b *Block) PaintAndComposite() {
	bgc := b.PrepareBuffer()
	b.DoPaint(bgc)
	if b.Compositor == nil {
		return
	}
	CompositeRequestChan(b.Compositor).Stack(CompositeRequest{
		Buffer: copyImage(b.Buffer),
	})
}
