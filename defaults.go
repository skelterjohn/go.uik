package uik

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/go.wde"
	"github.com/skelterjohn/go.wde/xgb"
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
