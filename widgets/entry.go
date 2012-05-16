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

package widgets

import (
	"code.google.com/p/draw2d/draw2d"
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
	"github.com/skelterjohn/go.wde"
	"image"
	"image/color"
	"time"
)

type Entry struct {
	uik.Block
	textBuffer   *image.RGBA
	text         []rune
	runeOffsets  []float64
	cursor       int
	selectCursor int
	selecting    bool
	selected     bool
	textOffset   float64
	fd           draw2d.FontData
	fontSize     float64
}

func NewEntry(size geom.Coord) (e *Entry) {
	e = new(Entry)
	e.Size = size
	e.Initialize()
	if uik.ReportIDs {
		uik.Report(e.ID, "entry")
	}

	e.text = []rune("hello world")
	e.cursor = len(e.text)

	e.render()

	go e.handleEvents()

	e.SetSizeHint(uik.SizeHint{
		MinSize:       e.Size,
		PreferredSize: e.Size,
		MaxSize:       e.Size,
	})

	e.Paint = func(gc draw2d.GraphicContext) {
		e.draw(gc)
	}

	return
}

func (e *Entry) Initialize() {
	e.Block.Initialize()

	e.fd = uik.DefaultFontData
	e.fontSize = 12
}

func (e *Entry) render() {
	const stretchFactor = 1.2

	text := string(e.text)

	height := uik.GetFontHeight(e.fd, e.fontSize) * stretchFactor
	widthMax := float64(len(text)) * e.fontSize

	buf := image.NewRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(widthMax + 1), int(height + 1)},
	})

	gc := draw2d.NewGraphicContext(buf)
	gc.Translate(0, height/stretchFactor)
	gc.SetFontData(e.fd)
	gc.SetFontSize(e.fontSize)
	gc.SetStrokeColor(color.Black)

	var left float64
	e.runeOffsets = []float64{0}
	for _, r := range e.text {
		rt := string(r)
		width := gc.FillString(rt)
		gc.Translate(width, 0)
		left += width
		e.runeOffsets = append(e.runeOffsets, left)
	}

	e.textBuffer = buf.SubImage(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(left + 1), int(height + 1)},
	}).(*image.RGBA)
}

func (e *Entry) GrabFocus() {
	if e.HasKeyFocus {
		return
	}
	e.Parent.UserEventsIn <- uik.KeyFocusRequest{
		Block: &e.Block,
	}
}

func (e *Entry) draw(gc draw2d.GraphicContext) {
	if e.textOffset+e.runeOffsets[e.cursor] < 5 {
		e.textOffset = 5 - e.runeOffsets[e.cursor]
	}
	if e.textOffset+e.runeOffsets[e.cursor] > e.Size.X-5 {
		e.textOffset = e.Size.X - 5 - e.runeOffsets[e.cursor]
	}

	gc.Clear()
	if e.HasKeyFocus {
		gc.SetFillColor(color.RGBA{150, 150, 150, 255})
		safeRect(gc, geom.Coord{0, 0}, e.Size)
		gc.Fill()
	}
	th := float64(e.textBuffer.Bounds().Max.Y - e.textBuffer.Bounds().Min.Y)
	gc.Save()
	gc.Translate(e.textOffset, 0)

	if e.selecting {
		start := e.runeOffsets[e.cursor]
		end := e.runeOffsets[e.selectCursor]
		if start > end {
			start, end = end, start
		}
		gc.SetFillColor(color.RGBA{200, 200, 200, 255})
		safeRect(gc, geom.Coord{start, 0}, geom.Coord{end, e.Size.Y})
		gc.Fill()
	}

	gc.Translate(0, (e.Size.Y-th)/2)

	gc.DrawImage(e.textBuffer)
	gc.Restore()
	if e.HasKeyFocus {
		un := time.Duration(time.Now().UnixNano())
		ms := un / time.Millisecond
		ms = ms % 750
		var intensity uint8 = 255
		if ms > 550 {
			diff := 650 - ms
			if diff < 0 {
				diff *= -1
			}
			intensity = uint8((diff * 255) / 200)
		}
		offset := float64(int(e.runeOffsets[e.cursor] + e.textOffset))
		gc.SetStrokeColor(color.RGBA{A: intensity})
		gc.MoveTo(offset, 0)
		gc.LineTo(offset, e.Size.Y)
		gc.Stroke()
		e.Invalidate()
	}
}

func (e *Entry) cursorForCoord(p geom.Coord) (cursor int) {
	textX := p.X - e.textOffset
	for i, co := range e.runeOffsets {
		// uik.Report(i, co)
		if textX > co {
			cursor = i
		} else {
			break
		}
	}
	return
}

func (e *Entry) getStartEndPositions() (start, end int) {
	start, end = e.cursor-1, e.cursor
	if e.selecting {
		start = e.cursor
		end = e.selectCursor
		if end < start {
			start, end = end, start
		}
	}
	return start, end
}

func (e *Entry) insertRunes(runes []rune) {
	start, end := e.cursor, e.cursor
	if e.selecting {
		end = e.selectCursor
		if end < start {
			start, end = end, start
		}
	}
	head := e.text[:start]
	tail := e.text[end:]
	e.text = make([]rune, len(head)+len(tail)+1)[:0]
	e.text = append(e.text, head...)

	newslice := make([]rune, len(e.text)+len(runes))
	copy(newslice, e.text)
	copy(newslice[len(e.text):], runes)
	e.text = newslice

	e.text = append(e.text, tail...)
	e.cursor = start + len(runes)
	e.selecting = false
}

func (e *Entry) handleEvents() {
	for {
		select {
		case ev := <-e.UserEvents:
			switch ev := ev.(type) {
			case uik.MouseDownEvent:
				e.GrabFocus()
				newcursor := e.cursorForCoord(ev.Loc)
				if e.cursor != newcursor {
					e.cursor = newcursor
					e.Invalidate()
				}
				e.selecting = true
				e.selectCursor = e.cursor
			case uik.MouseUpEvent:
				if e.selecting {
					e.Invalidate()
				}
			case uik.MouseDraggedEvent:
				if !e.selecting {
					break
				}
				newSelectCursor := e.cursorForCoord(ev.Loc)
				if e.cursor != newSelectCursor {
					e.cursor = newSelectCursor
					e.Invalidate()
				}
			case uik.KeyTypedEvent:
				// uik.Report("key", ev.Code, ev.Letter)
				if ev.Chord == wde.CutChord {
					if len(e.text) == 0 || !e.selecting {
						break
					}
					start, end := e.getStartEndPositions()
					wde.SetClipboardText(string(e.text[start:end]))
					copy(e.text[start:], e.text[end:])
					e.text = e.text[:len(e.text)-(end-start)]
					e.cursor = start
					e.selecting = false
				} else if ev.Chord == wde.PasteChord {
					e.insertRunes([]rune(wde.GetClipboardText()))
				} else if ev.Chord == wde.CopyChord {
					if len(e.text) == 0 || !e.selecting {
						break
					}
					start, end := e.getStartEndPositions()
					wde.SetClipboardText(string(e.text[start:end]))
				} else if ev.Glyph != "" {
					e.insertRunes([]rune(ev.Glyph))
				} else {
					switch ev.Key {
					case wde.KeyBackspace:
						if len(e.text) == 0 {
							break
						}
						if !e.selecting && e.cursor == 0 {
							break
						}
						start, end := e.getStartEndPositions()
						copy(e.text[start:], e.text[end:])
						e.text = e.text[:len(e.text)-(end-start)]
						e.cursor = start
						e.selecting = false
					case wde.KeyDelete:
						if len(e.text) == 0 {
							break
						}
						if !e.selecting && e.cursor == len(e.text) {
							break
						}
						start, end := e.getStartEndPositions()
						copy(e.text[start:], e.text[end:])
						e.text = e.text[:len(e.text)-(end-start)]
						e.cursor = start
						e.selecting = false
					case wde.KeyLeftArrow:
						if e.cursor > 0 {
							e.cursor--
						}
						e.selecting = false
					case wde.KeyRightArrow:
						if e.cursor < len(e.text) {
							e.cursor++
						}
						e.selecting = false
					}
				}
				e.render()
				e.Invalidate()
			case uik.KeyFocusEvent:
				e.HandleEvent(ev)
				e.Invalidate()
			default:
				e.HandleEvent(ev)
			}
		}
	}
}
