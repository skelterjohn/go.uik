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
	
	bsize := geom.Coord{100, 50}
	
	b := widgets.NewButton(bsize, "Hi")
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
	
	b2 := widgets.NewButton(bsize, "there")
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
	<-w.Done
}
