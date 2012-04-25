package uik

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.wde"
	"image"
	"time"
)

type Event struct {
	When time.Duration
}

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
	Event
	wde.MouseMovedEvent
	MouseLocator
	From geom.Coord
}

type MouseDraggedEvent struct {
	Event
	wde.MouseDraggedEvent
	MouseLocator
	From geom.Coord
}

type MouseDownEvent struct {
	Event
	wde.MouseDownEvent
	MouseLocator
}

type MouseUpEvent struct {
	Event
	wde.MouseUpEvent
	MouseLocator
}

type CloseEvent struct {
	Event
	wde.CloseEvent
}

type ResizeEvent struct {
	Event
	wde.ResizeEvent
	Size geom.Coord
}

type KeyDownEvent struct {
	Event
	wde.KeyDownEvent
}

type KeyUpEvent struct {
	Event
	wde.KeyUpEvent
}

type KeyTypedEvent struct {
	Event
	wde.KeyTypedEvent
}

type Invalidation struct {
	Bounds geom.Rect
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

type placementNotification struct {
	Foundation *Foundation
	SizeHints  SizeHintChan
}

type KeyFocusRequest struct {
	Block *Block
}

type KeyFocusEvent struct {
	Focus bool
}
