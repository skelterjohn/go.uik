package uik

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.wde"
	"image"
)

type MouseEvent interface {
	Where() geom.Coord
}

type MouseLocator struct {
	Loc geom.Coord
}

func (e *MouseLocator) Where() geom.Coord {
	return e.Loc
}

func (e *MouseLocator) Translate(offset geom.Coord) {
	e.Loc.X += offset.X
	e.Loc.Y += offset.Y
}

type MouseMovedEvent struct {
	wde.MouseMovedEvent
	MouseLocator
	From geom.Coord
}

type MouseDraggedEvent struct {
	wde.MouseDraggedEvent
	MouseLocator
	From geom.Coord
}

type MouseDownEvent struct {
	wde.MouseDownEvent
	MouseLocator
}

type MouseUpEvent struct {
	wde.MouseUpEvent
	MouseLocator
}

type CloseEvent struct {
	wde.CloseEvent
}

type ResizeEvent struct {
	wde.ResizeEvent
	Size geom.Coord
}

type RedrawEvent struct {
	Bounds geom.Rect
}

type CompositeRequest struct {
	Buffer image.Image
}

type SizeHint struct {
	MinSize, PreferredSize, MaxSize geom.Coord
}

type PlacementNotification struct {
	Foundation *Foundation
	SizeHints  SizeHintChan
}
