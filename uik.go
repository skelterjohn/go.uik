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

// The Block type is a basic unit that can receive events and draw itself.
type Block struct {
	Parent *Foundation

	CloseEvents     chan wde.CloseEvent
	MouseDownEvents chan MouseDownEvent
	MouseUpEvents   chan MouseUpEvent
	Draw            chan draw2d.GraphicContext
	Redraw          chan Bounds

	Paint func(gc draw2d.GraphicContext)

	// minimum point relative to the block's parent: only the parent should set this
	Min Coord
	// size of block
	Size Coord
}

func (b *Block) BoundsInParent() (bounds Bounds) {
	bounds.Min = b.Min
	bounds.Max = b.Min
	bounds.Max.X += b.Size.X
	bounds.Max.Y += b.Size.Y
	return
}

func (b *Block) MakeChannels() {
	b.CloseEvents = make(chan wde.CloseEvent)
	b.MouseDownEvents = make(chan MouseDownEvent)
	b.MouseUpEvents = make(chan MouseUpEvent)
	b.Draw = make(chan draw2d.GraphicContext)
	b.Redraw = make(chan Bounds)
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

func (f *Foundation) handleDrawing() {
	for {
		select {
		case gc := <-f.Draw:
			if f.Paint != nil {
				f.Paint(gc)
			}
			for _, child := range f.Children {
				gc.Save()

				// TODO: clip to child.BoundsInParent()?

				gc.Translate(child.Min.X, child.Min.Y)
				child.Draw <- gc

				gc.Restore()
			}
		case dirtyBounds := <-f.Redraw:
			dirtyBounds.Min.X -= f.Min.X
			dirtyBounds.Min.Y -= f.Min.Y
			f.Parent.Redraw <- dirtyBounds
		}
	}
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

func (f *Foundation) handleEvents() {
	for {
		select {
		case e := <-f.CloseEvents:
			for _, b := range f.Children {
				b.CloseEvents <- e
			}
		case e := <-f.MouseDownEvents:
			b := f.BlockForCoord(e.Loc)
			if b == nil {
				break
			}
			e.X -= int(b.Min.X)
			e.Y -= int(b.Min.Y)
			b.MouseDownEvents <- e
		case e := <-f.MouseUpEvents:
			b := f.BlockForCoord(e.Loc)
			if b == nil {
				break
			}
			e.X -= int(b.Min.X)
			e.Y -= int(b.Min.Y)
			b.MouseUpEvents <- e
		}
	}
}

// A foundation that wraps a wde.Window
type WindowFoundation struct {
	Foundation
	W wde.Window
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
	for e := range wf.W.EventChan() {
		switch e := e.(type) {
		case wde.CloseEvent:
			wf.CloseEvents <- e
			wf.W.Close()
		case wde.MouseDownEvent:
			wf.MouseDownEvents <- MouseDownEvent{
				MouseDownEvent: e,
				MouseLocator: MouseLocator {
					Loc: Coord{float64(e.X), float64(e.Y)},
				},
			}
		case wde.MouseUpEvent:
			wf.MouseUpEvents <- MouseUpEvent{
				MouseUpEvent: e,
				MouseLocator: MouseLocator {
					Loc: Coord{float64(e.X), float64(e.Y)},
				},
			}
		}
	}
}

func (wf *WindowFoundation) handleWindowDrawing() {


	for {
		select {
		case dirtyBounds := <-wf.Redraw:
			gc := draw2d.NewGraphicContext(wf.W.Screen())
			gc.Clear()
			gc.BeginPath()
			// TODO: pass dirtyBounds too, to avoid redrawing out of reach components
			_ = dirtyBounds
			if wf.Paint != nil {
				wf.Paint(gc)
			}
			for _, child := range wf.Children {
				gc.Save()

				// TODO: clip to child.BoundsInParent()?

				gc.Translate(child.Min.X, child.Min.Y)
				child.Draw <- gc

				gc.Restore()
			}

			wf.W.FlushImage()
		}
	}
}
