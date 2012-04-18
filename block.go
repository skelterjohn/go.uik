package uik

import (
	"image"
	"image/draw"
	"code.google.com/p/draw2d/draw2d"
)

// The Block type is a basic unit that can receive events and draw itself.
//
// This struct essentially defines an interface, except a synchronous interface
// based on channels rather than an asynchronous interface based on method
// calls.
type Block struct {
	Parent *Foundation

	ListenedChannels map[interface{}]bool

	CloseEvents     chan CloseEvent
	MouseDownEvents chan MouseDownEvent
	MouseUpEvents   chan MouseUpEvent
	RedrawIn        chan RedrawEvent
	RedrawOut       <-chan RedrawEvent

	allEventsIn     chan<- interface{}
	allEventsOut    <-chan interface{}


	Paint func(gc draw2d.GraphicContext)
	Buffer draw.Image
	Compositor chan CompositeRequest

	// minimum point relative to the block's parent: only the parent should set this
	Min Coord
	// size of block
	Size Coord
}

func (b *Block) Initialize() {
	b.ListenedChannels = make(map[interface{}]bool)
	b.CloseEvents = make(chan CloseEvent)
	b.MouseDownEvents = make(chan MouseDownEvent)
	b.MouseUpEvents = make(chan MouseUpEvent)
	b.RedrawIn = make(chan RedrawEvent)
	b.RedrawOut = StackRedrawEvents(b.RedrawIn)
	b.allEventsIn, b.allEventsOut = QueuePipe()
	go b.handleSplitEvents()
}

func (b *Block) PrepareBuffer() (gc draw2d.GraphicContext) {
	min := image.Point{int(b.Min.X-1), int(b.Min.Y-1)}
	max := image.Point{int(b.Min.X+b.Size.X+1), int(b.Min.Y+b.Size.Y+1)}
	if b.Buffer == nil || b.Buffer.Bounds().Min != min || b.Buffer.Bounds().Max != max {
		b.Buffer = image.NewRGBA(image.Rectangle {
			Min: min,
			Max: max,
		})
	}
	gc = draw2d.NewGraphicContext(b.Buffer)
	return
}

func (b *Block) doPaint(gc draw2d.GraphicContext) {
	if b.Paint != nil {
		b.Paint(gc)
	}
}

func (b *Block) handleSplitEvents() {
	for e := range b.allEventsOut {
		switch e := e.(type) {
		case MouseDownEvent:
			if b.ListenedChannels[b.MouseDownEvents] {
				b.MouseDownEvents <- e
			}
		case MouseUpEvent:
			if b.ListenedChannels[b.MouseUpEvents] {
				b.MouseUpEvents <- e
			}
		case CloseEvent:
			if b.ListenedChannels[b.CloseEvents] {
				b.CloseEvents <- e
			}
		}
	}
}

func (b *Block) BoundsInParent() (bounds Bounds) {
	bounds.Min = b.Min
	bounds.Max = b.Min
	bounds.Max.X += b.Size.X
	bounds.Max.Y += b.Size.Y
	return
}
