package main

import (
	"fmt"
	"github.com/skelterjohn/go.uik"
)

func main() {
	width := 480.0
	height := 320.0
	w, err := uik.NewWindow(nil, int(width), int(height))
	if err != nil {
		fmt.Println(err)
		return
	}
	w.W.SetTitle("GoUI")
	
	bsize := uik.Coord{100, 50}
	
	b := uik.NewButton(bsize, "Hi")
	ld := <-b.Label.GetConfig
	ld.Text = "clicked!"
	clicker := make(uik.Clicker)
	b.AddClicker<- clicker
	go func() {
		for _ = range clicker {
			b.Label.SetConfig<- ld
		}
	}()

	w.PlaceBlock(&b.Block, uik.Bounds{
		Min: uik.Coord{50, 150},
		Max: uik.Coord{150, 200},
	})

	w.Show()
	<-w.Done
}
