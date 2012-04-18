package uik

import (
	"image"
	"github.com/skelterjohn/go.wde"
	"math"
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
	Bounds Bounds
}

type CompositeRequest struct {
	Buffer image.Image
}