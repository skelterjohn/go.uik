/*
   Copyright 2012 the go.uik authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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
	if uik.ReportIDs {
		uik.Report(l.ID, "label")
	}

	// uik.Report(l.ID, "label")

	l.Size = size
	l.data = data

	l.render()

	go l.handleEvents()

	return
}

func (l *Label) Initialize() {
	l.Block.Initialize()

	l.setConfig = make(chan LabelData, 1)
	l.SetConfig = l.setConfig
	l.getConfig = make(chan LabelData, 1)
	l.GetConfig = l.getConfig

	l.Paint = func(gc draw2d.GraphicContext) {
		l.draw(gc)
	}
}

func (l *Label) render() {
	l.tbuf = uik.RenderString(l.data.Text, uik.DefaultFontData, l.data.FontSize, l.data.Color)
	s := geom.Coord{float64(l.tbuf.Bounds().Max.X), float64(l.tbuf.Bounds().Max.Y)}

	// go uik.ShowBuffer("label text render", l.tbuf)

	l.SetSizeHint(uik.SizeHint{
		MinSize:       s,
		PreferredSize: s,
		MaxSize:       s,
	})
}

func (l *Label) draw(gc draw2d.GraphicContext) {
	// uik.Report(l.ID, "Label.draw()")
	//gc.Clear()
	// gc.SetFillColor(color.RGBA{A: 1})
	// safeRect(gc, geom.Coord{0, 0}, l.Size)
	// gc.Fill()
	tw := float64(l.tbuf.Bounds().Max.X - l.tbuf.Bounds().Min.X)
	th := float64(l.tbuf.Bounds().Max.Y - l.tbuf.Bounds().Min.Y)
	gc.Translate((l.Size.X-tw)/2, (l.Size.Y-th)/2)
	gc.DrawImage(l.tbuf)
}

func (l *Label) handleEvents() {
	for {
		select {
		case e := <-l.UserEvents:
			switch e := e.(type) {
			case uik.ResizeEvent:
				if l.Size == e.Size {
					break
				}
				l.Block.HandleEvent(e)
				l.Invalidate()
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
			l.Invalidate()
			// go uik.ShowBuffer("label buffer", l.Buffer)
		case l.getConfig <- l.data:
			// go uik.ShowBuffer("label buffer", l.Buffer)
		}
	}
}
