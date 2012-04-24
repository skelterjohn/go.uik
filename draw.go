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

func copyRGBA(dst *image.RGBA, dstOffset image.Point, src *image.RGBA) {
	colMax := src.Rect.Max.X
	if colMax+dstOffset.X > dst.Rect.Max.X {
		colMax = dst.Rect.Max.X - dstOffset.X
	}
	rowMax := src.Rect.Max.Y
	if rowMax+dstOffset.Y > dst.Rect.Max.Y {
		rowMax = dst.Rect.Max.Y - dstOffset.Y
	}

	for row := 0; row < rowMax; row++ {
		srcBegin := row * src.Stride
		srcEnd := srcBegin + 4*colMax

		srcPix := src.Pix[srcBegin:srcEnd]
		dstBegin := (row+dstOffset.Y)*dst.Stride + dstOffset.X*4
		dstEnd := dstBegin + 4*colMax
		dstPix := dst.Pix[dstBegin:dstEnd]
		copy(dstPix, srcPix)
	}
	return
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
