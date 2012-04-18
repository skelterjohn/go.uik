package uik

import (
	"image"
	"image/color"
	"code.google.com/p/freetype-go/freetype/truetype"
	"code.google.com/p/draw2d/draw2d"
)

var DefaultFontData = draw2d.FontData {
	Name: "luxi",
	Family: draw2d.FontFamilySans,
	Style: draw2d.FontStyleNormal,
}

func GetFontHeight(fd draw2d.FontData, size float64) (height float64) {
	font := draw2d.GetFont(fd)
	bounds := font.Bounds()
	height = float64(bounds.YMax - bounds.YMin)*size/float64(font.UnitsPerEm())
	return
}

func RenderString(text string, fd draw2d.FontData, size float64, color color.Color) (buffer image.Image) {
	height := GetFontHeight(fd, size)
	widthMax := float64(len(text)) * size
	
	buf := image.NewRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(widthMax+1), int(height+1)},
	})

	gc := draw2d.NewGraphicContext(buf)
	gc.Translate(0, height)
	gc.SetFontData(fd)
	gc.SetFontSize(size)
	gc.SetStrokeColor(color)
	width := gc.FillString(text)

	buffer = buf.SubImage(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(width+1), int(height+1)},
	})

	return
}

func init() {
	font, err := truetype.Parse(luxisr_ttf())
	if err != nil {
		// TODO: log error
		println("error!")
		println(err.Error())
	}

	draw2d.RegisterFont(DefaultFontData, font)
}