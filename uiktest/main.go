package main

import (
	"fmt"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.uik/widgets"
	"github.com/skelterjohn/go.uik/layouts"
	"github.com/skelterjohn/geom"
)

func main() {
	wbounds := geom.Rect{
		Max: geom.Coord{480, 320},
	}
	w, err := uik.NewWindow(nil, int(wbounds.Max.X), int(wbounds.Max.Y))
	if err != nil {
		fmt.Println(err)
		return
	}
	w.W.SetTitle("GoUI")

	fl := layouts.NewFlow(wbounds.Max)
	w.PlaceBlock(&fl.Block, wbounds)
	
	b := widgets.NewButton(geom.Coord{100, 50}, "Hi")
	ld := <-b.Label.GetConfig
	ld.Text = "clicked!"
	clicker := make(widgets.Clicker)
	b.AddClicker<- clicker
	go func() {
		for _ = range clicker {
			b.Label.SetConfig<- ld
		}
	}()

	fl.PlaceBlock(&b.Block)
	
	b2 := widgets.NewButton(geom.Coord{70, 30}, "there")
	ld2 := <-b2.Label.GetConfig
	ld2.Text = "BAM"
	clicker2 := make(widgets.Clicker)
	b2.AddClicker<- clicker2
	go func() {
		for _ = range clicker2 {
			b.Label.SetConfig<- ld2
		}
	}()

	fl.PlaceBlock(&b2.Block)

	w.Show()

	done := make(chan interface{})
	isDone := func(e interface{}) (accept, done bool) {
		_, accept = e.(uik.CloseEvent)
		done = accept
		return
	}
	w.Block.Subscribe <- uik.EventSubscription{isDone, done}

	<-done
}
