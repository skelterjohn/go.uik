package uik

import (
	"github.com/skelterjohn/go.wde/xgb"
	"github.com/skelterjohn/go.wde"
)

func init() {
	WindowGenerator = func(parent wde.Window, width, height int) (window wde.Window, err error) {
		window, err = xgb.NewWindow(width, height)
		return
	}
}