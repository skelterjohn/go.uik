package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"image/color"
)

type Button struct {
	Block
	Label string
	pressed bool
}

func NewButton(label string) (b *Button) {
	b = new(Button)
	b.MakeChannels()
	b.Label = label

	b.Min = Coord{0, 0}
	b.Size = Coord{100, 30}

	go b.handleEvents()

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
}

func (b *Button) handleEvents() {
	b.ListenedChannels[b.MouseDownEvents] = true
	b.ListenedChannels[b.MouseUpEvents] = true
	for {
		select {
		case <-b.MouseDownEvents:
			b.pressed = true
			b.Parent.Redraw <- b.BoundsInParent()
		case <-b.MouseUpEvents:
			b.pressed = false
			b.Parent.Redraw <- b.BoundsInParent()
		case dr := <-b.Draw:
			b.doPaint(dr.GC)
			dr.Done<- true
		}
	}
}