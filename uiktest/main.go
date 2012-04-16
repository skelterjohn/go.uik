package main

import (
	// "fmt"
	"github.com/skelterjohn/go.uik"
)

func main() {
	f, err := uik.NewFoundation(uik.Size{200, 200})
	if err != nil {
		return
	}
	f.Main = uik.NewButton("hi")
	f.Window.Show()
	f.Draw()
	select{}
}
