package uik

import (
	"github.com/skelterjohn/go.wde"
)

var NewWindow func (x, y, width, height int) (window wde.Window, err error)

