package uik

import (
	"code.google.com/p/freetype-go/freetype/truetype"
	"code.google.com/p/draw2d/draw2d"
)

var DefaultFontData = draw2d.FontData {
	Name: "luxi",
	Family: draw2d.FontFamilySans,
	Style: draw2d.FontStyleNormal,
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