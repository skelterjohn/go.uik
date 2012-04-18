package uik

import (
	"image"
	"image/draw"
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
	wf.Initialize()

	wf.Size = Coord{float64(width), float64(height)}
	wf.Paint = ClearPaint

	go wf.handleWindowEvents()
	go wf.handleWindowDrawing()
	go wf.HandleEvents()

	return
}

func (wf *WindowFoundation) Show() {
	wf.W.Show()
	RedrawEventChan(wf.Redraw).Stack(RedrawEvent{wf.BoundsInParent()})
}

// wraps mouse events with float64 coordinates
func (wf *WindowFoundation) handleWindowEvents() {
	done := make(chan bool)
	wf.Done = done
	for e := range wf.W.EventChan() {
		switch e := e.(type) {
		case wde.CloseEvent:
			wf.CloseEvents <- CloseEvent{
				CloseEvent: e,
			}
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
	wf.Compositor = make(chan CompositeRequest)

	for {
		select {
		case ce := <- wf.Compositor:
			draw.Draw(wf.W.Screen(), ce.Buffer.Bounds(), ce.Buffer, image.Point{0, 0}, draw.Src)
			// TODO: don't do this every time - give a window for all expected buffers to 
			//       come in before flushing prematurely
			wf.W.FlushImage()
		}
	}
}
