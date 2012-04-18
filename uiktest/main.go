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
	w.AddBlock(&b.Block)
	w.Show()
	<-w.Done
}
