package main

import (
	"bytes"
	"fmt"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.uik/layouts"
	"github.com/skelterjohn/go.uik/widgets"
	"github.com/skelterjohn/go.wde"
	"image/color"
	"image/gif"
)

func main() {
	go uiktest()
	wde.Run()
}

func uiktest() {

	wbounds := geom.Rect{
		Max: geom.Coord{480, 320},
	}
	w, err := uik.NewWindow(nil, int(wbounds.Max.X), int(wbounds.Max.Y))
	if err != nil {
		fmt.Println(err)
		return
	}
	w.W.SetTitle("uiktest")

	// Create a button with the given size and label
	b := widgets.NewButton("Hi")
	// Here we get the button's label's data
	ld := b.Label.GetConfig()
	// we modify the copy for a special message to display
	ld.Text = "clicked!"

	l := widgets.NewLabel(geom.Coord{100, 50}, widgets.LabelConfig{"text", 14, color.Black})
	b2 := widgets.NewButton("there")
	ld2 := b2.Label.GetConfig()
	ld2.Text = "BAM"

	// the widget.Buttton has a special channels that sends out wde.Buttons
	// whenever its clicked. Here we set up something that changes the
	// label's text every time a click is received.
	clicker := make(widgets.Clicker)
	b.AddClicker <- clicker
	go func() {
		for _ = range clicker {
			b.Label.SetConfig(ld)
			l.SetConfig(widgets.LabelConfig{"ohnoes", 20, color.Black})
		}
	}()

	clicker2 := make(widgets.Clicker)
	b2.AddClicker <- clicker2
	go func() {
		for _ = range clicker2 {
			b.Label.SetConfig(ld2)
			b2.Label.SetConfig(ld)
			l.SetConfig(widgets.LabelConfig{"oops", 14, color.Black})
		}
	}()

	cb := widgets.NewCheckbox(geom.Coord{50, 50})

	kg := widgets.NewKeyGrab(geom.Coord{50, 50})

	ge := layouts.NewGridEngine(layouts.GridConfig{})
	g := layouts.NewLayouter(ge)

	l0_0 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"0, 0", 12, color.Black})
	l0_1 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"0, 1", 12, color.Black})
	l1_0 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"1, 0", 12, color.Black})
	l1_1 := widgets.NewLabel(geom.Coord{}, widgets.LabelConfig{"1, 1", 12, color.Black})

	ge.Add(&l0_0.Block, layouts.GridComponent{
		GridX: 0, GridY: 0,
	})
	ge.Add(&l0_1.Block, layouts.GridComponent{
		GridX: 0, GridY: 1,
	})
	ge.Add(&l1_0.Block, layouts.GridComponent{
		GridX: 1, GridY: 0,
		MinSize: geom.Coord{60, 60},
		AnchorX: layouts.AnchorMax,
		AnchorY: layouts.AnchorMax,
	})
	ge.Add(&l1_1.Block, layouts.GridComponent{
		GridX: 1, GridY: 1,
		AnchorX: layouts.AnchorMin,
		AnchorY: layouts.AnchorMin,
	})
	ge.Add(&widgets.NewButton("Spanner").Block, layouts.GridComponent{
		GridX: 0, GridY: 2,
		ExtraX:  1,
		AnchorX: layouts.AnchorMin | layouts.AnchorMax,
		AnchorY: layouts.AnchorMin | layouts.AnchorMax,
	})
	ge.Add(&widgets.NewButton("Offset Spanner").Block, layouts.GridComponent{
		GridX: 1, GridY: 3,
		ExtraX:  1,
		AnchorX: layouts.AnchorMin | layouts.AnchorMax,
		AnchorY: layouts.AnchorMin | layouts.AnchorMax,
	})

	if true {
		imgReader := bytes.NewReader(gordon_gif())
		gordonImage, gerr := gif.Decode(imgReader)
		if gerr == nil {
			im := widgets.NewImage(widgets.ImageConfig{
				Image: gordonImage,
			})
			ge.Add(&im.Block, layouts.GridComponent{
				GridX: 0, GridY: 4,
				ExtraX: 2,
			})
		} else {
			fmt.Println(gerr)
		}
	}

	clicker3 := make(widgets.Clicker)
	b.AddClicker <- clicker3
	go func() {
		for _ = range clicker3 {
			l0_0.SetConfig(widgets.LabelConfig{"Pow", 12, color.Black})
		}
	}()
	clicker4 := make(widgets.Clicker)
	b2.AddClicker <- clicker4
	go func() {
		for _ = range clicker4 {
			l0_0.SetConfig(widgets.LabelConfig{"gotcha", 12, color.Black})
		}
	}()

	e := widgets.NewEntry(geom.Coord{100, 30})

	_ = kg
	_ = cb
	_ = e

	// the HBox is a special type of grid that lines things up horizontally
	hb := layouts.HBox(layouts.GridConfig{},
		&b.Block,
		&l.Block,
		// &kg.Block,
		&b2.Block,
		&cb.Block,
		// &e.Block,
		&g.Block,
	)

	// set this HBox to be the window pane
	w.SetPane(&hb.Block)

	w.Show()

	// Here we set up a subscription on the window's close events.
	done := make(chan interface{}, 1)
	isDone := func(e interface{}) (accept, done bool) {
		_, accept = e.(uik.CloseEvent)
		done = accept
		return
	}
	w.Block.Subscribe <- uik.Subscription{isDone, done}

	// once a close event comes in on the subscription, end the program
	<-done

	w.W.Close()

	wde.Stop()
}
