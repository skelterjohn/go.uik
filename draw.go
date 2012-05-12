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

package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/go.wde"
	"github.com/skelterjohn/go.wde/xgb"
	"image"
	"image/color"
	"image/draw"
)

func ClearPaint(gc draw2d.GraphicContext) {
	gc.Clear()
	gc.SetFillColor(color.RGBA{155, 0, 0, 255})
	gc.Fill()
}

func ZeroRGBA(rgba *image.RGBA) {
	for y := rgba.Rect.Min.Y; y < rgba.Rect.Max.Y; y++ {
		rowStart := (y-rgba.Rect.Min.Y)*rgba.Stride - rgba.Rect.Min.X*4
		rowEnd := rowStart + (rgba.Rect.Max.X-rgba.Rect.Min.X)*4
		row := rgba.Pix[rowStart:rowEnd]
		for i := range row {
			row[i] = 0
		}
	}
}

func ShowBuffer(title string, buffer image.Image) {
	if buffer == nil {
		return
	}
	width, height := int(buffer.Bounds().Max.X), int(buffer.Bounds().Max.Y)
	if width == 0 || height == 0 {
		return
	}
	w, err := xgb.NewWindow(width, height)
	if err != nil {
		return
	}
	w.SetTitle(title)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if (x/10)%2 == (y/10)%2 {
				w.Screen().Set(x, y, color.White)
			} else {
				w.Screen().Set(x, y, color.RGBA{200, 200, 200, 255})
			}
		}
	}
	draw.Draw(w.Screen(), buffer.Bounds(), buffer, image.Point{0, 0}, draw.Over)
	w.FlushImage()
	w.Show()

loop:
	for e := range w.EventChan() {
		switch e.(type) {
		case wde.CloseEvent, wde.MouseDownEvent:
			break loop
		}
	}
	w.Close()
}
