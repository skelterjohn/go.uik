package uik

import (
	"image"
	"github.com/skelterjohn/go.wde"
	"github.com/skelterjohn/geom"
	"math"
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

type RedrawEventChan chan RedrawEvent
func (ch RedrawEventChan) Stack(e RedrawEvent) {
	for {
		select {
		case ch<- e:
			return
		case ne := <-ch:
			e.Bounds.Min.X = math.Min(e.Bounds.Min.X, ne.Bounds.Min.X)
			e.Bounds.Min.Y = math.Min(e.Bounds.Min.Y, ne.Bounds.Min.Y)
			e.Bounds.Max.X = math.Max(e.Bounds.Max.X, ne.Bounds.Max.X)
			e.Bounds.Max.Y = math.Max(e.Bounds.Max.Y, ne.Bounds.Max.Y)
		}
	}
}

type RedrawEvent struct {
	Bounds geom.Rect
}

type CompositeRequest struct {
	Buffer image.Image
}