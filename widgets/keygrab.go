package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/geom"
	"image/color"
	"image"
)

type KeyGrab struct {
	uik.Block
	kbuf image.Image
	key string
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
	l.Buffer = nil
}

func (l *KeyGrab) draw(gc draw2d.GraphicContext) {
	tw := float64(l.kbuf.Bounds().Max.X - l.kbuf.Bounds().Min.X)
	th := float64(l.kbuf.Bounds().Max.Y - l.kbuf.Bounds().Min.Y)
	gc.Translate((l.Size.X-tw)/2, (l.Size.Y-th)/2)
	gc.DrawImage(l.kbuf)
}

func (l *KeyGrab) GrabFocus() {
	
}

func (l *KeyGrab) handleEvents() {
	for {
		select {
		case e := <-l.Events:
			switch e := e.(type) {
			default:
				l.HandleEvent(e)
			}
		case <-l.Redraw:
			l.PaintAndComposite()
		}
	}
}
