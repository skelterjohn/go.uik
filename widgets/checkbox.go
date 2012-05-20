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
	"github.com/skelterjohn/geom"
	"github.com/skelterjohn/go.uik"
)

type Checker chan bool

type Checkbox struct {
	uik.Block

	state, pressed, pressHover bool
}

func NewCheckbox(size geom.Coord) (c *Checkbox) {
	c = new(Checkbox)
	c.Initialize()
	c.Paint = uik.LookupPaint("widgets.Checkbox", c)

	if uik.ReportIDs {
		uik.Report(c.ID, "checkbox")
	}
	c.Size = size

	go c.handleEvents()

	c.SetSizeHint(uik.SizeHint{
		MinSize:       size,
		PreferredSize: size,
		MaxSize:       size,
	})

	return
}

func (c *Checkbox) handleEvents() {
	for {
		select {
		case e := <-c.UserEvents:
			switch e := e.(type) {
			case uik.MouseEnteredEvent:
				if c.pressed {
					c.pressHover = true
					c.Invalidate()
				}
			case uik.MouseExitedEvent:
				if c.pressed {
					c.pressHover = false
					c.Invalidate()
				}
			case uik.MouseDownEvent:
				c.pressed = true
				c.pressHover = true
				c.Invalidate()
			case uik.MouseUpEvent:
				if c.pressHover {
					c.state = !c.state
					c.Invalidate()
				}
				c.pressHover = false
				c.pressed = false
			default:
				c.Block.HandleEvent(e)
			}
		}
	}
}
