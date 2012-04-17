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
	
	bw := width/4
	bh := width/4
	borigin := uik.Coord{width/2 - bw/2, height/2 - bh/2}
	bsize   := uik.Coord{bw, bh}
	
	b := uik.NewButton(borigin, bsize, "Hi")
	w.AddBlock(&b.Block)
	w.Show()
	<-w.Done
}
