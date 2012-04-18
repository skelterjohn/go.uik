package uik

import (
	"github.com/skelterjohn/go.wde"
	"code.google.com/p/draw2d/draw2d"
	"image/color"
)

type Button struct {
	Block
	Label *Label
	pressed bool

	Click chan<- wde.Button
	click chan wde.Button
}

func NewButton(size Coord, label string) (b *Button) {
	b = new(Button)
	b.Initialize()
	b.Label = NewLabel(size, label)

	b.Min = Coord{0, 0}
	b.Size = size

	b.click = make(chan wde.Button)
	b.Click = b.click

	go b.handleEvents()
	go b.handleState()

	b.Paint = func(gc draw2d.GraphicContext) {
		b.draw(gc)
	}

	return
}

func safeRect(path draw2d.GraphicContext, min, max Coord) {
	x1, y1 := min.X, min.Y
	x2, y2 := max.X, max.Y
	x, y := path.LastPoint()
    path.MoveTo(x1, y1)
    path.LineTo(x2, y1)
    path.LineTo(x2, y2)
    path.LineTo(x1, y2)
    path.Close()
    path.MoveTo(x, y)
}

func (b *Button) draw(gc draw2d.GraphicContext) {

	gc.SetStrokeColor(color.Black)
	if b.pressed {
		gc.SetFillColor(color.RGBA{150, 150, 150, 255})
	} else {
		gc.SetFillColor(color.White)
	}
	safeRect(gc, Coord{0, 0}, b.Size)
	gc.FillStroke()
	b.Label.DoPaint(gc)
}

func (b *Button) handleState() {
	for {
		select {
		case <-b.click:
			b.Label.TextCh <- "clicked!"
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
			b.Label.TextCh <- "pressing"
			b.PaintAndComposite()
		case e := <-b.MouseUpEvents:
			b.pressed = false
			b.Click <- e.Which
			b.PaintAndComposite()
		case <-b.RedrawOut:
			b.PaintAndComposite()
		}
	}
}