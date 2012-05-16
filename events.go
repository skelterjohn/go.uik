/*
   Copyright 2012 the go.uik authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package uik

import (
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.wde"
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

type MouseEnteredEvent struct {
	Event
	wde.MouseEnteredEvent
	MouseLocator
	From geom.Coord
}

type MouseExitedEvent struct {
	Event
	wde.MouseExitedEvent
	MouseLocator
	From geom.Coord
}

type CloseEvent struct {
	Event
	wde.CloseEvent
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
	Bounds []geom.Rect
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

type ResizeEvent struct {
	Size geom.Coord
}
