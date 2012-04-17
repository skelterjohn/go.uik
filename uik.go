package uik

import (
	"image"
	"image/draw"
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/go.wde"
	"github.com/skelterjohn/go.wde/xgb"
)

var WindowGenerator func(parent wde.Window, width, height int) (window wde.Window, err error)

func init() {
	WindowGenerator = func(parent wde.Window, width, height int) (window wde.Window, err error) {
		window, err = xgb.NewWindow(width, height)
		return
	}
}

type DrawRequest struct {
	Dirty Bounds
}

// The Block type is a basic unit that can receive events and draw itself.
type Block struct {
	Parent *Foundation

	ListenedChannels map[interface{}]bool

	CloseEvents     chan wde.CloseEvent
	MouseDownEvents chan MouseDownEvent
	MouseUpEvents   chan MouseUpEvent

	allEventsIn     chan<- interface{}
	allEventsOut    <-chan interface{}

	Draw            chan DrawRequest
	Redraw          chan Bounds

	Paint func(gc draw2d.GraphicContext)
	Buffer draw.Image
	ParentDrawBuffer chan image.Image

	// minimum point relative to the block's parent: only the parent should set this
	Min Coord
	// size of block
	Size Coord
}

func (b *Block) PrepareBuffer() (gc draw2d.GraphicContext) {
	min := image.Point{int(b.Min.X), int(b.Min.Y)}
	max := image.Point{int(b.Min.X+b.Size.X), int(b.Min.Y+b.Size.Y)}
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
		case wde.CloseEvent:
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

func (b *Block) MakeChannels() {
	b.ListenedChannels = make(map[interface{}]bool)
	b.CloseEvents = make(chan wde.CloseEvent)
	b.MouseDownEvents = make(chan MouseDownEvent)
	b.MouseUpEvents = make(chan MouseUpEvent)
	b.Draw = make(chan DrawRequest)
	b.Redraw = make(chan Bounds)
	b.allEventsIn, b.allEventsOut = QueuePipe()
	go b.handleSplitEvents()
}

// The foundation type is for channeling events to children, and passing along
// draw calls.
type Foundation struct {
	Block
	Children    []*Block
	DrawBuffers chan image.Image

	// this block currently has keyboard priority
	KeyboardBlock    *Block
}

func (f *Foundation) MakeChannels() {
	f.Block.MakeChannels()
	f.DrawBuffers = make(chan image.Image)
}

func (f *Foundation) AddBlock(b *Block) {
	// TODO: place the block somewhere clever
	// TODO: resize the block in a clever way
	f.Children = append(f.Children, b)
	b.Parent = f
	b.ParentDrawBuffer = f.DrawBuffers
}

func (f *Foundation) BlockForCoord(p Coord) (b *Block) {
	// quad-tree one day?
	for _, bl := range f.Children {
		if bl.BoundsInParent().Contains(p) {
			b = bl
			return
		}
	}
	return
}

func (f *Foundation) handleRedraw() {
	for dirtyBounds := range f.Redraw {
		dirtyBounds.Min.X -= f.Min.X
		dirtyBounds.Min.Y -= f.Min.Y
		f.Parent.Redraw <- dirtyBounds
	}
}

// dispense events to children, as appropriate
func (f *Foundation) handleEvents() {
	f.ListenedChannels[f.CloseEvents] = true
	f.ListenedChannels[f.MouseDownEvents] = true
	f.ListenedChannels[f.MouseUpEvents] = true

	var dragOriginBlocks = map[wde.Button]*Block{}
	// drag and up events for the same button get sent to the origin as well

	for {
		select {
		case e := <-f.CloseEvents:
			for _, b := range f.Children {
				b.allEventsIn <- e
			}
		case e := <-f.MouseDownEvents:
			b := f.BlockForCoord(e.Loc)
			if b == nil {
				break
			}
			dragOriginBlocks[e.Which] = b
			e.Loc.X -= b.Min.X
			e.Loc.Y -= b.Min.Y
			b.allEventsIn <- e
		case e := <-f.MouseUpEvents:
			b := f.BlockForCoord(e.Loc)
			if b != nil {
				be := e
				be.Loc.X -= b.Min.X
				be.Loc.Y -= b.Min.Y
				b.allEventsIn <- be
			}
			if origin, ok := dragOriginBlocks[e.Which]; ok && origin != b {
				oe := e
				oe.Loc.X -= origin.Min.X
				oe.Loc.Y -= origin.Min.Y
				origin.allEventsIn <- oe
			}

		case dr := <-f.Draw:
			bgc := f.PrepareBuffer()
			if f.Paint != nil {
				f.Paint(bgc)
			}
			for _, child := range f.Children {
				translatedDirty := dr.Dirty
				translatedDirty.Min.X -= child.Min.X
				translatedDirty.Min.Y -= child.Min.Y

				cdr := DrawRequest{
					Dirty: translatedDirty,
				}
				child.Draw <- cdr

			}
		case buffer := <-f.DrawBuffers:
			bgc := f.PrepareBuffer()
			bgc.DrawImage(buffer)
			if f.ParentDrawBuffer != nil {
				f.ParentDrawBuffer <- f.Buffer
			}

		}
	}
}

// A foundation that wraps a wde.Window
type WindowFoundation struct {
	Foundation
	W wde.Window
	Done <-chan bool
}

func NewWindow(parent wde.Window, width, height int) (wf *WindowFoundation, err error) {
	wf = new(WindowFoundation)

	wf.W, err = WindowGenerator(parent, width, height)
	if err != nil {
		return
	}
	wf.MakeChannels()

	wf.Size = Coord{float64(width), float64(height)}
	wf.Paint = func(gc draw2d.GraphicContext) {
		gc.Clear()
	}

	go wf.handleWindowEvents()
	go wf.handleWindowDrawing()
	go wf.handleEvents()

	return
}

func (wf *WindowFoundation) Show() {
	wf.W.Show()
	wf.Redraw <- wf.BoundsInParent()
}

// wraps mouse events with float64 coordinates
func (wf *WindowFoundation) handleWindowEvents() {
	done := make(chan bool)
	wf.Done = done
	for e := range wf.W.EventChan() {
		switch e := e.(type) {
		case wde.CloseEvent:
			wf.CloseEvents <- e
			wf.W.Close()
			done <- true
		case wde.MouseDownEvent:
			wf.MouseDownEvents <- MouseDownEvent{
				MouseDownEvent: e,
				MouseLocator: MouseLocator {
					Loc: Coord{float64(e.Where.X), float64(e.Where.Y)},
				},
			}
		case wde.MouseUpEvent:
			wf.MouseUpEvents <- MouseUpEvent{
				MouseUpEvent: e,
				MouseLocator: MouseLocator {
					Loc: Coord{float64(e.Where.X), float64(e.Where.Y)},
				},
			}

		}
	}
}

func (wf *WindowFoundation) handleWindowDrawing() {
	// TODO: collect a dirty region (possibly disjoint), and draw in one go?
	wf.ParentDrawBuffer = make(chan image.Image)

	for {
		select {
		case dirtyBounds := <-wf.Redraw:
			gc := draw2d.NewGraphicContext(wf.W.Screen())
			gc.Clear()
			gc.BeginPath()
			// TODO: pass dirtyBounds too, to avoid redrawing out of reach components
			_ = dirtyBounds
			wf.doPaint(gc)

			dr := DrawRequest{
				Dirty: dirtyBounds,
			}
			wf.Draw <-dr

			wf.W.FlushImage()
		case buffer := <- wf.ParentDrawBuffer:
			draw.Draw(wf.W.Screen(), buffer.Bounds(), buffer, image.Point{0, 0}, draw.Src)

			wf.W.FlushImage()
		}
	}
}
