package uik

import (
	"image"
	"image/draw"
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/go.wde"
)

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
			// TODO: don't do this every time - give a window for all expected buffers to 
			//       come in before flushing prematurely
			wf.W.FlushImage()
		}
	}
}
