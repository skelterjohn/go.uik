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

type KeyGrab struct {
	uik.Block
	kbuf image.Image
	key  string
}

func NewKeyGrab(size geom.Coord) (l *KeyGrab) {
	l = new(KeyGrab)
	l.Initialize()
	if uik.ReportIDs {
		uik.Report(l.ID, "keygrab")
	}

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
	l.Parent.UserEventsIn <- uik.KeyFocusRequest{
		Block: &l.Block,
	}
}

func (l *KeyGrab) handleEvents() {
	for {
		select {
		case e := <-l.UserEvents:
			switch e := e.(type) {
			case uik.MouseDownEvent:
				l.GrabFocus()
			case uik.KeyTypedEvent:
				l.key = e.Glyph
				l.render()
				l.Invalidate()
			case uik.KeyFocusEvent:
				l.HandleEvent(e)
				l.Invalidate()
			default:
				l.HandleEvent(e)
			}
		case e := <-l.ResizeEvents:
			l.DoResizeEvent(e)
		}
	}
}
