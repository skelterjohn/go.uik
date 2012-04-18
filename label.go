package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"image"
	"image/color"
)

type Label struct {
	Block

	Text string
	TextCh chan string

	FontSize float64
	FontSizeCh chan float64

	tbuf image.Image
}

func NewLabel(size Coord, text string) (l *Label) {
	l = new(Label)
	l.Initialize()

	l.Size = size

	l.TextCh = make(chan string)
	l.FontSizeCh = make(chan float64)

	go l.handleEvents()

	l.TextCh <- text
	l.FontSizeCh <- 12

	l.Paint = func(gc draw2d.GraphicContext) {
		l.draw(gc)
	}

	return
}

func (l *Label) draw(gc draw2d.GraphicContext) {
	gc.SetStrokeColor(color.Black)
	
	// height := GetFontHeight(gc.GetFontData(), l.FontSize)
	
	// offset := l.Size.Y - (l.Size.Y - height) / 2

	safeRect(gc, Coord{0, 0}, l.Size)
	gc.FillStroke()
	tw := float64(l.tbuf.Bounds().Max.X -  l.tbuf.Bounds().Min.X)
	th := float64(l.tbuf.Bounds().Max.Y -  l.tbuf.Bounds().Min.Y)
	gc.Translate((l.Size.X-tw)/2, (l.Size.Y-th)/2)
	gc.DrawImage(l.tbuf)

	// gc.Translate(10, offset)
	// gc.SetFontData(DefaultFontData)
	// gc.SetFontSize(l.FontSize)
	// gc.FillString(l.Text)
}

func (l *Label) handleEvents() {
	for {
		select {
		case l.Text = <-l.TextCh:
			l.tbuf = RenderString(l.Text, DefaultFontData, l.FontSize, color.Black)
			if l.Parent != nil {
				RedrawEventChan(l.Redraw).Stack(RedrawEvent{l.BoundsInParent()})
			}
		case l.FontSize = <-l.FontSizeCh:
			l.tbuf = RenderString(l.Text, DefaultFontData, l.FontSize, color.Black)
			if l.Parent != nil {
				RedrawEventChan(l.Redraw).Stack(RedrawEvent{l.BoundsInParent()})
			}
		case <-l.Redraw:
			bgc := l.PrepareBuffer()
			l.DoPaint(bgc)
			l.Compositor <- CompositeRequest {
				l.Buffer,
			}
		}
	}
}