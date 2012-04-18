package uik

import (
	"image"
	"image/draw"
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
)

type EventFilter func(e interface{}) (accept, done bool)

type EventSubscription struct {
	Filter EventFilter
	Ch chan<- interface{}
}

// The Block type is a basic unit that can receive events and draw itself.
//
// This struct essentially defines an interface, except a synchronous interface
// based on channels rather than an asynchronous interface based on method
// calls.
type Block struct {
	Parent *Foundation

	ListenedChannels map[interface{}]bool

	allEventsIn     chan<- interface{}
	allEventsOut    <-chan interface{}

	subscriptions map[*EventFilter]chan<- interface{}
	Subscribe chan EventSubscription

	// the event channels

	CloseEvents     chan CloseEvent
	MouseDownEvents chan MouseDownEvent
	MouseUpEvents   chan MouseUpEvent
	ResizeEvents    chan ResizeEvent

	Redraw        chan RedrawEvent



	Paint func(gc draw2d.GraphicContext)
	Buffer draw.Image
	Compositor chan CompositeRequest

	// size of block
	Size geom.Coord
}

func (b *Block) Initialize() {
	b.Paint = ClearPaint

	b.ListenedChannels = make(map[interface{}]bool)

	b.subscriptions = map[*EventFilter]chan<-interface{}{}
	b.Subscribe = make(chan EventSubscription)

	b.allEventsIn, b.allEventsOut = QueuePipe()

	b.CloseEvents = make(chan CloseEvent)
	b.MouseDownEvents = make(chan MouseDownEvent)
	b.MouseUpEvents = make(chan MouseUpEvent)
	b.ResizeEvents = make(chan ResizeEvent)

	b.Redraw = make(chan RedrawEvent, 1)
	go b.handleSplitEvents()
}

func (b *Block) Bounds() geom.Rect {
	return geom.Rect {
		geom.Coord{0, 0},
		b.Size,
	}
}

func (b *Block) PrepareBuffer() (gc draw2d.GraphicContext) {
	min := image.Point{0, 0}
	max := image.Point{int(b.Size.X), int(b.Size.Y)}
	if b.Buffer == nil || b.Buffer.Bounds().Min != min || b.Buffer.Bounds().Max != max {
		b.Buffer = image.NewRGBA(image.Rectangle {
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

func (b *Block) PaintAndComposite() {
	bgc := b.PrepareBuffer()
	b.DoPaint(bgc)
	if b.Compositor == nil {
		return
	}
	b.Compositor <- CompositeRequest{
		Buffer: b.Buffer,
	}
}

func (b *Block) handleSplitEvents() {
	for e := range b.allEventsOut {
		// get new subscriptions
		subloop:
		for {
			select {
			case es := <- b.Subscribe:
				b.subscribeToEvents(es.Filter, es.Ch)
			default:
				break subloop
			}
		}

		for filterp, ch := range b.subscriptions {
			accept, done := (*filterp)(e)
			if accept {
				ch <- e
			}
			if done {
				delete(b.subscriptions, filterp)
			}
		}

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
		case ResizeEvent:
			if b.ListenedChannels[b.ResizeEvents] {
				b.ResizeEvents <- e
			}
		}
	}
}

func (b *Block) subscribeToEvents(filter EventFilter, ch chan<- interface{}) {
	inch := make(chan interface{})
	go RingIQ(inch, ch, 0)
	b.subscriptions[&filter] = inch
	return
}
