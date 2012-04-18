package uik

import (
	"github.com/skelterjohn/go.wde/xgb"
	"github.com/skelterjohn/go.wde"
	"code.google.com/p/draw2d/draw2d"
)

func ClearPaint(gc draw2d.GraphicContext) {
	gc.Clear()
}

func init() {
	WindowGenerator = func(parent wde.Window, width, height int) (window wde.Window, err error) {
		window, err = xgb.NewWindow(width, height)
		return
	}
}