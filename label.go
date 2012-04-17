package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"image/color"
)

type Label struct {
	Block

	Text string
	TextCh chan string

	FontSize float64
	FontSizeCh chan float64
}

func NewLabel(size Coord, text string) (l *Label) {
	l = new(Label)
	l.MakeChannels()

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
	
	font := draw2d.GetFont(gc.GetFontData())
	bounds := font.Bounds()
	height := float64(bounds.YMax - bounds.YMin)*l.FontSize/float64(font.UnitsPerEm())
	
	offset := l.Size.Y - (l.Size.Y - height) / 2

	safeRect(gc, Coord{0, 0}, l.Size)
	gc.FillStroke()
	gc.Translate(10, offset)
	gc.SetFontSize(l.FontSize)
	gc.FillString(l.Text)
}

func (l *Label) handleEvents() {
	for {
		select {
		case l.Text = <-l.TextCh:
			if l.Parent != nil {
				l.Parent.Redraw <- l.BoundsInParent()
			}
		case l.FontSize = <-l.FontSizeCh:
			if l.Parent != nil {
				l.Parent.Redraw <- l.BoundsInParent()
			}
		case dr := <-l.Draw:
			l.doPaint(dr.GC)
			dr.Done<- true
		}
	}
}