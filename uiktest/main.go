package main

import (
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.uik/layouts"
	"github.com/skelterjohn/go.uik/widgets"
)

// button makes a button, configuring its normal and pressed label

func button(bsize geom.Coord, label, plabel string) *widgets.Button {
	b := widgets.NewButton(bsize, label)
	ld := <-b.Label.GetConfig
	ld.Text = plabel
	clicker := make(widgets.Clicker)
	b.AddClicker <- clicker
	go func() {
		for _ = range clicker {
			b.Label.SetConfig <- ld
		}
	}()
	return b
}

// topbox makes the top-level window with specified size and label,
// creating a foundation for others to build on

func topbox(title string, width, height float64) (*uik.WindowFoundation, *layouts.Flow, error) {
	wbounds := geom.Rect{Max: geom.Coord{width, height}}
	w, err := uik.NewWindow(nil, int(wbounds.Max.X), int(wbounds.Max.Y))
	if err != nil {
		return nil, nil, err
	}
	w.W.SetTitle("GoUI")
	f := layouts.NewFlow(wbounds.Max)
	w.PlaceBlock(&f.Block, wbounds)
	return w, f, err
}

// shutdown sets up a subscription on the window's close events.
func shutdown(w *uik.WindowFoundation) {
	done := make(chan interface{})
	isDone := func(e interface{}) (accept, done bool) {
		_, accept = e.(uik.CloseEvent)
		done = accept
		return
	}
	w.Block.Subscribe <- uik.EventSubscription{isDone, done}

	// once a close event comes in on the subscription, end the program
	<-done
}

// create a window with two buttons and a checkbox
func main() {
	w, f, err := topbox("GoUI", 480, 320)
	if err != nil {
		fmt.Println(err)
		return
	}
	bsize := geom.Coord{100, 50}
	b1 := button(bsize, "Hi", "clicked!")
	b2 := button(bsize, "there", "BAM")
	cb := widgets.NewCheckbox(geom.Coord{50, 50})
	f.PlaceBlock(&b1.Block)
	f.PlaceBlock(&b2.Block)
	f.PlaceBlock(&cb.Block)
	w.Show()
	shutdown(w)
}
