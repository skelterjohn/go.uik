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
	"image/color"
)

func init() {
	uik.RegisterPaint("widgets.Button", func(x interface{}) uik.PaintFunc {
		b := x.(*Button)
		return func(gc draw2d.GraphicContext) {
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
	})

	uik.RegisterPaint("widgets.Checkbox", func(x interface{}) uik.PaintFunc {
		c := x.(*Checkbox)
		return func(gc draw2d.GraphicContext) {
			gc.Clear()
			if c.pressed {
				if c.pressHover {
					gc.SetFillColor(color.RGBA{200, 0, 0, 255})
				} else {
					gc.SetFillColor(color.RGBA{155, 0, 0, 255})
				}
			} else {
				gc.SetFillColor(color.RGBA{255, 0, 0, 255})
			}

			// Draw background rect
			x, y := gc.LastPoint()
			gc.MoveTo(0, 0)
			gc.LineTo(c.Size.X, 0)
			gc.LineTo(c.Size.X, c.Size.Y)
			gc.LineTo(0, c.Size.Y)
			gc.Close()
			gc.Fill()

			// Draw inner rect
			if c.state {
				gc.SetFillColor(color.Black)
				gc.MoveTo(5, 5)
				gc.LineTo(c.Size.X-5, 5)
				gc.LineTo(c.Size.X-5, c.Size.Y-5)
				gc.LineTo(5, c.Size.Y-5)
				gc.Close()
				gc.Fill()
			}

			gc.MoveTo(x, y)
		}
	})
}
