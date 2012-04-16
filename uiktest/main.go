package main

import (
	// "fmt"
	"github.com/skelterjohn/go.uik"
)

func main() {
	w, err := uik.NewWindow(nil, 200, 200)
	if err != nil {
		panic(err)
	}
	w.W.SetTitle("go.uik")
	b := uik.NewButton("Hi")
	w.AddBlock(&b.Block)
	w.Show()
	<-w.Done
}
