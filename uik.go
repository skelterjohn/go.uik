package uik

import (
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
	GC draw2d.GraphicContext
	Dirty Bounds
	Done chan bool
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

	// minimum point relative to the block's parent: only the parent should set this
	Min Coord
	// size of block
	Size Coord
}

func (b *Block) doPaint(gc draw2d.GraphicContext) {
	if b.Paint != nil {
		b.Paint(gc)
	}
}

func (b *Block) draw() {
	for dr := range b.Draw {
		b.doPaint(dr.GC)
		dr.Done <- true
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

	// this block currently has keyboard priority
	KeyboardBlock    *Block
}

func (f *Foundation) AddBlock(b *Block) {
	// TODO: place the block somewhere clever
	// TODO: resize the block in a clever way
	f.Children = append(f.Children, b)
	b.Parent = f
}

// func (f *Foundation) handleDrawing() {
// 	for {
// 		select {
// 		case dr := <-f.Draw:
// 			if f.Paint != nil {
// 				f.Paint(dr.GC)
// 			}
// 			for _, child := range f.Children {
// 				dr.GC.Save()

// 				translatedDirty := dr.Dirty
// 				translatedDirty.Min.X -= child.Min.X
// 				translatedDirty.Min.Y -= child.Min.Y

// 				// TODO: clip to child.BoundsInParent()?
// 				dr.GC.Translate(child.Min.X, child.Min.Y)
// 				cdr := DrawRequest{
// 					GC: dr.GC,
// 					Dirty: translatedDirty,
// 					Done: make(chan bool),
// 				}
// 				child.Draw <- dr
// 				<- cdr.Done

// 				dr.GC.Restore()
// 			}
// 			dr.Done<- true
// 		case dirtyBounds := <-f.Redraw:
// 			dirtyBounds.Min.X -= f.Min.X
// 			dirtyBounds.Min.Y -= f.Min.Y
// 			f.Parent.Redraw <- dirtyBounds
// 		}
// 	}
// }

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
			if origin, ok := dragOriginBlocks[e.Which]; ok {
				oe := e
				oe.Loc.X -= origin.Min.X
				oe.Loc.Y -= origin.Min.Y
				origin.allEventsIn <- oe
			}

		case dr := <-f.Draw:
			if f.Paint != nil {
				f.Paint(dr.GC)
			}
			for _, child := range f.Children {
				dr.GC.Save()

				translatedDirty := dr.Dirty
				translatedDirty.Min.X -= child.Min.X
				translatedDirty.Min.Y -= child.Min.Y

				// TODO: clip to child.BoundsInParent()?
				dr.GC.Translate(child.Min.X, child.Min.Y)
				cdr := DrawRequest{
					GC: dr.GC,
					Dirty: translatedDirty,
					Done: make(chan bool),
				}
				child.Draw <- cdr
				<- cdr.Done

				dr.GC.Restore()
			}
			dr.Done<- true
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
				GC: gc,
				Dirty: dirtyBounds,
				Done: make(chan bool),
			}
			wf.Draw <-dr
			<-dr.Done

			wf.W.FlushImage()
		}
	}
}
