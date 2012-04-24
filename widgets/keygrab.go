package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image"
	"image/color"
)

type KeyGrab struct {
	uik.Block
	kbuf image.Image
	key  string
}

func NewKeyGrab(size geom.Coord) (l *KeyGrab) {
	l = new(KeyGrab)
	l.Initialize()

	l.Size = size
	l.key = "x"
	l.render()

	go l.handleEvents()

	l.SetSizeHint(uik.SizeHint{
		MinSize:       l.Size,
		PreferredSize: l.Size,
		MaxSize:       l.Size,
	})

	l.Paint = func(gc draw2d.GraphicContext) {
		l.draw(gc)
	}

	return
}

func (l *KeyGrab) render() {
	l.kbuf = uik.RenderString(l.key, uik.DefaultFontData, 12, color.Black)
}

func (l *KeyGrab) draw(gc draw2d.GraphicContext) {
	gc.Clear()
	if l.HasKeyFocus {
		gc.SetFillColor(color.RGBA{150, 150, 150, 255})
		safeRect(gc, geom.Coord{0, 0}, l.Size)
		gc.FillStroke()
	}
	tw := float64(l.kbuf.Bounds().Max.X - l.kbuf.Bounds().Min.X)
	th := float64(l.kbuf.Bounds().Max.Y - l.kbuf.Bounds().Min.Y)
	gc.Translate((l.Size.X-tw)/2, (l.Size.Y-th)/2)
	gc.DrawImage(l.kbuf)
}

func (l *KeyGrab) GrabFocus() {
	if l.HasKeyFocus {
		return
	}
	l.Parent.EventsIn <- uik.KeyFocusRequest{
		Block: &l.Block,
	}
}

func (l *KeyGrab) handleEvents() {
	for {
		select {
		case e := <-l.Events:
			switch e := e.(type) {
			case uik.MouseDownEvent:
				l.GrabFocus()
			case uik.KeyTypedEvent:
				l.key = e.Letter
				l.render()
				l.PaintAndComposite()
			case uik.KeyFocusEvent:
				l.HandleEvent(e)
				l.PaintAndComposite()
			default:
				l.HandleEvent(e)
			}
		case <-l.Redraw:
			l.PaintAndComposite()
		}
	}
}
