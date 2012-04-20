package uik

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.wde"
	"image"
	"image/draw"
	"time"
)

const FrameDelay = 16 * 1e6

// A foundation that wraps a wde.Window
type WindowFoundation struct {
	Foundation
	W               wde.Window
	waitForRepaint  chan bool
	doRepaintWindow chan bool
}

func NewWindow(parent wde.Window, width, height int) (wf *WindowFoundation, err error) {
	wf = new(WindowFoundation)

	wf.W, err = WindowGenerator(parent, width, height)
	if err != nil {
		return
	}
	wf.Initialize()

	wf.waitForRepaint = make(chan bool)
	wf.doRepaintWindow = make(chan bool)

	wf.Size = geom.Coord{float64(width), float64(height)}
	wf.Paint = ClearPaint

	wf.Compositor = make(chan CompositeRequest)

	go wf.handleWindowEvents()
	go wf.handleWindowDrawing()
	go wf.HandleEvents()

	return
}

func (wf *WindowFoundation) Show() {
	wf.W.Show()
	RedrawEventChan(wf.Redraw).Stack(RedrawEvent{
		geom.Rect{
			geom.Coord{0, 0},
			wf.Size,
		},
	})
}

// wraps mouse events with float64 coordinates
func (wf *WindowFoundation) handleWindowEvents() {
	for e := range wf.W.EventChan() {
		switch e := e.(type) {
		case wde.CloseEvent:
			wf.eventsIn <- CloseEvent{
				CloseEvent: e,
			}
			wf.W.Close()
		case wde.MouseDownEvent:
			wf.eventsIn <- MouseDownEvent{
				MouseDownEvent: e,
				MouseLocator: MouseLocator{
					Loc: geom.Coord{float64(e.Where.X), float64(e.Where.Y)},
				},
			}
		case wde.MouseUpEvent:
			wf.eventsIn <- MouseUpEvent{
				MouseUpEvent: e,
				MouseLocator: MouseLocator{
					Loc: geom.Coord{float64(e.Where.X), float64(e.Where.Y)},
				},
			}
		case wde.ResizeEvent:
			wf.waitForRepaint <- true
			wf.eventsIn <- ResizeEvent{
				ResizeEvent: e,
				Size: geom.Coord{
					X: float64(e.Width),
					Y: float64(e.Height),
				},
			}
			wf.Redraw <- RedrawEvent{
				wf.Bounds(),
			}
			go wf.SleepRepaint(FrameDelay)
		}
	}
}

func (wf *WindowFoundation) SleepRepaint(delay time.Duration) {
	time.Sleep(delay)
	wf.doRepaintWindow <- true
}

func (wf *WindowFoundation) handleWindowDrawing() {

	waitingForRepaint := false
	newStuff := false

	flush := func() {
		wf.W.FlushImage()
		newStuff = false
		waitingForRepaint = true
		go wf.SleepRepaint(FrameDelay)
	}

	for {
		select {
		case ce := <-wf.Compositor:
			draw.Draw(wf.W.Screen(), ce.Buffer.Bounds(), ce.Buffer, image.Point{0, 0}, draw.Src)
			if waitingForRepaint {
				newStuff = true
			} else {
				flush()
			}
		case waitingForRepaint = <-wf.waitForRepaint:
		case <-wf.doRepaintWindow:
			waitingForRepaint = false
			if !newStuff {
				break
			}
			flush()
		}
	}
}
