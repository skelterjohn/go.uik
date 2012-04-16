package uik

import (
	"github.com/skelterjohn/go.wde"
)

type MouseEvent interface {
	Where() Coord
}

type MouseLocator struct {
	Loc Coord
}

func (e *MouseLocator) Where() Coord {
	return e.Loc
}

type MouseMovedEvent struct {
	wde.MouseMovedEvent
	MouseLocator
	From Coord
}

type MouseDraggedEvent struct {
	wde.MouseDraggedEvent
	MouseLocator
	From Coord
}

type MouseDownEvent struct {
	wde.MouseDownEvent
	MouseLocator
}

type MouseUpEvent struct {
	wde.MouseUpEvent
	MouseLocator
}