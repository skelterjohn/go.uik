package uik

import (
	"github.com/skelterjohn/go.wde"
	"code.google.com/p/draw2d/draw2d"
	"image/color"
	"fmt"
)

type Button struct {
	Block
	Label string
	pressed bool

	Click chan<- wde.Button
	click chan wde.Button
}

func NewButton(label string) (b *Button) {
	b = new(Button)
	b.MakeChannels()
	b.Label = label

	b.Min = Coord{0, 0}
	b.Size = Coord{100, 50}

	b.click = make(chan wde.Button)
	b.Click = b.click

	go b.handleEvents()
	go b.handleState()

	b.Paint = func(gc draw2d.GraphicContext) {
		b.draw(gc)
	}

	return
}

func (b *Button) draw(gc draw2d.GraphicContext) {
	gc.SetStrokeColor(color.Black)
	if b.pressed {
		gc.SetFillColor(color.RGBA{150, 150, 150, 255})
	} else {
		gc.SetFillColor(color.White)
	}
	draw2d.Rect(gc, 0, 0, b.Size.X, b.Size.Y)
	gc.FillStroke()
	gc.SetFontSize(12)
	gc.FillString(b.Label)
	gc.FillStroke()
	gc.Stroke()
}

func (b *Button) handleState() {
	for {
		select {
		case which := <-b.click:
			fmt.Println("clicked", which)
		}
	}
}

func (b *Button) handleEvents() {
	b.ListenedChannels[b.MouseDownEvents] = true
	b.ListenedChannels[b.MouseUpEvents] = true
	for {
		select {
		case <-b.MouseDownEvents:
			b.pressed = true
			b.Parent.Redraw <- b.BoundsInParent()
		case e := <-b.MouseUpEvents:
			b.pressed = false
			b.Click <- e.Which
			b.Parent.Redraw <- b.BoundsInParent()
		case dr := <-b.Draw:
			b.doPaint(dr.GC)
			dr.Done<- true
		}
	}
}