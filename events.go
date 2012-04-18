package uik

import (
	"image"
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

func (e *MouseLocator) Translate(offset Coord) {
	e.Loc.X += offset.X
	e.Loc.Y += offset.Y
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

type CloseEvent wde.CloseEvent

type RedrawEvent struct {
	Bounds Bounds
}

type CompositeRequest struct {
	Buffer image.Image
}