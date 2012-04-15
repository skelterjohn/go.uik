package main

import (
	// "fmt"
"math"
	"image"
	"github.com/skelterjohn/go.wde"
	"github.com/skelterjohn/go.wde/xgb"
	"code.google.com/p/draw2d/draw2d"
)

var gc *draw2d.ImageGraphicContext
var x, y float64

func render(w wde.Window) {
	gc.Clear()

	for i := 0.0; i < 360; i = i + 36 { // Go from 0 to 360 degrees in 10 degree steps
		gc.BeginPath() // Start a new path
		gc.Save()      // Keep rotations temporary
		gc.Translate(150, 150)
		gc.Rotate(i * (math.Pi / 180.0)) // Rotate by degrees on stack from 'for'
		draw2d.Rect(gc, -50, -50, 50, 50)
		// gc.RLineTo(100, 0)
		gc.Stroke()
		gc.Restore() // Get back the unrotated state
	}

	w.FlushImage()
}

func main() {
	w, err := xgb.NewWindow(300, 300)
	if err != nil {
		panic(err)
	}

	gc = draw2d.NewGraphicContext(w.Screen())
	gc.SetStrokeColor(image.Black)
	gc.SetFillColor(image.White)
	

	w.Show()
	ech := w.EventChan()
	for event := range ech {
		switch event := event.(type) {
		case wde.MouseDownEvent:
			render(w)
			x, y = float64(event.X), float64(event.Y)
		case wde.CloseEvent:
			w.Close()
			return
		}
	}
	select{}
}