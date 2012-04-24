package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"image"
	"image/color"
)

type LabelData struct {
	Text     string
	FontSize float64
	Color    color.Color
}

type Label struct {
	uik.Block

	data      LabelData
	SetConfig chan<- LabelData
	setConfig chan LabelData
	GetConfig <-chan LabelData
	getConfig chan LabelData

	tbuf image.Image
}

func NewLabel(size geom.Coord, data LabelData) (l *Label) {
	l = new(Label)
	l.Initialize()
	// uik.Report(l.ID, "is label")

	l.Size = size
	l.data = data

	l.render()

	go l.handleEvents()

	return
}

func (l *Label) Initialize() {
	l.Block.Initialize()

	l.setConfig = make(chan LabelData)
	l.SetConfig = l.setConfig
	l.getConfig = make(chan LabelData)
	l.GetConfig = l.getConfig

	l.Paint = func(gc draw2d.GraphicContext) {
		l.draw(gc)
	}
}

func (l *Label) render() {
	l.tbuf = uik.RenderString(l.data.Text, uik.DefaultFontData, l.data.FontSize, l.data.Color)
	l.Buffer = nil
	s := geom.Coord{float64(l.tbuf.Bounds().Max.X), float64(l.tbuf.Bounds().Max.Y)}

	// go uik.ShowBuffer("label text render", l.tbuf)

	l.SetSizeHint(uik.SizeHint{
		MinSize:       s,
		PreferredSize: s,
		MaxSize:       s,
	})
}

func (l *Label) draw(gc draw2d.GraphicContext) {
	// gc.Clear()
	gc.SetFillColor(color.RGBA{A: 1})
	// safeRect(gc, geom.Coord{0, 0}, l.Size)
	gc.Fill()
	tw := float64(l.tbuf.Bounds().Max.X - l.tbuf.Bounds().Min.X)
	th := float64(l.tbuf.Bounds().Max.Y - l.tbuf.Bounds().Min.Y)
	gc.Translate((l.Size.X-tw)/2, (l.Size.Y-th)/2)
	gc.DrawImage(l.tbuf)
}

func (l *Label) handleEvents() {
	for {
		select {
		case e := <-l.Events:
			switch e := e.(type) {
			case uik.ResizeEvent:
				l.Block.HandleEvent(e)
				l.PaintAndComposite()
				// go uik.ShowBuffer("label buffer", l.Buffer)
			default:
				l.HandleEvent(e)
			}
		case data := <-l.setConfig:
			if l.data == data {
				break
			}
			l.data = data
			l.render()
			l.PaintAndComposite()
			// go uik.ShowBuffer("label buffer", l.Buffer)
		case l.getConfig <- l.data:
		case <-l.Redraw:
			l.PaintAndComposite()
			// go uik.ShowBuffer("label buffer", l.Buffer)
		}
	}
}
