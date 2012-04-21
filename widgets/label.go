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
	GetConfig <-chan LabelData

	tbuf image.Image
}

func NewLabel(size geom.Coord, data LabelData) (l *Label) {
	l = new(Label)
	l.Initialize()

	l.Size = size
	l.data = data

	l.render()

	setConfig := make(chan LabelData)
	l.SetConfig = setConfig
	getConfig := make(chan LabelData)
	l.GetConfig = getConfig

	go l.handleEvents(setConfig, getConfig)

	l.Paint = func(gc draw2d.GraphicContext) {
		l.draw(gc)
	}

	return
}

func (l *Label) render() {
	l.tbuf = uik.RenderString(l.data.Text, uik.DefaultFontData, l.data.FontSize, l.data.Color)
	l.Buffer = nil

	s := geom.Coord{float64(l.tbuf.Bounds().Max.X), float64(l.tbuf.Bounds().Max.Y)}

	l.SetSizeHint(uik.SizeHint{
		MinSize:       s,
		PreferredSize: s,
		MaxSize:       s,
	})
}

func (l *Label) draw(gc draw2d.GraphicContext) {
	tw := float64(l.tbuf.Bounds().Max.X - l.tbuf.Bounds().Min.X)
	th := float64(l.tbuf.Bounds().Max.Y - l.tbuf.Bounds().Min.Y)
	gc.Translate((l.Size.X-tw)/2, (l.Size.Y-th)/2)
	gc.DrawImage(l.tbuf)
}

func (l *Label) handleEvents(setConfig, getConfig chan LabelData) {
	for {
		select {
		case e := <-l.Events:
			switch e := e.(type) {
			default:
				l.HandleEvent(e)
			}
		case l.data = <-setConfig:
			l.render()
		case getConfig <- l.data:
		case <-l.Redraw:
			l.PaintAndComposite()
		}
	}
}
