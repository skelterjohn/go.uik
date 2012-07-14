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
	"image"
	"image/color"
)

func ClearPaint(gc draw2d.GraphicContext) {
	if true {
		gc.Clear()
		gc.SetFillColor(color.RGBA{155, 0, 0, 255})
		gc.Fill()
	}
}

func ZeroRGBA(rgba *image.RGBA) {
	for y := rgba.Rect.Min.Y; y < rgba.Rect.Max.Y; y++ {
		rowStart := rgba.PixOffset(rgba.Rect.Min.X, y)
		rowEnd := rgba.PixOffset(rgba.Rect.Max.X, y)
		row := rgba.Pix[rowStart:rowEnd]
		for i := range row {
			row[i] = 0
		}
	}
}

type PaintFunc func(draw2d.GraphicContext)
type PaintGen func(interface{}) PaintFunc

var paintGens = map[string]PaintGen{
	"window": windowPaintGen,
}

func RegisterPaint(path string, dg PaintGen) {
	paintGens[path] = dg
}

func LookupPaint(path string, x interface{}) (pf PaintFunc) {
	pg, ok := paintGens[path]
	if !ok {
		return
	}
	pf = pg(x)
	return
}

func windowPaintGen(x interface{}) (pf PaintFunc) {
	wf := x.(*WindowFoundation)
	return func(gc draw2d.GraphicContext) {
		gc.SetFillColor(color.White)
		draw2d.Rect(gc, 0, 0, wf.Size.X, wf.Size.Y)
		gc.Fill()
	}
	return
}
