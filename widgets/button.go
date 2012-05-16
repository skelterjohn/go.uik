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
	"github.com/skelterjohn/go.wde"
	"image/color"
	"image/draw"
	"math"
)

type Clicker chan wde.Button

type ButtonConfig struct {
	Color color.Color
}

type Button struct {
	uik.Foundation
	Label   *Label
	pressed bool

	config ButtonConfig

	setConfig chan ButtonConfig
	SetConfig chan<- ButtonConfig
	getConfig chan ButtonConfig
	GetConfig <-chan ButtonConfig

	Clickers      map[Clicker]bool
	AddClicker    chan Clicker
	RemoveClicker chan Clicker
}

func NewButton(label string) (b *Button) {
	b = new(Button)
	b.Initialize()
	if uik.ReportIDs {
		uik.Report(b.ID, "button")
	}

	b.Size = geom.Coord{70, 30}

	// uik.Report(b.ID, "is button")

	b.Label.SetConfig <- LabelData{
		Text:     label,
		FontSize: 12,
		Color:    color.Black,
	}

	go b.handleEvents()

	return
}

func (b *Button) Initialize() {
	b.Foundation.Initialize()

	b.DrawOp = draw.Over

	b.Label = NewLabel(b.Size, LabelData{
		Text:     "",
		FontSize: 12,
		Color:    color.Black,
	})
	// uik.Report(b.ID, "has label", b.Label.ID)
	lbounds := b.Bounds()
	lbounds.Min.X += 1
	lbounds.Min.Y += 1
	lbounds.Max.X -= 1
	lbounds.Max.Y -= 1
	b.PlaceBlock(&b.Label.Block, lbounds)

	b.Clickers = map[Clicker]bool{}
	b.AddClicker = make(chan Clicker, 1)
	b.RemoveClicker = make(chan Clicker, 1)

	b.Paint = func(gc draw2d.GraphicContext) {
		b.draw(gc)
	}

	var cs geom.Coord
	cs.X, cs.Y = b.Bounds().Size()
	sh := uik.SizeHint{
		MinSize:       cs,
		PreferredSize: cs,
		MaxSize:       geom.Coord{math.Inf(1), math.Inf(1)},
	}
	b.SetSizeHint(sh)

	b.setConfig = make(chan ButtonConfig, 1)
	b.SetConfig = b.setConfig
	b.getConfig = make(chan ButtonConfig, 1)
	b.GetConfig = b.getConfig
}

func safeRect(path draw2d.GraphicContext, min, max geom.Coord) {
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
	gc.Clear()

	bbounds := geom.Rect{
		Min: geom.Coord{0, 0},
		Max: geom.Coord{b.Size.X, b.Size.Y},
	}

	// gc.SetStrokeColor(color.Black)
	if b.pressed {
		gc.SetFillColor(color.RGBA{150, 150, 150, 255})
		safeRect(gc, bbounds.Min, bbounds.Max)
		gc.Fill()
	} else {
		if b.config.Color != nil {
			gc.SetFillColor(b.config.Color)
		} else {
			gc.SetFillColor(color.RGBA{200, 200, 200, 255})
		}
		safeRect(gc, bbounds.Min, bbounds.Max)
		gc.Fill()
	}
}

func (b *Button) handleEvents() {

	for {
		select {
		case e := <-b.UserEvents:
			switch e := e.(type) {
			case uik.MouseEnteredEvent:

			case uik.MouseExitedEvent:

			case uik.MouseDownEvent:
				b.pressed = true
				b.Invalidate()
			case uik.MouseUpEvent:
				b.pressed = false
				for c := range b.Clickers {
					select {
					case c <- e.Which:
					default:
					}
				}
				b.Invalidate()
				// go uik.ShowBuffer("button buffer", b.Buffer)
			case uik.ResizeEvent:
				if b.Size != e.Size {
					b.Foundation.HandleEvent(e)
				}
				lbounds := b.Bounds()
				lbounds.Min.X += 1
				lbounds.Min.Y += 1
				lbounds.Max.X -= 1
				lbounds.Max.Y -= 1
				b.ChildrenBounds[&b.Label.Block] = lbounds
				b.Label.UserEventsIn <- uik.ResizeEvent{
					Size: b.Size,
				}
			default:
				b.Foundation.HandleEvent(e)
			}
		case <-b.BlockInvalidations:
			b.Invalidate()
		case c := <-b.AddClicker:
			b.Clickers[c] = true
		case c := <-b.RemoveClicker:
			if b.Clickers[c] {
				delete(b.Clickers, c)
			}
		case bsh := <-b.BlockSizeHints:
			sh := bsh.SizeHint
			sh.PreferredSize.X += 10
			sh.PreferredSize.Y += 10
			sh.PreferredSize.X = math.Max(sh.PreferredSize.X, b.Size.X)
			sh.PreferredSize.Y = math.Max(sh.PreferredSize.Y, b.Size.Y)
			sh.MaxSize.X = math.Inf(1)
			sh.MaxSize.Y = math.Inf(1)
			b.SetSizeHint(sh)
		case b.config = <-b.setConfig:
			b.Invalidate()
		case b.getConfig <- b.config:
		}
	}
}
