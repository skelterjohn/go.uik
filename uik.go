package uik

import (
	"github.com/skelterjohn/go.wde"
)

var WindowGenerator func(parent wde.Window, width, height int) (window wde.Window, err error)

type DrawRequest struct {
	Dirty Bounds
}

